package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/internal/apierrors"
	"github.com/distr-sh/distr/internal/auth"
	"github.com/distr-sh/distr/internal/authjwt"
	"github.com/distr-sh/distr/internal/authkey"
	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/db"
	"github.com/distr-sh/distr/internal/mail"
	"github.com/distr-sh/distr/internal/mailsending"
	"github.com/distr-sh/distr/internal/mailtemplates"
	"github.com/distr-sh/distr/internal/mapping"
	"github.com/distr-sh/distr/internal/middleware"
	"github.com/distr-sh/distr/internal/security"
	"github.com/distr-sh/distr/internal/types"
	"github.com/distr-sh/distr/internal/util"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/httprate"
	"github.com/google/uuid"
	"github.com/oaswrap/spec/adapter/chiopenapi"
	"github.com/oaswrap/spec/option"
	"go.uber.org/zap"
)

func SettingsRouter(r chiopenapi.Router) {
	r.Route("/user", func(r chiopenapi.Router) {
		r.WithOptions(option.GroupTags("Settings"))

		r.Post("/", userSettingsUpdateHandler).
			With(option.Description("Update user settings")).
			With(option.Request(api.UpdateUserAccountRequest{})).
			With(option.Response(http.StatusOK, types.UserAccount{}))

		r.Post("/email", userSettingsUpdateEmailHandler()).
			With(option.Description("Update current user email address")).
			With(option.Request(api.UpdateUserAccountEmailRequest{})).
			With(option.Response(http.StatusAccepted, nil))
	})

	r.Route("/verify", func(r chiopenapi.Router) {
		r.WithOptions(option.GroupHidden(true))

		r.With(requestVerificationMailRateLimitPerUser).
			Post("/request", userSettingsVerifyRequestHandler)

		r.Post("/confirm", userSettingsVerifyConfirmHandler)
	})

	r.Route("/tokens", func(r chiopenapi.Router) {
		r.WithOptions(option.GroupTags("Access Tokens"))

		r.Use(middleware.RequireOrgAndRole)

		r.Get("/", getAccessTokensHandler()).
			With(option.Description("List all access tokens")).
			With(option.Response(http.StatusOK, []api.AccessToken{}))

		r.Post("/", createAccessTokenHandler()).
			With(option.Description("Create a new access token")).
			With(option.Request(api.CreateAccessTokenRequest{})).
			With(option.Response(http.StatusCreated, api.AccessTokenWithKey{}))

		r.Route("/{accessTokenId}", func(r chiopenapi.Router) {
			type AccessTokenIDRequest struct {
				AccessTokenID uuid.UUID `path:"accessTokenId"`
			}

			r.Delete("/", deleteAccessTokenHandler()).
				With(option.Description("Delete an access token")).
				With(option.Request(AccessTokenIDRequest{}))
		})
	})
}

func userSettingsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)
	body, err := JsonBody[api.UpdateUserAccountRequest](w, r)
	if err != nil {
		return
	}

	if err := body.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := auth.CurrentUser()
	isUpdateNeeded := false

	if body.Name != nil && *body.Name != user.Name {
		user.Name = *body.Name
		isUpdateNeeded = true
	}

	if body.Password != nil {
		user.Password = *body.Password
		if err := security.HashPassword(user); err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			log.Error("failed to hash password", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		isUpdateNeeded = true
	}

	if body.ImageID != nil && !util.PtrEq(user.ImageID, body.ImageID) {
		user.ImageID = body.ImageID
		isUpdateNeeded = true
	}

	if user.EmailVerifiedAt == nil && auth.CurrentUserEmailVerified() {
		// because reset tokens can also verify the users email address
		user.EmailVerifiedAt = util.PtrTo(time.Now())
		isUpdateNeeded = true
	}

	if isUpdateNeeded {
		if err := db.UpdateUserAccount(ctx, user); err != nil {
			if errors.Is(err, apierrors.ErrNotFound) {
				http.Error(w, err.Error(), http.StatusBadRequest)
			} else {
				log.Error("failed to update user", zap.Error(err))
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	RespondJSON(w, user)
}

func userSettingsUpdateEmailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		mailer := internalctx.GetMailer(ctx)
		log := internalctx.GetLogger(ctx)
		auth := auth.Authentication.Require(ctx)
		user := auth.CurrentUser()

		body, err := JsonBody[api.UpdateUserAccountEmailRequest](w, r)
		if err != nil {
			return
		}

		if err := body.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if user.Email == body.Email {
			http.Error(w, "new email must be different from current email", http.StatusBadRequest)
			return
		}

		if exists, err := db.ExistsUserAccountWithEmail(ctx, body.Email); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if exists {
			http.Error(w, "email already in use", http.StatusBadRequest)
			return
		}

		// Set new email on the UserAccount to generate a verification token
		// This is not saved to the DB yet!
		oldEmail := user.Email
		user.Email = body.Email
		_, token, err := authjwt.GenerateVerificationTokenValidFor(*user)
		if err != nil {
			log.Error("failed to send email verification", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, "failed to generate verification token", http.StatusInternalServerError)
			return
		}
		user.Email = oldEmail

		msg := mail.New(
			mail.To(body.Email),
			mail.Subject("[Action required] Distr E-Mail address change"),
			mail.HtmlBodyTemplate(mailtemplates.UpdateEmail(*user, *auth.CurrentOrg(), token)),
		)

		if err := mailer.Send(ctx, msg); err != nil {
			log.Error("failed to send email verification", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func userSettingsVerifyRequestHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)
	userAccount := auth.CurrentUser()
	if userAccount.EmailVerifiedAt != nil {
		w.WriteHeader(http.StatusNoContent)
	} else if err := mailsending.SendUserVerificationMail(ctx, *userAccount, *auth.CurrentOrg()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sentry.GetHubFromContext(ctx).CaptureException(err)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func userSettingsVerifyConfirmHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)
	userAccount := auth.CurrentUser()
	if !auth.CurrentUserEmailVerified() {
		http.Error(w, "token does not have verified claim", http.StatusForbidden)
		return
	}

	if userAccount.Email != auth.CurrentUserEmail() {
		userAccount.Email = auth.CurrentUserEmail()
	} else if userAccount.EmailVerifiedAt != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := db.UpdateUserAccountEmailVerified(ctx, userAccount); err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			http.Error(w, "could not update user", http.StatusBadRequest)
		} else {
			log.Error("could not update user", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, "could not update user", http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func getAccessTokensHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		auth := auth.Authentication.Require(ctx)
		tokens, err := db.GetAccessTokens(ctx, auth.CurrentUserID(), *auth.CurrentOrgID())
		if err != nil {
			log.Warn("error getting tokens", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			RespondJSON(w, mapping.List(tokens, mapping.AccessTokenToDTO))
		}
	}
}

func createAccessTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		auth := auth.Authentication.Require(ctx)
		request, err := JsonBody[api.CreateAccessTokenRequest](w, r)
		if err != nil {
			return
		}

		key, err := authkey.NewKey()
		if err != nil {
			log.Warn("error creating token", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		token := types.AccessToken{
			ExpiresAt:      request.ExpiresAt,
			Label:          request.Label,
			UserAccountID:  auth.CurrentUserID(),
			Key:            key,
			OrganizationID: *auth.CurrentOrgID(),
		}
		if err := db.CreateAccessToken(ctx, &token); err != nil {
			log.Warn("error creating token", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			RespondJSON(w, mapping.AccessTokenToDTO(token).WithKey(token.Key))
		}
	}
}

func deleteAccessTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		tokenID, err := uuid.Parse(r.PathValue("accessTokenId"))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		auth := auth.Authentication.Require(ctx)
		if err := db.DeleteAccessToken(ctx, tokenID, auth.CurrentUserID()); err != nil {
			log.Warn("error deleting token", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

var requestVerificationMailRateLimitPerUser = httprate.Limit(
	3,
	10*time.Minute,
	httprate.WithKeyFuncs(middleware.RateLimitUserIDKey),
)

var inviteUserRateLimiter = httprate.Limit(
	3,
	10*time.Minute,
	httprate.WithKeyFuncs(middleware.RateLimitUserIDKey, middleware.RateLimitPathValueKey("userId")),
)

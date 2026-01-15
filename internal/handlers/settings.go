package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/internal/apierrors"
	"github.com/distr-sh/distr/internal/auth"
	"github.com/distr-sh/distr/internal/authkey"
	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/db"
	"github.com/distr-sh/distr/internal/mailsending"
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
	r.Post("/user", userSettingsUpdateHandler).
		With(option.Tags("Settings")).
		With(option.Description("Update user settings")).
		With(option.Request(api.UpdateUserAccountRequest{})).
		With(option.Response(http.StatusOK, types.UserAccount{}))
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

	if body.Name != "" {
		user.Name = body.Name
	}
	if body.Password != nil {
		user.Password = *body.Password
		if err := security.HashPassword(user); err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			log.Error("failed to hash password", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	if user.EmailVerifiedAt == nil && auth.CurrentUserEmailVerified() {
		// because reset tokens can also verify the users email address
		user.EmailVerifiedAt = util.PtrTo(time.Now())
	}

	if err := db.UpdateUserAccount(ctx, user); errors.Is(err, apierrors.ErrNotFound) {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if err != nil {
		log.Error("failed to update user", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		RespondJSON(w, user)
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
	} else if userAccount.EmailVerifiedAt != nil {
		w.WriteHeader(http.StatusNoContent)
	} else if err := db.UpdateUserAccountEmailVerified(ctx, userAccount); err != nil {
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

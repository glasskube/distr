package handlers

import (
	"errors"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/cloud/internal/auth"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/mailsending"
	"github.com/glasskube/cloud/internal/mapping"
	"github.com/glasskube/cloud/internal/middleware"
	"github.com/glasskube/cloud/internal/types"
	"go.uber.org/zap"

	"github.com/glasskube/cloud/api"
	"github.com/glasskube/cloud/internal/apierrors"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/security"
	"github.com/go-chi/chi/v5"
)

func SettingsRouter(r chi.Router) {
	r.Post("/user", userSettingsUpdateHandler)
	r.Route("/verify", func(r chi.Router) {
		r.With(middleware.RateLimitPerUser).Post("/request", userSettingsVerifyRequestHandler)
		r.Post("/confirm", userSettingsVerifyConfirmHandler)
	})
	r.Route("/tokens", func(r chi.Router) {
		r.Get("/", getAccessTokensHandler())
		r.Post("/", createAccessTokenHandler())
		r.Route("/{id}", func(r chi.Router) {
			r.Delete("/", deleteAccessTokenHandler())
		})
	})
}

func userSettingsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	body, err := JsonBody[api.UpdateUserAccountRequest](w, r)
	if err != nil {
		return
	}

	if err := body.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := db.GetCurrentUser(ctx)
	if err != nil {
		log.Error("failed to get current user", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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
	if userAccount, err := db.GetCurrentUser(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if userAccount.EmailVerifiedAt != nil {
		w.WriteHeader(http.StatusNoContent)
	} else if err := mailsending.SendUserVerificationMail(ctx, *userAccount); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sentry.GetHubFromContext(ctx).CaptureException(err)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func userSettingsVerifyConfirmHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	if userAccount, err := db.GetCurrentUser(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if verifiedInToken, err := auth.CurrentUserEmailVerified(ctx); err != nil {
		log.Warn("could not check token has verified claim", zap.Error(err))
		http.Error(w, "could not check token has verified claim", http.StatusBadRequest)
	} else if !verifiedInToken {
		http.Error(w, "token does not have verified claim", http.StatusForbidden)
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
		userID, err := auth.CurrentUserId(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		tokens, err := db.GetAccessTokensByUserAccountID(ctx, userID)
		if err != nil {
			log.Warn("error getting tokens", zap.Error(err))
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
		userID, err := auth.CurrentUserId(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		request, err := JsonBody[api.CreateAccessTokenRequest](w, r)
		if err != nil {
			return
		}
		token := types.AccessToken{
			ExpiresAt:     request.ExpiresAt,
			Label:         request.Label,
			UserAccountID: userID,
		}
		if err := db.CreateAccessToken(ctx, &token); err != nil {
			log.Warn("error creating token", zap.Error(err))
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
		tokenID := r.PathValue("id")
		if userID, err := auth.CurrentUserId(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if err := db.DeleteAccessToken(ctx, tokenID, userID); err != nil {
			log.Warn("error deleting token", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

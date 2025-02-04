package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/mailsending"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func UserAccountsRouter(r chi.Router) {
	r.Use(middleware.RequireOrgID, middleware.RequireUserRole)
	r.With(requireUserRoleVendor).Group(func(r chi.Router) {
		r.Get("/", getUserAccountsHandler)
		r.Post("/", createUserAccountHandler)
		r.Route("/{userId}", func(r chi.Router) {
			r.Use(userAccountMiddleware)
			r.Delete("/", deleteUserAccountHandler)
		})
	})
}

func getUserAccountsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)
	if userAccoutns, err := db.GetUserAccountsByOrgID(ctx, *auth.CurrentOrgID()); err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		RespondJSON(w, userAccoutns)
	}
}

func createUserAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)

	body, err := JsonBody[api.CreateUserAccountRequest](w, r)
	if err != nil {
		return
	}

	var organization types.OrganizationWithBranding
	userAccount := types.UserAccount{
		Email: body.Email,
		Name:  body.Name,
	}

	if err := db.RunTx(ctx, pgx.TxOptions{}, func(ctx context.Context) error {
		if result, err := db.GetOrganizationWithBranding(ctx, *auth.CurrentOrgID()); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return err
		} else {
			organization = *result
		}

		if err := db.CreateUserAccount(ctx, &userAccount); errors.Is(err, apierrors.ErrAlreadyExists) {
			// TODO: In the future this should not be an error, but we don't support multi-org users yet, so for now it is
			http.Error(w, err.Error(), http.StatusBadRequest)
			return err
		} else if err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		if err := db.CreateUserAccountOrganizationAssignment(
			ctx,
			userAccount.ID,
			organization.ID,
			body.UserRole,
		); err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		if err := mailsending.SendUserInviteMail(
			ctx,
			userAccount,
			organization,
			body.UserRole,
			body.ApplicationName,
		); err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		return nil
	}); err != nil {
		log.Warn("could not create user", zap.Error(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteUserAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	userAccount := internalctx.GetUserAccount(ctx)
	auth := auth.Authentication.Require(ctx)
	if userAccount.ID == auth.CurrentUserID() {
		http.Error(w, "UserAccount deleting themselves is not allowed", http.StatusForbidden)
	} else if err := db.DeleteUserAccountWithID(ctx, userAccount.ID); err != nil {
		log.Warn("error deleting user", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func userAccountMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		if userId := r.PathValue("userId"); userId == "" {
			http.Error(w, "missing userId", http.StatusBadRequest)
		} else if userAccount, err := db.GetUserAccountByID(ctx, userId); err != nil {
			if errors.Is(err, apierrors.ErrNotFound) {
				http.NotFound(w, r)
			} else {
				log.Warn("error getting user", zap.Error(err))
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			h.ServeHTTP(w, r.WithContext(internalctx.WithUserAccount(ctx, userAccount)))
		}
	})
}

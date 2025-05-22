package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	"github.com/glasskube/distr/internal/authjwt"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/customdomains"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/mailsending"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func UserAccountsRouter(r chi.Router) {
	r.With(requireUserRoleVendor, middleware.RequireOrgAndRole).Group(func(r chi.Router) {
		r.Get("/", getUserAccountsHandler)
		r.Post("/", createUserAccountHandler)
		r.Route("/{userId}", func(r chi.Router) {
			r.Use(userAccountMiddleware)
			r.Delete("/", deleteUserAccountHandler)
			r.Patch("/image", patchImageUserAccount)
		})
	})
	r.Get("/status", getUserAccountStatusHandler)
}

func getUserAccountsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)
	if userAccounts, err := db.GetUserAccountsByOrgID(ctx, *auth.CurrentOrgID(), nil); err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		RespondJSON(w, api.MapUserAccountsToResponse(userAccounts))
	}
}

func getUserAccountStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)
	userAccount := auth.CurrentUser()
	RespondJSON(w, map[string]any{
		"active": userAccount.PasswordHash != nil,
	})
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
	var inviteURL string
	userHasExisted := false

	if err := db.RunTx(ctx, func(ctx context.Context) error {
		if result, err := db.GetOrganizationWithBranding(ctx, *auth.CurrentOrgID()); err != nil {
			err = fmt.Errorf("failed to get org with branding: %w", err)
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return err
		} else {
			organization = *result
		}

		if existingUA, err := db.GetUserAccountByEmail(ctx, body.Email); errors.Is(err, apierrors.ErrNotFound) {
			if err := db.CreateUserAccount(ctx, &userAccount); err != nil {
				err = fmt.Errorf("failed to create user account: %w", err)
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return err
			}
		} else if err != nil {
			err = fmt.Errorf("failed to get existing user account: %w", err)
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return err
		} else {
			userHasExisted = true
			userAccount = *existingUA
		}

		if err := db.CreateUserAccountOrganizationAssignment(
			ctx,
			userAccount.ID,
			organization.ID,
			body.UserRole,
		); errors.Is(err, apierrors.ErrAlreadyExists) {
			http.Error(w, "user is already part of this organization", http.StatusBadRequest)
			return err
		} else if err != nil {
			err = fmt.Errorf("failed to create user org assignment: %w", err)
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		if !userHasExisted {
			// TODO: Should probably use a different mechanism for invite tokens but for now this should work OK
			if _, token, err := authjwt.GenerateVerificationTokenValidFor(userAccount); err != nil {
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			} else {
				inviteURL = fmt.Sprintf(
					"%v/join?jwt=%v",
					customdomains.AppDomainOrDefault(organization.Organization),
					url.QueryEscape(token),
				)
			}
		}

		if err := mailsending.SendUserInviteMail(
			ctx,
			userAccount,
			organization,
			body.UserRole,
			inviteURL,
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

	RespondJSON(w, api.CreateUserAccountResponse{ID: userAccount.ID, InviteURL: inviteURL})
}

func deleteUserAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	userAccount := internalctx.GetUserAccount(ctx)
	auth := auth.Authentication.Require(ctx)
	if userAccount.ID == auth.CurrentUserID() {
		http.Error(w, "UserAccount deleting themselves is not allowed", http.StatusForbidden)
	} else if managesDts, err := db.UserManagesDeploymentTargetInOrganization(
		ctx, userAccount.ID, *auth.CurrentOrgID()); err != nil {
		log.Error("error checking user/org removal", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else if managesDts {
		http.Error(w, "Please ensure there are no deployments managed by this user and try again", http.StatusBadRequest)
	} else if err := db.DeleteUserAccountFromOrganization(ctx, userAccount.ID, *auth.CurrentOrgID()); err != nil {
		log.Error("error removing user from organization", zap.Error(err))
		if errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNoContent)
		} else {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

var patchImageUserAccount = patchImageHandler(func(ctx context.Context, body api.PatchImageRequest) (any, error) {
	user := internalctx.GetUserAccount(ctx)
	if err := db.UpdateUserAccountImage(ctx, user, body.ImageID); err != nil {
		return nil, err
	} else {
		return api.AsUserAccount(*user), nil
	}
})

func userAccountMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		auth := auth.Authentication.Require(ctx)
		log := internalctx.GetLogger(ctx)
		if userId, err := uuid.Parse(r.PathValue("userId")); err != nil {
			http.NotFound(w, r)
		} else if userAccount, err := db.GetUserAccountWithRole(ctx, userId, *auth.CurrentOrgID()); err != nil {
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

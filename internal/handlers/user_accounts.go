package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	"github.com/glasskube/distr/internal/authjwt"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/customdomains"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/mailsending"
	"github.com/glasskube/distr/internal/mapping"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func UserAccountsRouter(r chi.Router) {
	r.With(middleware.RequireOrgAndRole).Group(func(r chi.Router) {
		r.Get("/", getUserAccountsHandler)
		r.Post("/", createUserAccountHandler)
		r.Route("/{userId}", func(r chi.Router) {
			r.Use(userAccountMiddleware)
			r.Delete("/", deleteUserAccountHandler)
			r.Patch("/image", patchImageUserAccount)
			r.With(inviteUserRateLimiter).
				Post("/invite", resendUserInviteHandler())
		})
	})
	r.Get("/status", getUserAccountStatusHandler)
}

func getUserAccountsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)

	var userAccounts []types.UserAccountWithUserRole
	var err error

	if customerOrgID := auth.CurrentCustomerOrgID(); customerOrgID != nil {
		userAccounts, err = db.GetUserAccountsByCustomerOrgID(ctx, *customerOrgID)
	} else {
		userAccounts, err = db.GetUserAccountsByOrgID(ctx, *auth.CurrentOrgID())
	}

	if err != nil {
		log.Error("failed to get user accounts", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		RespondJSON(w, mapping.List(userAccounts, mapping.UserAccountToAPI))
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

	if customerOrgID := auth.CurrentCustomerOrgID(); customerOrgID != nil {
		if body.UserRole == types.UserRoleVendor {
			http.Error(w, "insufficient permissions", http.StatusForbidden)
			return
		}
		body.CustomerOrganizationID = customerOrgID
	}

	if body.UserRole == types.UserRoleVendor && body.CustomerOrganizationID != nil {
		http.Error(w, "customer organization not applicable for vendor user", http.StatusBadRequest)
		return
	} else if body.UserRole == types.UserRoleCustomer && body.CustomerOrganizationID == nil {
		http.Error(w, "customer organization is required for customer user", http.StatusBadRequest)
		return
	}

	if body.CustomerOrganizationID != nil {
		if co, err := db.GetCustomerOrganizationByID(
			ctx,
			*body.CustomerOrganizationID,
		); errors.Is(err, apierrors.ErrNotFound) || (err == nil && co.OrganizationID != *auth.CurrentOrgID()) {
			http.Error(w, "customer organization does not exist", http.StatusBadRequest)
			return
		} else if err != nil {
			err = fmt.Errorf("failed to get customer organization: %w", err)
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

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
			body.CustomerOrganizationID,
		); errors.Is(err, apierrors.ErrAlreadyExists) {
			http.Error(w, "user is already part of this organization", http.StatusBadRequest)
			return err
		} else if err != nil {
			err = fmt.Errorf("failed to create user org assignment: %w", err)
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		if !userHasExisted || userAccount.EmailVerifiedAt == nil {
			if inviteURL, err = generateUserInviteUrl(userAccount, organization.Organization); err != nil {
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
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

	RespondJSON(w, api.CreateUserAccountResponse{
		User:      userAccount.AsUserAccountWithRole(body.UserRole, body.CustomerOrganizationID, time.Now()),
		InviteURL: inviteURL,
	})
}

func resendUserInviteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		auth := auth.Authentication.Require(ctx)
		userAccountIDStr := r.PathValue("userId")
		userAccountID, err := uuid.Parse(userAccountIDStr)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		userAccount, err := db.GetUserAccountWithRole(ctx, userAccountID, *auth.CurrentOrgID(), auth.CurrentCustomerOrgID())
		if errors.Is(err, apierrors.ErrNotFound) {
			http.NotFound(w, r)
			return
		} else if err != nil {
			err = fmt.Errorf("failed to get org with branding: %w", err)
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		} else if userAccount.EmailVerified {
			http.Error(w, "UserAccount is already verified", http.StatusBadRequest)
			return
		}

		organization, err := db.GetOrganizationWithBranding(ctx, *auth.CurrentOrgID())
		if err != nil {
			err = fmt.Errorf("failed to get org with branding: %w", err)
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		inviteURL, err := generateUserInviteUrl(userAccount.AsUserAccount(), organization.Organization)
		if err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := mailsending.SendUserInviteMail(
			ctx,
			userAccount.AsUserAccount(),
			*organization,
			userAccount.UserRole,
			inviteURL,
		); err != nil {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		RespondJSON(w, api.CreateUserAccountResponse{User: *userAccount, InviteURL: inviteURL})
	}
}

func deleteUserAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	userAccount := internalctx.GetUserAccount(ctx)
	auth := auth.Authentication.Require(ctx)
	if userAccount.ID == auth.CurrentUserID() {
		http.Error(w, "UserAccount deleting themselves is not allowed", http.StatusForbidden)
	} else if err := db.RunTx(ctx, func(ctx context.Context) error {
		if ok, err := userCanBeRemovedFromOrg(ctx, userAccount.ID, *auth.CurrentOrgID()); err != nil {
			return err
		} else if !ok {
			http.Error(w, "Please ensure there are no deployment targets and licenses owned by this user and try again",
				http.StatusBadRequest)
			return nil
		} else if err := db.DeleteUserAccountFromOrganization(ctx, userAccount.ID, *auth.CurrentOrgID()); err != nil {
			if errors.Is(err, apierrors.ErrNotFound) {
				w.WriteHeader(http.StatusNoContent)
				return nil
			} else {
				return err
			}
		} else if err := db.DeleteAccessTokensOfUserInOrg(ctx, userAccount.ID, *auth.CurrentOrgID()); err != nil {
			return err
		} else if err := db.DeleteTutorialProgressesOfUserInOrg(ctx, userAccount.ID, *auth.CurrentOrgID()); err != nil {
			return err
		} else {
			w.WriteHeader(http.StatusNoContent)
			return nil
		}
	}); err != nil {
		log.Error("error removing user from org", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func userCanBeRemovedFromOrg(ctx context.Context, userID, orgID uuid.UUID) (bool, error) {
	if managesDts, err := db.UserManagesDeploymentTargetInOrganization(ctx, userID, orgID); err != nil {
		return false, err
	} else {
		return !managesDts, nil
	}
}

var patchImageUserAccount = patchImageHandler(func(ctx context.Context, body api.PatchImageRequest) (any, error) {
	user := internalctx.GetUserAccount(ctx)
	if err := db.UpdateUserAccountImage(ctx, user, body.ImageID); err != nil {
		return nil, err
	} else {
		return mapping.UserAccountToAPI(*user), nil
	}
})

func userAccountMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		auth := auth.Authentication.Require(ctx)
		log := internalctx.GetLogger(ctx)
		if userId, err := uuid.Parse(r.PathValue("userId")); err != nil {
			http.NotFound(w, r)
		} else if userAccount, err := db.GetUserAccountWithRole(
			ctx,
			userId,
			*auth.CurrentOrgID(),
			auth.CurrentCustomerOrgID(),
		); err != nil {
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

func generateUserInviteUrl(userAccount types.UserAccount, organization types.Organization) (string, error) {
	// TODO: Should probably use a different mechanism for invite tokens but for now this should work OK
	if _, token, err := authjwt.GenerateVerificationTokenValidFor(userAccount); err != nil {
		return "", fmt.Errorf("failed to generate invite URL: %w", err)
	} else {
		return fmt.Sprintf(
			"%v/join?jwt=%v",
			customdomains.AppDomainOrDefault(organization),
			url.QueryEscape(token),
		), nil
	}
}

package handlers

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func OrganizationRouter(r chi.Router) {
	r.Use(middleware.RequireOrgAndRole)
	r.Get("/", getOrganization)
	r.Group(func(r chi.Router) {
		r.Use(requireUserRoleVendor)
		r.Put("/", updateOrganization)
		r.Post("/", createOrganization)
	})
	r.Route("/branding", OrganizationBrandingRouter)
}

func OrganizationsRouter(r chi.Router) {
	r.Use(middleware.RequireOrgAndRole)
	r.Get("/", getOrganizations)
}

func getOrganization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)
	RespondJSON(w, auth.CurrentOrg())
}

func updateOrganization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)

	organization, err := JsonBody[types.Organization](w, r)
	if err != nil {
		return
	} else if ok := validateOrganizationRequest(w, &organization); !ok {
		return
	}

	existingOrganization := auth.CurrentOrg()
	if existingOrganization.Slug != nil && *existingOrganization.Slug != "" {
		if organization.Slug == nil || *organization.Slug == "" {
			http.Error(w, "Slug can not get unset", http.StatusBadRequest)
			return
		}
	}

	if organization.ID == uuid.Nil {
		organization.ID = existingOrganization.ID
	} else if organization.ID != existingOrganization.ID {
		http.Error(w, "organization id does not match", http.StatusBadRequest)
		return
	}

	if err := db.UpdateOrganization(ctx, &organization); errors.Is(err, apierrors.ErrConflict) {
		http.Error(w, "Slug is not available", http.StatusBadRequest)
	} else if err != nil {
		internalctx.GetLogger(ctx).Error("failed to update organization", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		RespondJSON(w, organization)
	}
}

func createOrganization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)
	log := internalctx.GetLogger(ctx)

	organization, err := JsonBody[types.Organization](w, r)
	if err != nil {
		return
	} else if ok := validateOrganizationRequest(w, &organization); !ok {
		return
	}

	if err := db.RunTx(ctx, func(ctx context.Context) error {
		if err := db.CreateOrganization(ctx, &organization); err != nil {
			return err
		}
		if err := db.CreateUserAccountOrganizationAssignment(
			ctx, auth.CurrentUserID(), organization.ID, types.UserRoleVendor); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if errors.Is(err, apierrors.ErrConflict) {
			http.Error(w, "Slug is not available", http.StatusBadRequest)
		} else {
			log.Error("could not create organization", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	} else {
		RespondJSON(w, types.OrganizationWithUserRole{
			Organization: organization,
			UserRole:     types.UserRoleVendor,
			JoinedOrgAt:  time.Now(),
		})
	}
}

func validateOrganizationRequest(w http.ResponseWriter, organization *types.Organization) bool {
	if organization.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return false
	}
	if organization.Slug != nil {
		slugPattern := "^[a-z0-9]+((\\.|_|__|-+)[a-z0-9]+)*$"
		slugMaxLength := 64
		if matched, _ := regexp.MatchString(slugPattern, *organization.Slug); !matched {
			http.Error(w, "Slug is invalid", http.StatusBadRequest)
			return false
		} else if len(*organization.Slug) > slugMaxLength {
			http.Error(w, "Slug too long (max 64 chars)", http.StatusBadRequest)
			return false
		}
	}
	return true
}

func getOrganizations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)

	if orgs, err := db.GetOrganizationsForUser(ctx, auth.CurrentUserID()); err != nil {
		internalctx.GetLogger(ctx).Error("failed to get organizations", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		RespondJSON(w, orgs)
	}
}

package handlers

import (
	"errors"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"net/http"
	"regexp"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func OrganizationRouter(r chi.Router) {
	r.Use(middleware.RequireOrgID, middleware.RequireUserRole)
	r.Get("/", getOrganization)
	r.With(requireUserRoleVendor).Put("/", updateOrganization)
	r.Route("/branding", OrganizationBrandingRouter)

}

func getOrganization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)

	if organization, err :=
		db.GetOrganizationByID(ctx, *auth.CurrentOrgID()); errors.Is(err, apierrors.ErrNotFound) {
		http.NotFound(w, r)
	} else if err != nil {
		internalctx.GetLogger(ctx).Error("failed to get organization", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		RespondJSON(w, organization)
	}
}

func updateOrganization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)

	organization, err := JsonBody[types.Organization](w, r)
	if err != nil {
		return
	} else if organization.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if existingOrganization, err := db.GetOrganizationByID(ctx, *auth.CurrentOrgID()); err != nil {
		internalctx.GetLogger(ctx).Error("failed to get organization before update", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		if existingOrganization.Slug != nil {
			if organization.Slug == nil || *organization.Slug == "" {
				http.Error(w, "Slug can not get unset", http.StatusBadRequest)
				return
			}
		}

		slugPattern := "^[a-z0-9]+((\\.|_|__|-+)[a-z0-9]+)*$"
		slugMaxLength := 64
		if matched, _ := regexp.MatchString(slugPattern, *organization.Slug); !matched {
			http.Error(w, "Slug is invalid", http.StatusBadRequest)
			return
		} else if len(*organization.Slug) > slugMaxLength {
			http.Error(w, "Slug too long (max 64 chars)", http.StatusBadRequest)
			return
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
}

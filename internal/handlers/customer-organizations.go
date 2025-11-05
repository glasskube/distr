package handlers

import (
	"errors"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/mapping"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func CustomerOrganizationsRouter(r chi.Router) {
	r.With(requireUserRoleVendor, middleware.RequireOrgAndRole).Group(func(r chi.Router) {
		r.Get("/", getCustomerOrganizationsHandler())
		r.Post("/", createCustomerOrganizationHandler())
		r.Put("/{id}", updateCustomerOrganizationHandler())
		r.Delete("/{id}", deleteCustomerOrganizationHandler())
	})
}

func getCustomerOrganizationsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		auth := auth.Authentication.Require(ctx)
		if customerOrganizations, err := db.GetCustomerOrganizationsByOrganizationID(ctx, *auth.CurrentOrgID()); err != nil {
			log.Error("failed to get customer orgs", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			RespondJSON(w, mapping.List(customerOrganizations, mapping.CustomerOrganizationWithUserCountToAPI))
		}
	}
}

func createCustomerOrganizationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		auth := auth.Authentication.Require(ctx)
		request, err := JsonBody[api.CreateUpdateCustomerOrganizationRequest](w, r)
		if err != nil {
			return
		}

		customerOrganization := types.CustomerOrganization{
			OrganizationID: *auth.CurrentOrgID(),
			Name:           request.Name,
			ImageID:        request.ImageID,
		}

		if err := db.CreateCustomerOrganization(ctx, &customerOrganization); err != nil {
			log.Error("failed to create customer org", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			RespondJSON(w, mapping.CustomerOrganizationToAPI(customerOrganization))
		}
	}
}

func updateCustomerOrganizationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		auth := auth.Authentication.Require(ctx)
		request, err := JsonBody[api.CreateUpdateCustomerOrganizationRequest](w, r)
		if err != nil {
			return
		}

		customerOrganization := types.CustomerOrganization{
			ID:             id,
			OrganizationID: *auth.CurrentOrgID(),
			Name:           request.Name,
			ImageID:        request.ImageID,
		}

		if err := db.UpdateCustomerOrganization(ctx, &customerOrganization); err != nil {
			log.Error("failed to update customer org", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			RespondJSON(w, mapping.CustomerOrganizationToAPI(customerOrganization))
		}
	}
}

func deleteCustomerOrganizationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		auth := auth.Authentication.Require(ctx)

		if err := db.DeleteCustomerOrganizationWithID(ctx, id, *auth.CurrentOrgID()); errors.Is(err, apierrors.ErrNotFound) {
			http.NotFound(w, r)
		} else if errors.Is(err, apierrors.ErrConflict) {
			http.Error(w, "customer organization is not empty", http.StatusConflict)
		} else if err != nil {
			log.Error("failed to delete customer org", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	"github.com/glasskube/distr/internal/buildconfig"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/mapping"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/subscription"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
	"github.com/google/uuid"
	"github.com/oaswrap/spec/adapter/chiopenapi"
	"github.com/oaswrap/spec/option"
	"go.uber.org/zap"
)

func OrganizationRouter(r chiopenapi.Router) {
	r.WithOptions(option.GroupTags("Organizations"))
	r.Use(middleware.RequireOrgAndRole)
	r.Get("/", getOrganization).
		With(option.Description("Get current organization")).
		With(option.Response(http.StatusOK, api.OrganizationResponse{}))
	r.With(middleware.RequireVendor).Group(func(r chiopenapi.Router) {
		r.With(middleware.RequireAdmin).
			Put("/", updateOrganization).
			With(option.Description("Update current organization")).
			With(option.Request(api.CreateUpdateOrganizationRequest{})).
			With(option.Response(http.StatusOK, types.Organization{}))
		r.Post("/", createOrganization).
			With(option.Description("Create a new organization")).
			With(option.Request(api.CreateUpdateOrganizationRequest{})).
			With(option.Response(http.StatusOK, types.OrganizationWithUserRole{}))
	})
	r.Route("/branding", OrganizationBrandingRouter)
}

func getOrganization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)
	RespondJSON(w, mapping.OrganizationToAPI(*auth.CurrentOrg()))
}

func updateOrganization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)

	body, err := JsonBody[api.CreateUpdateOrganizationRequest](w, r)
	if err != nil {
		return
	} else if ok := validateOrganizationRequest(w, body); !ok {
		return
	}

	if org, err := handleUpdateOrganization(ctx, *auth.CurrentOrgID(), body); err != nil {
		switch {
		case errors.Is(err, apierrors.ErrBadRequest):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, apierrors.ErrConflict):
			http.Error(w, "Slug is not available", http.StatusBadRequest)
		default:
			internalctx.GetLogger(ctx).Error("failed to update organization", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		RespondJSON(w, org)
	}
}

func createOrganization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)
	log := internalctx.GetLogger(ctx)

	body, err := JsonBody[api.CreateUpdateOrganizationRequest](w, r)
	if err != nil {
		return
	} else if ok := validateOrganizationRequest(w, body); !ok {
		return
	}

	organization := types.Organization{
		Name:                body.Name,
		Slug:                body.Slug,
		SubscriptionType:    types.SubscriptionTypeTrial,
		Features:            subscription.ProFeatures,
		PreConnectScript:    body.PreConnectScript,
		PostConnectScript:   body.PostConnectScript,
		ConnectScriptIsSudo: body.ConnectScriptIsSudo,
	}

	if buildconfig.IsCommunityEdition() {
		organization.SubscriptionType = types.SubscriptionTypeCommunity
		organization.Features = []types.Feature{}
	}

	if err := db.RunTx(ctx, func(ctx context.Context) error {
		if err := db.CreateOrganization(ctx, &organization); err != nil {
			return err
		}
		if err := db.CreateUserAccountOrganizationAssignment(
			ctx, auth.CurrentUserID(), organization.ID, types.UserRoleAdmin, nil); err != nil {
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
			UserRole:     types.UserRoleAdmin,
			JoinedOrgAt:  time.Now(),
		})
	}
}

func validateOrganizationRequest(w http.ResponseWriter, organization api.CreateUpdateOrganizationRequest) bool {
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

func validateOrganizationUpdate(body api.CreateUpdateOrganizationRequest, org types.Organization) error {
	if org.Slug != nil && *org.Slug != "" {
		if body.Slug == nil || *body.Slug == "" {
			return fmt.Errorf("%w: slug can not get unset", apierrors.ErrBadRequest)
		}
	}

	return nil
}

func handleUpdateOrganization(
	ctx context.Context,
	id uuid.UUID,
	request api.CreateUpdateOrganizationRequest,
) (*types.Organization, error) {
	var org *types.Organization

	err := db.RunTxRR(ctx, func(ctx context.Context) error {
		var err error
		org, err = db.GetOrganizationByID(ctx, id)
		if err != nil {
			return err
		}

		if err := validateOrganizationUpdate(request, *org); err != nil {
			return err
		}

		needsUpdate := false

		if org.Name != request.Name {
			org.Name = request.Name
			needsUpdate = true
		}

		if !util.PtrEq(org.Slug, request.Slug) {
			org.Slug = request.Slug
			needsUpdate = true
		}

		if !util.PtrEq(org.PreConnectScript, request.PreConnectScript) {
			org.PreConnectScript = request.PreConnectScript
			needsUpdate = true
		}

		if !util.PtrEq(org.PostConnectScript, request.PostConnectScript) {
			org.PostConnectScript = request.PostConnectScript
			needsUpdate = true
		}

		if request.ConnectScriptIsSudo != org.ConnectScriptIsSudo {
			org.ConnectScriptIsSudo = request.ConnectScriptIsSudo
			needsUpdate = true
		}

		if needsUpdate {
			return db.UpdateOrganization(ctx, org)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return org, nil
}

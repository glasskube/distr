package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func OrganizationBrandingRouter(r chi.Router) {
	r.Use(middleware.RequireOrgID, middleware.RequireUserRole)
	r.Get("/", getOrganizationBranding)
	r.With(requireUserRoleVendor).Group(func(r chi.Router) {
		r.Post("/", createOrganizationBranding)
		r.Put("/{organizationBrandingId}", updateOrganizationBranding)
	})
}

func getOrganizationBranding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)

	if organizationBranding, err :=
		db.GetOrganizationBranding(r.Context(), *auth.CurrentOrgID()); errors.Is(err, apierrors.ErrNotFound) {
		http.NotFound(w, r)
	} else if err != nil {
		internalctx.GetLogger(r.Context()).Error("failed to get organizationBranding", zap.Error(err))
		sentry.GetHubFromContext(r.Context()).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		RespondJSON(w, organizationBranding)
	}
}

func createOrganizationBranding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)

	if organizationBranding, err := getOrganizationBrandingFromRequest(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if err := setMetadataForOrganizationBranding(ctx, organizationBranding); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if err = db.CreateOrganizationBranding(r.Context(), organizationBranding); err != nil {
		log.Warn("could not create organizationBranding", zap.Error(err))
		sentry.GetHubFromContext(r.Context()).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		RespondJSON(w, organizationBranding)
	}
}

func updateOrganizationBranding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)

	if organizationBranding, err := getOrganizationBrandingFromRequest(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if err := setMetadataForOrganizationBranding(ctx, organizationBranding); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if err = db.UpdateOrganizationBranding(r.Context(), organizationBranding); err != nil {
		log.Warn("could not create organizationBranding", zap.Error(err))
		sentry.GetHubFromContext(r.Context()).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		RespondJSON(w, organizationBranding)
	}
}

func getOrganizationBrandingFromRequest(r *http.Request) (*types.OrganizationBranding, error) {
	if err := r.ParseMultipartForm(102400); err != nil {
		return nil, fmt.Errorf("failed to parse form: %w", err)
	}
	organizationBranding := types.OrganizationBranding{
		Title:       util.PtrTo(r.Form.Get("title")),
		Description: util.PtrTo(r.Form.Get("description")),
	}

	if brandingID, err := uuid.Parse(r.PathValue("organizationBrandingId")); err != nil {
		return nil, err
	} else {
		organizationBranding.ID = brandingID
	}

	if file, head, err := r.FormFile("logo"); err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			return nil, err
		} else {
			// no logo uploaded
			organizationBranding.Logo = nil
			organizationBranding.LogoFileName = nil
			organizationBranding.LogoContentType = nil
		}
	} else if head.Size > 102400 {
		return nil, errors.New("file too large (max 100 KiB)")
	} else if data, err := io.ReadAll(file); err != nil {
		return nil, err
	} else {
		organizationBranding.Logo = data
		organizationBranding.LogoFileName = &head.Filename
		organizationBranding.LogoContentType = util.PtrTo(head.Header.Get("Content-Type"))
	}

	return &organizationBranding, nil
}

func setMetadataForOrganizationBranding(ctx context.Context, t *types.OrganizationBranding) error {
	if auth, err := auth.Authentication.Get(ctx); err != nil {
		return err
	} else {
		t.OrganizationID = *auth.CurrentOrgID()
		t.UpdatedByUserAccountID = util.PtrTo(auth.CurrentUserID())
		t.UpdatedAt = time.Now()
		return nil
	}
}

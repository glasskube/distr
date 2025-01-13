package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/glasskube/cloud/internal/apierrors"
	"github.com/glasskube/cloud/internal/util"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/cloud/internal/auth"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/types"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func OrganizationBrandingRouter(r chi.Router) {
	r.Get("/", getOrganizationBranding)
	r.With(requireUserRoleVendor).Group(func(r chi.Router) {
		r.Post("/", createOrganizationBranding)
		r.Put("/{organizationBrandingId}", updateOrganizationBranding)
	})
}

func getOrganizationBranding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if orgID, err := auth.CurrentOrgId(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if organizationBranding, err :=
		db.GetOrganizationBranding(r.Context(), orgID); errors.Is(err, apierrors.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
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

	if organizationBranding, err := parseRequest(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if err := setMetadata(organizationBranding, ctx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if err = db.CreateOrganizationBranding(r.Context(), organizationBranding); err != nil {
		log.Warn("could not create organizationBranding", zap.Error(err))
		sentry.GetHubFromContext(r.Context()).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err = fmt.Fprintln(w, err); err != nil {
			log.Error("failed to write error to response", zap.Error(err))
		}
	} else if err = json.NewEncoder(w).Encode(organizationBranding); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func updateOrganizationBranding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)

	if organizationBranding, err := parseRequest(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if err := setMetadata(organizationBranding, ctx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if err = db.UpdateOrganizationBranding(r.Context(), organizationBranding); err != nil {
		log.Warn("could not create organizationBranding", zap.Error(err))
		sentry.GetHubFromContext(r.Context()).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err = fmt.Fprintln(w, err); err != nil {
			log.Error("failed to write error to response", zap.Error(err))
		}
	} else if err = json.NewEncoder(w).Encode(organizationBranding); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func parseRequest(r *http.Request) (*types.OrganizationBranding, error) {
	if err := r.ParseMultipartForm(102400); err != nil {
		return nil, fmt.Errorf("failed to parse form: %w", err)
	}
	organizationBranding := types.OrganizationBranding{
		Title:       r.Form.Get("title"),
		Description: r.Form.Get("description"),
	}

	organizationBranding.ID = r.PathValue("organizationBrandingId")

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

func setMetadata(t *types.OrganizationBranding, ctx context.Context) error {
	if orgID, err := auth.CurrentOrgId(ctx); err != nil {
		return err
	} else if id, err := auth.CurrentUserId(ctx); err != nil {
		return err
	} else {
		t.OrganizationID = orgID
		t.UpdatedByUserAccountID = id
		t.UpdatedAt = time.Now()
	}
	return nil
}

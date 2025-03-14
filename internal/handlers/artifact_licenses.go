package handlers

import (
	"context"
	"errors"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func ArtifactLicensesRouter(r chi.Router) {
	r.Use(middleware.RequireOrgID, middleware.RequireUserRole, requireUserRoleVendor)
	r.Get("/", getArtifactLicenses)
	r.Post("/", createArtifactLicense)
	r.Route("/{artifactLicenseId}", func(r chi.Router) {
		r.With(artifactLicenseMiddleware).Group(func(r chi.Router) {
			r.With(requireUserRoleVendor).Put("/", updateArtifactLicense)
		})
	})
}

func getArtifactLicenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)

	if licenses, err := db.GetArtifactLicenses(ctx, *auth.CurrentOrgID()); err != nil {
		log.Error("failed to get artifact licenses", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else {
		RespondJSON(w, licenses)
	}
}

func createArtifactLicense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)

	license, err := JsonBody[types.ArtifactLicense](w, r)
	if err != nil {
		return
	}
	license.OrganizationID = *auth.CurrentOrgID()

	_ = db.RunTx(ctx, pgx.TxOptions{}, func(ctx context.Context) error {
		if err := db.CreateArtifactLicense(ctx, &license.ArtifactLicenseBase); errors.Is(err, apierrors.ErrConflict) {
			http.Error(w, "An artifact license with this name already exists", http.StatusBadRequest)
			return err
		} else if err != nil {
			log.Warn("could not create artifact license", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		for _, selection := range license.Artifacts {
			if len(selection.Versions) == 0 {
				if err := db.AddArtifactToArtifactLicense(ctx, license.ID, selection.Artifact.ID, nil); err != nil {
					log.Warn("could not add version to license", zap.Error(err))
					sentry.GetHubFromContext(ctx).CaptureException(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return err
				}
			}
			for _, version := range selection.Versions {
				if err := db.AddArtifactToArtifactLicense(ctx, license.ID, selection.Artifact.ID, &version.ID); err != nil {
					log.Warn("could not add version to license", zap.Error(err))
					sentry.GetHubFromContext(ctx).CaptureException(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return err
				}
			}
		}

		// TODO maybe completely read again
		RespondJSON(w, license)
		return nil
	})
}

func updateArtifactLicense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)

	license, err := JsonBody[types.ArtifactLicense](w, r)
	if err != nil {
		return
	}
	license.OrganizationID = *auth.CurrentOrgID()

	existing := internalctx.GetArtifactLicense(ctx)
	if license.ID == uuid.Nil {
		license.ID = existing.ID
	} else if license.ID != existing.ID {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if existing.OwnerUserAccountID != nil &&
		(license.OwnerUserAccountID == nil || *existing.OwnerUserAccountID != *license.OwnerUserAccountID) {
		http.Error(w, "Changing the license owner is not allowed", http.StatusBadRequest)
		return
	}

	// TODO
	/*
		_ = db.RunTx(ctx, pgx.TxOptions{}, func(ctx context.Context) error {
			if err := db.CreateArtifactLicense(ctx, &license.ArtifactLicenseBase); errors.Is(err, apierrors.ErrConflict) {
				http.Error(w, "An artifact license with this name already exists", http.StatusBadRequest)
				return err
			} else if err != nil {
				log.Warn("could not create artifact license", zap.Error(err))
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			// TODO maybe completely read again
			RespondJSON(w, license)
			return nil
		})*/
}

func artifactLicenseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if licenseId, err := uuid.Parse(r.PathValue("artifactLicenseId")); err != nil {
			http.Error(w, "artifactLicenseId is not a valid UUID", http.StatusBadRequest)
		} else if license, err := db.GetArtifactLicenseByID(ctx, licenseId); errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else if err != nil {
			internalctx.GetLogger(r.Context()).Error("failed to get license", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			ctx = internalctx.WithArtifactLicense(ctx, license)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

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
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func ArtifactsRouter(r chi.Router) {
	r.Use(middleware.RequireOrgAndRole)
	r.Get("/", getArtifacts)
	r.Route("/{artifactId}", func(r chi.Router) {
		r.Use(artifactMiddleware)
		r.Get("/", getArtifact)
		r.With(requireUserRoleVendor).Patch("/image", patchImageArtifactHandler)
		r.With(requireUserRoleVendor).Delete("/", deleteArtifactHandler)
		r.Route("/versions/{versionId}", func(r chi.Router) {
			r.With(requireUserRoleVendor).Delete("/", deleteArtifactVersionHandler)
		})
	})
}

func getArtifacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)

	var artifacts []types.ArtifactWithDownloads
	var err error
	if *auth.CurrentUserRole() == types.UserRoleCustomer && auth.CurrentOrg().HasFeature(types.FeatureLicensing) {
		artifacts, err = db.GetArtifactsByLicenseOwnerID(ctx, *auth.CurrentOrgID(), auth.CurrentUserID())
	} else {
		artifacts, err = db.GetArtifactsByOrgID(ctx, *auth.CurrentOrgID())
	}

	if err != nil {
		log.Error("failed to get artifacts", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else {
		RespondJSON(w, api.MapArtifactsToResponse(artifacts))
	}
}

func getArtifact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	RespondJSON(w, api.AsArtifact(*internalctx.GetArtifact(ctx)))
}

var patchImageArtifactHandler = patchImageHandler(func(ctx context.Context, body api.PatchImageRequest) (any, error) {
	artifact := internalctx.GetArtifact(ctx)
	if err := db.UpdateArtifactImage(ctx, artifact, body.ImageID); err != nil {
		return nil, err
	} else {
		return api.AsArtifact(*artifact), nil
	}
})

func deleteArtifactHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	artifact := internalctx.GetArtifact(ctx)

	if err := db.RunTx(ctx, func(ctx context.Context) error {
		if isReferenced, err := db.ArtifactIsReferencedInLicenses(ctx, artifact.ID); err != nil {
			return err
		} else if isReferenced {
			http.Error(w, "Cannot delete artifact: it is referenced in one or more licenses. Please remove the artifact from all licenses before deleting.", http.StatusBadRequest)
			return nil
		} else if err := db.DeleteArtifactWithID(ctx, artifact.ID); err != nil {
			if errors.Is(err, apierrors.ErrNotFound) {
				w.WriteHeader(http.StatusNoContent)
				return nil
			} else {
				return err
			}
		} else {
			w.WriteHeader(http.StatusNoContent)
			return nil
		}
	}); err != nil {
		log.Error("error deleting artifact", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func deleteArtifactVersionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)

	versionId, parseErr := uuid.Parse(r.PathValue("versionId"))
	if parseErr != nil {
		http.NotFound(w, r)
		return
	}

	if err := db.RunTx(ctx, func(ctx context.Context) error {
		if isReferenced, err := db.ArtifactVersionIsReferencedInLicenses(ctx, versionId); err != nil {
			return err
		} else if isReferenced {
			http.Error(w, "Cannot delete artifact version: it is referenced in one or more licenses. Please remove the version from all licenses before deleting.", http.StatusBadRequest)
			return nil
		} else if err := db.DeleteArtifactVersionWithID(ctx, versionId); err != nil {
			if errors.Is(err, apierrors.ErrNotFound) {
				w.WriteHeader(http.StatusNoContent)
				return nil
			} else {
				return err
			}
		} else {
			w.WriteHeader(http.StatusNoContent)
			return nil
		}
	}); err != nil {
		log.Error("error deleting artifact version", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func artifactMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		auth := auth.Authentication.Require(ctx)

		var artifact *types.ArtifactWithTaggedVersion
		var err error

		if artifactId, parseErr := uuid.Parse(r.PathValue("artifactId")); parseErr != nil {
			http.NotFound(w, r)
			return
		} else if *auth.CurrentUserRole() == types.UserRoleCustomer && auth.CurrentOrg().HasFeature(types.FeatureLicensing) {
			artifact, err = db.GetArtifactByID(ctx, *auth.CurrentOrgID(), artifactId, util.PtrTo(auth.CurrentUserID()))
		} else {
			artifact, err = db.GetArtifactByID(ctx, *auth.CurrentOrgID(), artifactId, nil)
		}

		if err != nil {
			if errors.Is(err, apierrors.ErrNotFound) {
				http.NotFound(w, r)
			} else {
				log.Error("failed to get artifact", zap.Error(err))
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		} else {
			h.ServeHTTP(w, r.WithContext(internalctx.WithArtifact(ctx, artifact)))
		}
	})
}

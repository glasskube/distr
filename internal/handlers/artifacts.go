package handlers

import (
	"errors"
	"github.com/glasskube/distr/api"
	"net/http"

	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/util"
	"github.com/google/uuid"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func ArtifactsRouter(r chi.Router) {
	r.Use(middleware.RequireOrgAndRole, middleware.RegistryFeatureFlagEnabledMiddleware)
	r.Get("/", getArtifacts)
	r.Route("/{artifactId}", func(r chi.Router) {
		r.Use(artifactMiddleware)
		r.Get("/", getArtifact)
		r.With(requireUserRoleVendor).Patch("/image", patchImageArtifactHandler)
	})
}

func getArtifacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)

	var artifacts []types.ArtifactWithDownloads
	var err error
	if *auth.CurrentUserRole() == types.UserRoleCustomer {
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
	RespondJSON(w, api.AsArtifact(internalctx.GetArtifact(ctx)))
}

func patchImageArtifactHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	artifact := internalctx.GetArtifact(ctx)

	body, err := JsonBody[types.PatchImageRequest](w, r)

	if err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if body.ImageID == uuid.Nil {
		http.Error(w, "imageId can not be empty", http.StatusBadRequest)
		return
	}

	if err := db.UpdateArtifactImage(ctx, artifact, body.ImageID); err != nil {
		log.Warn("error patching user image id", zap.Error(err))
		if errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNoContent)
		} else if errors.Is(err, apierrors.ErrConflict) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		RespondJSON(w, api.AsArtifact(artifact))
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
		} else if *auth.CurrentUserRole() == types.UserRoleCustomer {
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

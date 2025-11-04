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
	"github.com/glasskube/distr/internal/mapping"
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
	})
}

func getArtifacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)

	var artifacts []types.ArtifactWithDownloads
	var err error
	if *auth.CurrentUserRole() == types.UserRoleCustomer && auth.CurrentOrg().HasFeature(types.FeatureLicensing) {
		artifacts, err = db.GetArtifactsByLicenseOwnerID(ctx, *auth.CurrentOrgID(), *auth.CurrentCustomerOrgID())
	} else {
		artifacts, err = db.GetArtifactsByOrgID(ctx, *auth.CurrentOrgID())
	}

	if err != nil {
		log.Error("failed to get artifacts", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else {
		RespondJSON(w, mapping.List(artifacts, mapping.ArtifactsWithDownloadsToAPI))
	}
}

func getArtifact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	RespondJSON(w, mapping.ArtifactToAPI(*internalctx.GetArtifact(ctx)))
}

var patchImageArtifactHandler = patchImageHandler(func(ctx context.Context, body api.PatchImageRequest) (any, error) {
	artifact := internalctx.GetArtifact(ctx)
	if err := db.UpdateArtifactImage(ctx, artifact, body.ImageID); err != nil {
		return nil, err
	} else {
		return mapping.ArtifactToAPI(*artifact), nil
	}
})

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

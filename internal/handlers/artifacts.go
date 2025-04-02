package handlers

import (
	"errors"
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
	r.Get("/{artifactId}", getArtifact)
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
		RespondJSON(w, artifacts)
	}
}

func getArtifact(w http.ResponseWriter, r *http.Request) {
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
		RespondJSON(w, artifact)
	}
}

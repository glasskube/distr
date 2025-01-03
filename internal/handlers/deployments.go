package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/cloud/internal/apierrors"
	"github.com/glasskube/cloud/internal/auth"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/types"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func DeploymentsRouter(r chi.Router) {
	r.Get("/", getDeployments)
	r.Post("/", createDeployment)
	r.Route("/{deploymentId}", func(r chi.Router) {
		r.Use(deploymentMiddleware)
		r.Get("/", getDeployment)
	})
}

func createDeployment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)

	deployment, err := JsonBody[types.Deployment](w, r)
	if err != nil {
		return
	}

	if orgId, err := auth.CurrentOrgId(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if application, err := db.GetApplicationForApplicationVersionID(
		ctx, deployment.ApplicationVersionId,
	); errors.Is(err, apierrors.ErrNotFound) {
		http.Error(w, "application does not exist", http.StatusBadRequest)
	} else if err != nil {
		http.Error(w, "an inernal error occurred", http.StatusInternalServerError)
	} else if deploymentTarget, err := db.GetDeploymentTarget(
		ctx, deployment.DeploymentTargetId, &orgId,
	); errors.Is(err, apierrors.ErrNotFound) {
		http.Error(w, "deployment target does not exist", http.StatusBadRequest)
	} else if err != nil {
		http.Error(w, "an inernal error occurred", http.StatusInternalServerError)
	} else if deploymentTarget.Type != application.Type {
		http.Error(w, "application and deployment target must have the same type", http.StatusBadRequest)
	} else if err = db.CreateDeployment(r.Context(), &deployment); err != nil {
		log.Warn("could not create deployment", zap.Error(err))
		sentry.GetHubFromContext(r.Context()).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err = fmt.Fprintln(w, err); err != nil {
			log.Error("failed to write error to response", zap.Error(err))
		}
	} else if err = json.NewEncoder(w).Encode(deployment); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func getDeployments(w http.ResponseWriter, r *http.Request) {
	deploymentTargetId := r.URL.Query().Get("deploymentTargetId")
	if deployments, err := db.GetDeploymentsForDeploymentTarget(r.Context(), deploymentTargetId); err != nil {
		internalctx.GetLogger(r.Context()).Error("failed to get deployments", zap.Error(err))
		sentry.GetHubFromContext(r.Context()).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		err := json.NewEncoder(w).Encode(deployments)
		if err != nil {
			internalctx.GetLogger(r.Context()).Error("failed to encode to json", zap.Error(err))
		}
	}
}

func getDeployment(w http.ResponseWriter, r *http.Request) {
	deployment := internalctx.GetDeployment(r.Context())
	err := json.NewEncoder(w).Encode(deployment)
	if err != nil {
		internalctx.GetLogger(r.Context()).Error("failed to encode to json", zap.Error(err))
	}
}

func deploymentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		deploymentId := r.PathValue("deploymentId")
		deployment, err := db.GetDeployment(ctx, deploymentId)
		if errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else if err != nil {
			internalctx.GetLogger(r.Context()).Error("failed to get deployment", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			ctx = internalctx.WithDeployment(ctx, deployment)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

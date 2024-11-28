package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/glasskube/cloud/internal/apierrors"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/types"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func DeploymentsRouter(r chi.Router) {
	// TODO r.Use(AuthMiddleware)
	r.Get("/", getDeployments)
	r.Post("/", createDeployment)
	r.Route("/{deploymentId}", func(r chi.Router) {
		r.Use(deploymentMiddleware)
		r.Get("/", getDeployment)
	})
}

func createDeployment(w http.ResponseWriter, r *http.Request) {
	log := internalctx.GetLogger(r.Context())
	var deployment types.Deployment
	if err := json.NewDecoder(r.Body).Decode(&deployment); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if err = db.CreateDeployment(r.Context(), &deployment); err != nil {
		log.Warn("could not create deployment", zap.Error(err))
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
		if errors.Is(err, apierrors.NotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else if err != nil {
			internalctx.GetLogger(r.Context()).Error("failed to get deployment", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			ctx = internalctx.WithDeployment(ctx, deployment)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

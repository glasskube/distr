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

func DeploymentTargetsRouter(r chi.Router) {
	// TODO r.Use(AuthMiddleware)
	r.Get("/", getDeploymentTargets)
	r.Post("/", createDeploymentTarget)
	r.Route("/{deploymentTargetId}", func(r chi.Router) {
		r.Use(deploymentTargetMiddelware)
		r.Get("/", getDeploymentTarget)
		r.Put("/", updateDeploymentTarget)
	})
}

func getDeploymentTargets(w http.ResponseWriter, r *http.Request) {
	if deploymentTargets, err := db.GetDeploymentTargets(r.Context()); err != nil {
		internalctx.GetLoggerOrPanic(r.Context()).Error("failed to get DeploymentTargets", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		err := json.NewEncoder(w).Encode(deploymentTargets)
		if err != nil {
			internalctx.GetLoggerOrPanic(r.Context()).Error("failed to encode to json", zap.Error(err))
		}
	}
}

func getDeploymentTarget(w http.ResponseWriter, r *http.Request) {
	dt := internalctx.GetDeploymentTarget(r.Context())
	err := json.NewEncoder(w).Encode(dt)
	if err != nil {
		internalctx.GetLoggerOrPanic(r.Context()).Error("failed to encode to json", zap.Error(err))
	}
}

func createDeploymentTarget(w http.ResponseWriter, r *http.Request) {
	log := internalctx.GetLoggerOrPanic(r.Context())
	var dt types.DeploymentTarget
	if err := json.NewDecoder(r.Body).Decode(&dt); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, err)
	} else if err = db.CreateDeploymentTarget(r.Context(), &dt); err != nil {
		log.Warn("could not create DeploymentTarget", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
	} else if err = json.NewEncoder(w).Encode(dt); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func updateDeploymentTarget(w http.ResponseWriter, r *http.Request) {
	log := internalctx.GetLoggerOrPanic(r.Context())
	var dt types.DeploymentTarget
	if err := json.NewDecoder(r.Body).Decode(&dt); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, err)
		return
	}

	existing := internalctx.GetDeploymentTarget(r.Context())
	if dt.ID == "" {
		dt.ID = existing.ID
	} else if dt.ID != existing.ID {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "wrong id")
		return
	}

	if err := db.UpdateDeploymentTarget(r.Context(), &dt); err != nil {
		log.Warn("could not update DeploymentTarget", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
	} else if err = json.NewEncoder(w).Encode(dt); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func deploymentTargetMiddelware(wh http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.PathValue("deploymentTargetId")
		deploymentTarget, err := db.GetDeploymentTarget(ctx, id)
		if errors.Is(err, apierrors.NotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else if err != nil {
			internalctx.GetLoggerOrPanic(r.Context()).Error("failed to get application", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			ctx = internalctx.WithDeploymentTarget(ctx, deploymentTarget)
			wh.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

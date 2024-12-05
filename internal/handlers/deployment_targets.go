package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/glasskube/cloud/api"
	"github.com/glasskube/cloud/internal/apierrors"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/security"
	"github.com/glasskube/cloud/internal/types"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

func DeploymentTargetsRouter(r chi.Router) {
	r.Get("/", getDeploymentTargets)
	r.Post("/", createDeploymentTarget)
	r.Route("/{deploymentTargetId}", func(r chi.Router) {
		r.Use(deploymentTargetMiddelware)
		r.Get("/", getDeploymentTarget)
		r.Get("/latest-deployment", getLatestDeployment)
		r.Put("/", updateDeploymentTarget)
		r.Post("/access-request", createAccessForDeploymentTarget)
		r.Get("/connect", connectDeploymentTarget)
	})
}

func getLatestDeployment(w http.ResponseWriter, r *http.Request) {
	dt := internalctx.GetDeploymentTarget(r.Context())
	if deployment, err := db.GetLatestDeploymentForDeploymentTarget(r.Context(), dt.ID); err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			internalctx.GetLogger(r.Context()).Error("failed to get latest deployment", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		err := json.NewEncoder(w).Encode(deployment)
		if err != nil {
			internalctx.GetLogger(r.Context()).Error("failed to encode to json", zap.Error(err))
		}
	}
}

func getDeploymentTargets(w http.ResponseWriter, r *http.Request) {
	if deploymentTargets, err := db.GetDeploymentTargets(r.Context()); err != nil {
		internalctx.GetLogger(r.Context()).Error("failed to get DeploymentTargets", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		err := json.NewEncoder(w).Encode(deploymentTargets)
		if err != nil {
			internalctx.GetLogger(r.Context()).Error("failed to encode to json", zap.Error(err))
		}
	}
}

func getDeploymentTarget(w http.ResponseWriter, r *http.Request) {
	dt := internalctx.GetDeploymentTarget(r.Context())
	err := json.NewEncoder(w).Encode(dt)
	if err != nil {
		internalctx.GetLogger(r.Context()).Error("failed to encode to json", zap.Error(err))
	}
}

func createDeploymentTarget(w http.ResponseWriter, r *http.Request) {
	log := internalctx.GetLogger(r.Context())
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
	log := internalctx.GetLogger(r.Context())
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

func createAccessForDeploymentTarget(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	deploymentTarget := internalctx.GetDeploymentTarget(ctx)

	if deploymentTarget.CurrentStatus != nil {
		// TODO test
		log.Warn("access key cannot be regenerated because deployment target has already been connected")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var accessKey string
	var err error
	if accessKey, err = security.GenerateAccessKey(); err != nil {
		log.Error("failed to generate access key", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if salt, hash, err := security.HashAccessKey(accessKey); err != nil {
		log.Error("failed to hash access key", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		deploymentTarget.AccessKeySalt = &salt
		deploymentTarget.AccessKeyHash = &hash
	}

	if err := db.UpdateDeploymentTargetAccess(ctx, deploymentTarget); err != nil {
		log.Warn("could not update DeploymentTarget", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
	} else if err = json.NewEncoder(w).Encode(api.DeploymentTargetAccessTokenResponse{
		Token: accessKey,
	}); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func connectDeploymentTarget(w http.ResponseWriter, r *http.Request) {

}

func deploymentTargetMiddelware(wh http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.PathValue("deploymentTargetId")
		deploymentTarget, err := db.GetDeploymentTarget(ctx, id)
		if errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else if err != nil {
			internalctx.GetLogger(r.Context()).Error("failed to get DeploymentTarget", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			ctx = internalctx.WithDeploymentTarget(ctx, deploymentTarget)
			wh.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

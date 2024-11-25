package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
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
	w.WriteHeader(http.StatusInternalServerError)
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
	w.WriteHeader(http.StatusInternalServerError)
	// TODO: implement
	fmt.Fprintln(w, "not implemented")
}

func createDeploymentTarget(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	// TODO: implement
	fmt.Fprintln(w, "not implemented")
}

func updateDeploymentTarget(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	// TODO: implement
	fmt.Fprintln(w, "not implemented")
}

func deploymentTargetMiddelware(wh http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement
		wh.ServeHTTP(w, r)
	})
}

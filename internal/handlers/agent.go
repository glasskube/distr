package handlers

import (
	"errors"
	"github.com/glasskube/cloud/internal/apierrors"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/resources"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"text/template"
)

func AgentRouter(r chi.Router) {
	r.Get("/connect", connect)
	r.Get("/resources", downloadResources)
}

func connect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	accessKeyId := r.URL.Query().Get("accessKeyId")
	accessKeySecret := r.URL.Query().Get("accessKeySecret")

	if accessKeyId == "" || accessKeySecret == "" {
		// TODO test
		log.Warn("connect endpoint called without accessKeyId or accessKeySecret", zap.String("accessKeyId", accessKeyId))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	deploymentTarget, err := db.GetDeploymentTargetUnauthenticated(ctx, accessKeyId)
	if errors.Is(err, apierrors.ErrNotFound) {
		log.Warn("connect failed: deployment target not found", zap.String("accessKeyId", accessKeyId))
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if err != nil {
		log.Error("failed to get DeploymentTarget", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if deploymentTarget.CurrentStatus != nil {
		// TODO test
		log.Warn("deployment target has already been connected")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO compare token

	w.Header().Add("Content-Type", "application/yaml")
	if yamlTemplate, err := resources.Get("embedded/agent-base.yaml"); err != nil {
		log.Error("failed to get agent yaml template", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	} else if tmpl, err := template.New("agent").Parse(string(yamlTemplate)); err != nil {
		log.Error("failed to get parse yaml template", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	} else if err := tmpl.Execute(w, map[string]string{
		"accessKeyId":     accessKeyId,
		"accessKeySecret": accessKeySecret,
	}); err != nil {
		log.Error("failed to execute yaml template", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func downloadResources(w http.ResponseWriter, r *http.Request) {

}

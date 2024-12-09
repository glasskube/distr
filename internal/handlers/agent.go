package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/glasskube/cloud/internal/types"
	"net/http"
	"net/url"
	"text/template"

	"github.com/glasskube/cloud/internal/apierrors"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/env"
	"github.com/glasskube/cloud/internal/resources"
	"github.com/glasskube/cloud/internal/security"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AgentRouter(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(queryAuthDeploymentTargetCtxMiddleware)
		r.Get("/connect", connect)
	})
	r.Group(func(r chi.Router) {
		r.Use(basicAuthDeploymentTargetCtxMiddleware)
		r.Get("/agent/resources", downloadResources)
		r.Get("/agent/status", postAgentStatus)
	})
}

func connect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	deploymentTarget := internalctx.GetDeploymentTarget(ctx)
	if deploymentTarget.CurrentStatus != nil {
		log.Warn("deployment target has already been connected")
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.Header().Add("Content-Type", "application/yaml")
		if yamlTemplate, err := resources.Get("embedded/agent-base.yaml"); err != nil {
			log.Error("failed to get agent yaml template", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		} else if tmpl, err := template.New("agent").Parse(string(yamlTemplate)); err != nil {
			log.Error("failed to get parse yaml template", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		} else if resourcesUrl, err := buildResourcesUrl(); err != nil {
			log.Error("failed to build resources url", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		} else if err := tmpl.Execute(w, map[string]string{
			"resourcesUrl":    resourcesUrl,
			"accessKeyId":     r.URL.Query().Get("accessKeyId"),
			"accessKeySecret": r.URL.Query().Get("accessKeySecret"),
		}); err != nil {
			log.Error("failed to execute yaml template", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func buildResourcesUrl() (string, error) {
	if u, err := url.Parse(env.Host()); err != nil {
		return "", err
	} else {
		u = u.JoinPath("/api/v1/agent/resources")
		return u.String(), nil
	}
}

func downloadResources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	deploymentTarget := internalctx.GetDeploymentTarget(ctx)
	var statusMessage string
	if composeFileData, err :=
		db.GetLatestDeploymentComposeFileUnauthenticated(ctx, deploymentTarget.ID); err != nil && !errors.Is(err, apierrors.ErrNotFound) {
		msg := "failed to get compose file from DB"
		statusMessage = fmt.Sprintf("%v: %v", msg, err.Error())
		log.Error(msg, zap.Error(err), zap.String("deploymentTargetId", deploymentTarget.ID))
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		if errors.Is(err, apierrors.ErrNotFound) {
			statusMessage = "EMPTY"
		} else {
			statusMessage = "OK"
		}
		w.Header().Add("Content-Type", "application/yaml")
		if _, err := w.Write(composeFileData); err != nil {
			msg := "failed to write compose file"
			statusMessage = fmt.Sprintf("%v: %v", msg, err.Error())
			log.Error(msg, zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
	if err := db.CreateDeploymentTargetStatus(ctx, deploymentTarget, statusMessage); err != nil {
		log.Error("failed to create deployment target status", zap.Error(err),
			zap.String("deploymentTargetId", deploymentTarget.ID),
			zap.String("statusMessage", statusMessage))
	}
}

func postAgentStatus(w http.ResponseWriter, r *http.Request) {
	/*ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	deploymentTarget := internalctx.GetDeploymentTarget(ctx)
	// TODO get latest deployment
	if err := db.CreateDeploymentStatus(ctx, deploymentTarget, statusMessage); err != nil {
		log.Error("failed to create deployment target status", zap.Error(err),
			zap.String("deploymentTargetId", deploymentTarget.ID),
			zap.String("statusMessage", statusMessage))
	}*/
}

func queryAuthDeploymentTargetCtxMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		accessKeyId := r.URL.Query().Get("accessKeyId")
		accessKeySecret := r.URL.Query().Get("accessKeySecret")

		if deploymentTarget, err := getVerifiedDeploymentTarget(ctx, accessKeyId, accessKeySecret); err != nil {
			log.Error("failed to get deployment target from query auth", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			ctx = internalctx.WithDeploymentTarget(ctx, deploymentTarget)
			// TODO set current org ID into context -> no db "Unauthenticated" stuff necessary
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func basicAuthDeploymentTargetCtxMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		if accessKeyId, accessKeySecret, ok := r.BasicAuth(); !ok {
			log.Error("invalid Basic Auth")
			w.WriteHeader(http.StatusUnauthorized)
		} else if deploymentTarget, err := getVerifiedDeploymentTarget(ctx, accessKeyId, accessKeySecret); err != nil {
			log.Error("failed to get deployment target from query auth", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			ctx = internalctx.WithDeploymentTarget(ctx, deploymentTarget)
			// TODO set current org ID into context -> no db "Unauthenticated" stuff necessary
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func getVerifiedDeploymentTarget(ctx context.Context, accessKeyId string, accessKeySecret string) (*types.DeploymentTarget, error) {
	if deploymentTarget, err := db.GetDeploymentTargetUnauthenticated(ctx, accessKeyId); err != nil {
		return nil, fmt.Errorf("failed to get deployment target from DB: %w", err)
	} else if deploymentTarget.AccessKeySalt == nil || deploymentTarget.AccessKeyHash == nil {
		return nil, errors.New("deployment target does not have key and salt")
	} else if err := security.VerifyAccessKey(*deploymentTarget.AccessKeySalt, *deploymentTarget.AccessKeyHash, accessKeySecret); err != nil {
		return nil, fmt.Errorf("failed to verify access: %w", err)
	} else {
		return deploymentTarget, nil
	}
}

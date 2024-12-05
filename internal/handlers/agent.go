package handlers

import (
	"encoding/base64"
	"errors"
	"github.com/glasskube/cloud/internal/apierrors"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/env"
	"github.com/glasskube/cloud/internal/resources"
	"github.com/glasskube/cloud/internal/security"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strings"
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
		log.Warn("connect endpoint called without accessKeyId or accessKeySecret", zap.String("accessKeyId", accessKeyId))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	deploymentTarget, err := db.GetDeploymentTargetUnauthenticated(ctx, accessKeyId)
	if errors.Is(err, apierrors.ErrNotFound) {
		log.Warn("connect failed: deployment target not found", zap.String("accessKeyId", accessKeyId))
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Error("failed to get DeploymentTarget", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if deploymentTarget.CurrentStatus != nil {
		log.Warn("deployment target has already been connected")
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else if deploymentTarget.AccessKeySalt == nil || deploymentTarget.AccessKeyHash == nil {
		log.Warn("deployment target does not have access key salt and hash configured", zap.String("deploymentTargetId", deploymentTarget.ID))
		w.WriteHeader(http.StatusUnauthorized)
	} else if err := security.VerifyAccessKey(*deploymentTarget.AccessKeySalt, *deploymentTarget.AccessKeyHash, accessKeySecret); err != nil {
		log.Error("failed to verify access key secret", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
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
			"accessKeyId":     accessKeyId,
			"accessKeySecret": accessKeySecret,
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
		u = u.JoinPath("/api/resources")
		return u.String(), nil
	}
}

func downloadResources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	authHeader := r.Header.Get("Authorization")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Basic" {
		log.Warn("received download request without Basic Authorization header", zap.String("authHeader", authHeader))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if authDecoded, err := base64.StdEncoding.DecodeString(parts[1]); err != nil {
		log.Error("failed to decode auth string", zap.Error(err))
		w.WriteHeader(500)
		return
	} else {
		authParts := strings.Split(string(authDecoded), ":")
		if len(authParts) != 2 {
			log.Warn("received download request with invalid Basic Authorization header")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		accessKeyId := authParts[0]
		accessKeySecret := authParts[1]

		if deploymentTarget, err := db.GetDeploymentTargetUnauthenticated(ctx, accessKeyId); err != nil {
			log.Error("failed to get deployment target", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			if err := security.VerifyAccessKey(*deploymentTarget.AccessKeySalt, *deploymentTarget.AccessKeyHash, accessKeySecret); err != nil {
				log.Error("failed to verify access key secret", zap.Error(err))
				w.WriteHeader(http.StatusUnauthorized)
			} else if composeFileData, err := db.GetLatestDeploymentComposeFileUnauthenticated(ctx, deploymentTarget.ID); err != nil && !errors.Is(err, apierrors.ErrNotFound) {
				log.Error("failed to get compose file from DB", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Add("Content-Type", "application/yaml")
				if _, err := w.Write(composeFileData); err != nil {
					log.Error("failed to write compose file", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
			// TODO should probably also write success/error into the status?
			if err := db.CreateDeploymentTargetStatus(ctx, deploymentTarget, "lol"); err != nil {
				log.Error("failed to create deployment target status", zap.Error(err), zap.String("deploymentTargetId", deploymentTarget.ID))
			}
		}
	}
}

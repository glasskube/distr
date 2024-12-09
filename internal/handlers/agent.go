package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"text/template"

	"github.com/glasskube/cloud/internal/types"

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
		r.Post("/agent/status", postAgentStatus)
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
		} else if resourcesEndpoint, statusEndpoint, err := buildEndpoints(); err != nil {
			log.Error("failed to build resources url", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		} else if err := tmpl.Execute(w, map[string]string{
			"resourcesEndpoint": resourcesEndpoint,
			"statusEndpoint":    statusEndpoint,
			"targetId":          r.URL.Query().Get("targetId"),
			"targetSecret":      r.URL.Query().Get("targetSecret"),
		}); err != nil {
			log.Error("failed to execute yaml template", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func buildEndpoints() (string, string, error) {
	if u, err := url.Parse(env.Host()); err != nil {
		return "", "", err
	} else {
		u = u.JoinPath("/api/v1/agent")
		return u.JoinPath("resources").String(), u.JoinPath("status").String(), nil
	}
}

func downloadResources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	deploymentTarget := internalctx.GetDeploymentTarget(ctx)
	var statusMessage string
	orgId := internalctx.GetOrgId(ctx)
	if deploymentId, composeFileData, err := db.GetLatestDeploymentComposeFile(
		ctx, deploymentTarget.ID, orgId); err != nil && !errors.Is(err, apierrors.ErrNotFound) {
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
		w.Header().Add("X-Resource-Correlation-ID", deploymentId)
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
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	correlationID := r.Header.Get("X-Resource-Correlation-ID")
	if correlationID == "" {
		log.Info("received status without correlation ID")
		w.WriteHeader(http.StatusBadRequest)
	} else {
		deploymentTarget := internalctx.GetDeploymentTarget(ctx)
		if body, err := io.ReadAll(r.Body); err != nil {
			log.Error("failed to read status body", zap.Error(err))
		} else if err := db.CreateDeploymentStatus(ctx, correlationID, string(body)); err != nil {
			log.Error("failed to create deployment target status", zap.Error(err),
				zap.String("correlationID", correlationID),
				zap.String("deploymentTargetId", deploymentTarget.ID),
				zap.String("statusMessage", string(body)))
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}

func queryAuthDeploymentTargetCtxMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		targetId := r.URL.Query().Get("targetId")
		targetSecret := r.URL.Query().Get("targetSecret")

		if deploymentTarget, err := getVerifiedDeploymentTarget(ctx, targetId, targetSecret); err != nil {
			log.Error("failed to get deployment target from query auth", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			ctx = internalctx.WithDeploymentTarget(ctx, deploymentTarget)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func basicAuthDeploymentTargetCtxMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		if targetId, targetSecret, ok := r.BasicAuth(); !ok {
			log.Error("invalid Basic Auth")
			w.WriteHeader(http.StatusUnauthorized)
		} else if deploymentTarget, err := getVerifiedDeploymentTarget(ctx, targetId, targetSecret); err != nil {
			log.Error("failed to get deployment target from query auth", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			ctx = internalctx.WithDeploymentTarget(ctx, deploymentTarget)
			ctx = internalctx.WithOrgId(ctx, deploymentTarget.OrganizationID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func getVerifiedDeploymentTarget(
	ctx context.Context,
	targetId string,
	targetSecret string,
) (*types.DeploymentTarget, error) {
	if deploymentTarget, err := db.GetDeploymentTarget(ctx, targetId, nil); err != nil {
		return nil, fmt.Errorf("failed to get deployment target from DB: %w", err)
	} else if deploymentTarget.AccessKeySalt == nil || deploymentTarget.AccessKeyHash == nil {
		return nil, errors.New("deployment target does not have key and salt")
	} else if err := security.VerifyAccessKey(
		*deploymentTarget.AccessKeySalt, *deploymentTarget.AccessKeyHash, targetSecret); err != nil {
		return nil, fmt.Errorf("failed to verify access: %w", err)
	} else {
		return deploymentTarget, nil
	}
}

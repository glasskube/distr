package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"text/template"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/cloud/api"
	"github.com/glasskube/cloud/internal/auth"
	"github.com/glasskube/cloud/internal/middleware"
	"github.com/go-chi/jwtauth/v5"

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
	r.Route("/agent", func(r chi.Router) {
		// agent login (from basic auth to token)
		r.Post("/login", agentLogin)

		r.Group(func(r chi.Router) {
			// agent routes, authenticated via token
			r.Use(jwtauth.Verifier(auth.JWTAuth))
			r.Use(jwtauth.Authenticator(auth.JWTAuth))
			r.Use(middleware.SentryUser)
			r.Use(agentAuthDeploymentTargetCtxMiddleware)
			r.Get("/resources", downloadResources)
			r.Post("/status", postAgentStatus)
		})
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
			sentry.GetHubFromContext(ctx).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else if tmpl, err := template.New("agent").Parse(string(yamlTemplate)); err != nil {
			log.Error("failed to get parse yaml template", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else if loginEndpoint, resourcesEndpoint, statusEndpoint, err := buildEndpoints(); err != nil {
			log.Error("failed to build resources url", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else if err := tmpl.Execute(w, map[string]string{
			"loginEndpoint":     loginEndpoint,
			"resourcesEndpoint": resourcesEndpoint,
			"statusEndpoint":    statusEndpoint,
			"targetId":          r.URL.Query().Get("targetId"),
			"targetSecret":      r.URL.Query().Get("targetSecret"),
			"agentInterval":     env.AgentInterval().String(),
		}); err != nil {
			log.Error("failed to execute yaml template", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func buildEndpoints() (string, string, string, error) {
	if u, err := url.Parse(env.Host()); err != nil {
		return "", "", "", err
	} else {
		u = u.JoinPath("/api/v1/agent")
		return u.JoinPath("login").String(), u.JoinPath("resources").String(), u.JoinPath("status").String(), nil
	}
}

func agentLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)

	if targetId, targetSecret, ok := r.BasicAuth(); !ok {
		log.Error("invalid Basic Auth")
		w.WriteHeader(http.StatusUnauthorized)
	} else if deploymentTarget, err := getVerifiedDeploymentTarget(ctx, targetId, targetSecret); err != nil {
		log.Error("failed to get deployment target from query auth", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		// TODO maybe even randomize token valid duration
		if _, token, err := auth.GenerateAgentTokenValidFor(
			deploymentTarget.ID, deploymentTarget.OrganizationID, env.AgentTokenMaxValidDuration()); err != nil {
			log.Error("failed to create agent token", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			_ = json.NewEncoder(w).Encode(api.AuthLoginResponse{Token: token})
		}
	}
}

func downloadResources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	deploymentTarget := internalctx.GetDeploymentTarget(ctx)
	var statusMessage string
	if orgId, err := auth.CurrentOrgId(ctx); err != nil {
		msg := "failed to get compose file from DB"
		statusMessage = fmt.Sprintf("%v: %v", msg, err.Error())
		log.Error(msg, zap.Error(err), zap.String("deploymentTargetId", deploymentTarget.ID))
		w.WriteHeader(http.StatusInternalServerError)
	} else if deploymentId, composeFileData, err := db.GetLatestDeploymentComposeFile(
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

	// not in a TX because insertion should not be rolled back when the cleanup fails
	if err := db.CreateDeploymentTargetStatus(ctx, deploymentTarget, statusMessage); err != nil {
		log.Error("failed to create deployment target status – skipping cleanup of old statuses", zap.Error(err),
			zap.String("deploymentTargetId", deploymentTarget.ID),
			zap.String("statusMessage", statusMessage))
	} else if cnt, err := db.CleanupDeploymentTargetStatus(ctx, deploymentTarget, env.StatusEntriesMaxAge()); err != nil {
		log.Error("failed to cleanup old deployment target status", zap.Error(err),
			zap.String("deploymentTargetId", deploymentTarget.ID),
			zap.String("statusMessage", statusMessage),
			zap.Duration("maxAge", env.StatusEntriesMaxAge()))
	} else if cnt > 0 {
		log.Debug("cleaned up old statuses of deployment target",
			zap.String("deploymentTargetId", deploymentTarget.ID),
			zap.Int64("count", cnt),
			zap.Duration("maxAge", env.StatusEntriesMaxAge()))
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
		} else {
			// not in a TX because insertion should not be rolled back when the cleanup fails
			if err := db.CreateDeploymentStatus(ctx, correlationID, string(body)); err != nil {
				log.Error("failed to create deployment target status – skipping cleanup of old statuses", zap.Error(err),
					zap.String("deploymentId", correlationID),
					zap.String("deploymentTargetId", deploymentTarget.ID),
					zap.String("statusMessage", string(body)))
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				w.WriteHeader(http.StatusOK)
			}

			if cnt, err := db.CleanupDeploymentStatus(ctx, correlationID, env.StatusEntriesMaxAge()); err != nil {
				log.Error("failed to cleanup old deployment status", zap.Error(err),
					zap.String("deploymentId", correlationID),
					zap.String("statusMessage", string(body)),
					zap.Duration("maxAge", env.StatusEntriesMaxAge()))
			} else if cnt > 0 {
				log.Debug("cleaned up old statuses of deployment",
					zap.String("deploymentId", correlationID),
					zap.Int64("count", cnt),
					zap.Duration("maxAge", env.StatusEntriesMaxAge()))
			}
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

func agentAuthDeploymentTargetCtxMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if orgId, err := auth.CurrentOrgId(ctx); err != nil {
			internalctx.GetLogger(r.Context()).Error("failed to get orgId from token", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		} else if targetId, err := auth.CurrentSubject(ctx); err != nil {
			internalctx.GetLogger(r.Context()).Error("failed to get subject from token", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		} else if deploymentTarget, err :=
			db.GetDeploymentTarget(ctx, targetId, &orgId); errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else if err != nil {
			internalctx.GetLogger(r.Context()).Error("failed to get DeploymentTarget", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			ctx = internalctx.WithDeploymentTarget(ctx, &deploymentTarget.DeploymentTarget)
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
		return &deploymentTarget.DeploymentTarget, nil
	}
}

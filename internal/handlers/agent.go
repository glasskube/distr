package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/cloud/api"
	"github.com/glasskube/cloud/internal/apierrors"
	"github.com/glasskube/cloud/internal/auth"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/env"
	"github.com/glasskube/cloud/internal/middleware"
	"github.com/glasskube/cloud/internal/resources"
	"github.com/glasskube/cloud/internal/security"
	"github.com/glasskube/cloud/internal/types"
	"github.com/glasskube/cloud/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"
)

func AgentRouter(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(queryAuthDeploymentTargetCtxMiddleware)
		r.Get("/connect", connectHandler())
	})
	r.Route("/agent", func(r chi.Router) {
		// agent login (from basic auth to token)
		r.Post("/login", agentLoginHandler)

		r.Group(func(r chi.Router) {
			// agent routes, authenticated via token
			r.Use(jwtauth.Verifier(auth.JWTAuth))
			r.Use(jwtauth.Authenticator(auth.JWTAuth))
			r.Use(middleware.SentryUser)
			r.Use(agentAuthDeploymentTargetCtxMiddleware)
			r.Use(rateLimitPerAgent)
			r.Get("/resources", agentResourcesHandler)
			r.Post("/status", angentPostStatusHandler)
		})
	})
}

func connectHandler() http.HandlerFunc {
	dockerTempl := util.Require(resources.GetTemplate("embedded/agent/docker/docker-compose.yaml"))
	kubernetesTempl := util.Require(resources.GetTemplate("embedded/agent/kubernetes/manifest.yaml"))
	loginEndpoint, resourcesEndpoint, statusEndpoint, err := buildEndpoints()
	util.Must(err)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		deploymentTarget := internalctx.GetDeploymentTarget(ctx)
		templateData := map[string]any{
			"loginEndpoint":     loginEndpoint,
			"resourcesEndpoint": resourcesEndpoint,
			"statusEndpoint":    statusEndpoint,
			"targetId":          r.URL.Query().Get("targetId"),
			"targetSecret":      r.URL.Query().Get("targetSecret"),
			"agentInterval":     env.AgentInterval(),
		}
		if deploymentTarget.CurrentStatus != nil {
			log.Warn("deployment target has already been connected")
			http.Error(w, "deployment target has already been connected", http.StatusBadRequest)
		} else if deploymentTarget.Type == types.DeploymentTypeDocker {
			w.Header().Add("Content-Type", "application/yaml")
			if err := dockerTempl.Execute(w, templateData); err != nil {
				log.Error("failed to execute yaml template", zap.Error(err))
				sentry.GetHubFromContext(ctx).CaptureException(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.Header().Add("Content-Type", "application/yaml")
			if err := kubernetesTempl.Execute(w, templateData); err != nil {
				log.Error("failed to execute yaml template", zap.Error(err))
				sentry.GetHubFromContext(ctx).CaptureException(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
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

func agentLoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)

	if targetId, targetSecret, ok := r.BasicAuth(); !ok {
		log.Error("invalid Basic Auth")
		w.WriteHeader(http.StatusUnauthorized)
	} else if agentLoginPerTargetIdRateLimiter.RespondOnLimit(w, r, targetId) {
		return
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

func agentResourcesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	deploymentTarget := internalctx.GetDeploymentTarget(ctx)
	log := internalctx.GetLogger(ctx).With(zap.String("deploymentTargetId", deploymentTarget.ID))

	var statusMessage string
	var appVersion *types.ApplicationVersion
	deployment, err := db.GetLatestDeploymentForDeploymentTarget(ctx, deploymentTarget.ID)
	if errors.Is(err, apierrors.ErrNotFound) {
		log.Info("latest deployment not found", zap.Error(err))
		statusMessage = "EMPTY"
	} else if err != nil {
		msg := "failed to get latest Deployment from DB"
		log.Error(msg, zap.Error(err))
		statusMessage = fmt.Sprintf("%v: %v", msg, err)
		w.WriteHeader(http.StatusInternalServerError)
	} else if av, err := db.GetApplicationVersion(ctx, deployment.ApplicationVersionId); err != nil {
		msg := "failed to get ApplicationVersion from DB"
		log.Error(msg, zap.Error(err))
		statusMessage = fmt.Sprintf("%v: %v", msg, err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		statusMessage = "OK"
		appVersion = av
	}

	if deploymentTarget.Type == types.DeploymentTypeDocker {
		if deployment != nil && appVersion != nil {
			w.Header().Add("Content-Type", "application/yaml")
			w.Header().Add("X-Resource-Correlation-ID", deployment.ID)
			if _, err := w.Write(appVersion.ComposeFileData); err != nil {
				msg := "failed to write compose file"
				statusMessage = fmt.Sprintf("%v: %v", msg, err)
				log.Error(msg, zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			// it the status wasn't previously set to something else send a 204 code
			w.WriteHeader(http.StatusNoContent)
		}
	} else {
		respose := api.KubernetesAgentResource{
			Namespace: *deploymentTarget.Namespace,
		}
		if deployment != nil && appVersion != nil {
			// TODO: parse values yaml and merge
			w.Header().Add("X-Resource-Correlation-ID", deployment.ID)
			respose.Deployment = &api.KubernetesAgentDeployment{
				RevisionID:   deployment.ID, // TODO: Update to use DeploymentRevision.ID once implemented
				ReleaseName:  *deployment.ReleaseName,
				ChartUrl:     *appVersion.ChartUrl,
				ChartVersion: *appVersion.ChartVersion,
			}
			if *appVersion.ChartType == types.HelmChartTypeRepository {
				respose.Deployment.ChartName = *appVersion.ChartName
			}
		}
		RespondJSON(w, respose)
	}

	// not in a TX because insertion should not be rolled back when the cleanup fails
	if err := db.CreateDeploymentTargetStatus(ctx, deploymentTarget, statusMessage); err != nil {
		log.Error("failed to create deployment target status – skipping cleanup of old statuses", zap.Error(err),
			zap.String("deploymentTargetId", deploymentTarget.ID),
			zap.String("statusMessage", statusMessage))
	} else if cnt, err := db.CleanupDeploymentTargetStatus(ctx, deploymentTarget); err != nil {
		log.Error("failed to cleanup old deployment target status", zap.Error(err),
			zap.String("deploymentTargetId", deploymentTarget.ID))
	} else if cnt > 0 {
		log.Debug("old deployment target statuses deleted",
			zap.String("deploymentTargetId", deploymentTarget.ID),
			zap.Int64("count", cnt),
			zap.Duration("maxAge", *env.StatusEntriesMaxAge()))
	}
}

func angentPostStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	correlationID := r.Header.Get("X-Resource-Correlation-ID")
	if correlationID == "" {
		log.Info("received status without correlation ID")
		w.WriteHeader(http.StatusBadRequest)
	} else {
		deploymentTarget := internalctx.GetDeploymentTarget(ctx)
		if status, err := JsonBody[api.AgentDeploymentStatus](w, r); err != nil {
			return
		} else {
			if err := db.CreateDeploymentStatus(ctx, correlationID, status.Type, status.Message); err != nil {
				log.Error("failed to create deployment target status – skipping cleanup of old statuses", zap.Error(err),
					zap.String("deploymentId", correlationID),
					zap.String("deploymentTargetId", deploymentTarget.ID),
					zap.String("statusType", string(status.Type)),
					zap.String("statusMessage", status.Message))
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				w.WriteHeader(http.StatusOK)
			}

			// not in a TX because insertion should not be rolled back when the cleanup fails
			if cnt, err := db.CleanupDeploymentStatus(ctx, correlationID); err != nil {
				log.Error("failed to cleanup old deployment status", zap.Error(err), zap.String("deploymentId", correlationID))
			} else if cnt > 0 {
				log.Debug("old deployment statuses deleted",
					zap.String("deploymentId", correlationID),
					zap.Int64("count", cnt),
					zap.Duration("maxAge", *env.StatusEntriesMaxAge()))
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

		if agentConnectPerTargetIdRateLimiter.RespondOnLimit(w, r, targetId) {
			return
		} else if deploymentTarget, err := getVerifiedDeploymentTarget(ctx, targetId, targetSecret); err != nil {
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

var agentConnectPerTargetIdRateLimiter = httprate.NewRateLimiter(5, time.Minute)
var agentLoginPerTargetIdRateLimiter = httprate.NewRateLimiter(5, time.Minute)

var rateLimitPerAgent = httprate.Limit(
	2*15, // as long as we have 5 sec interval: 12 resources, 12 status requests
	1*time.Minute,
	httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
		return auth.CurrentUserId(r.Context())
	}),
)

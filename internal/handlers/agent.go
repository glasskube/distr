package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/cloud/api"
	"github.com/glasskube/cloud/internal/agentclient/useragent"
	"github.com/glasskube/cloud/internal/agentmanifest"
	"github.com/glasskube/cloud/internal/apierrors"
	"github.com/glasskube/cloud/internal/authjwt"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/env"
	"github.com/glasskube/cloud/internal/middleware"
	"github.com/glasskube/cloud/internal/security"
	"github.com/glasskube/cloud/internal/types"
	"github.com/glasskube/cloud/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"
)

func AgentRouter(r chi.Router) {
	r.With(
		queryAuthDeploymentTargetCtxMiddleware,
	).Group(func(r chi.Router) {
		r.Get("/connect", connectHandler())
	})
	r.Route("/agent", func(r chi.Router) {
		// agent login (from basic auth to token)
		r.Post("/login", agentLoginHandler)

		r.With(
			jwtauth.Verifier(authjwt.JWTAuth),
			jwtauth.Authenticator(authjwt.JWTAuth),
			middleware.SentryUser,
			agentAuthDeploymentTargetCtxMiddleware,
			rateLimitPerAgent,
		).Group(func(r chi.Router) {
			// agent routes, authenticated via token
			r.Get("/manifest", agentManifestHandler())
			r.Get("/resources", agentResourcesHandler)
			r.Post("/status", angentPostStatusHandler)
		})
	})
}

func connectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		deploymentTarget := internalctx.GetDeploymentTarget(ctx)

		if deploymentTarget.CurrentStatus != nil &&
			deploymentTarget.CurrentStatus.CreatedAt.Add(2*env.AgentInterval()).After(time.Now()) {
			http.Error(
				w,
				fmt.Sprintf(
					"deployment target is already connected and appears to be still running (last status %v)",
					deploymentTarget.CurrentStatus.CreatedAt),
				http.StatusBadRequest,
			)
			return
		}

		secret := r.URL.Query().Get("targetSecret")
		if manifest, err := agentmanifest.Get(ctx, *deploymentTarget, &secret); err != nil {
			log.Error("could not get agent manifest", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Header().Add("Content-Type", "application/yaml")
			if _, err := io.Copy(w, manifest); err != nil {
				log.Warn("writing to client failed", zap.Error(err))
			}
		}
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
		if _, token, err := authjwt.GenerateAgentTokenValidFor(
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

	// TODO: Consider consolidating all types into the same response format
	if deploymentTarget.Type == types.DeploymentTypeDocker {
		if deployment != nil && appVersion != nil {
			response := api.DockerAgentResource{
				AgentResource: api.AgentResource{RevisionID: deployment.DeploymentRevisionID},
				ComposeFile:   appVersion.ComposeFileData,
			}
			RespondJSON(w, response)
		} else {
			// it the status wasn't previously set to something else send a 204 code
			w.WriteHeader(http.StatusNoContent)
		}
	} else {
		response := api.KubernetesAgentResource{
			Namespace: *deploymentTarget.Namespace,
			Version:   deploymentTarget.AgentVersion,
		}
		if deployment != nil && appVersion != nil {
			response.AgentResource = api.AgentResource{RevisionID: deployment.DeploymentRevisionID}
			response.Deployment = &api.KubernetesAgentDeployment{
				RevisionID:   deployment.DeploymentRevisionID,
				ReleaseName:  *deployment.ReleaseName,
				ChartUrl:     *appVersion.ChartUrl,
				ChartVersion: *appVersion.ChartVersion,
			}
			if versionValues, err := appVersion.ParsedValuesFile(); err != nil {
				log.Warn("parse error", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			} else if deploymentValues, err := deployment.ParsedValuesFile(); err != nil {
				log.Warn("parse error", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			} else if merged, err := util.MergeAllRecursive(versionValues, deploymentValues); err != nil {
				log.Warn("merge error", zap.Error(err))
				http.Error(w, fmt.Sprintf("error merging values files: %v", err), http.StatusInternalServerError)
				return
			} else {
				response.Deployment.Values = merged
			}
			if *appVersion.ChartType == types.HelmChartTypeRepository {
				response.Deployment.ChartName = *appVersion.ChartName
			}
		}
		RespondJSON(w, response)
	}

	// not in a TX because insertion should not be rolled back when the cleanup fails
	if err := db.CreateDeploymentTargetStatus(ctx, &deploymentTarget.DeploymentTarget, statusMessage); err != nil {
		log.Error("failed to create deployment target status – skipping cleanup of old statuses", zap.Error(err),
			zap.String("deploymentTargetId", deploymentTarget.ID),
			zap.String("statusMessage", statusMessage))
	} else if cnt, err := db.CleanupDeploymentTargetStatus(ctx, &deploymentTarget.DeploymentTarget); err != nil {
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

	status, err := JsonBody[api.AgentDeploymentStatus](w, r)
	if err != nil {
		return
	}
	if err := db.CreateDeploymentRevisionStatus(ctx, status.RevisionID, status.Type, status.Message); err != nil {
		log.Error("failed to create deployment revision status – skipping cleanup of old statuses", zap.Error(err),
			zap.Reflect("status", status))
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusOK)
	}

	// not in a TX because insertion should not be rolled back when the cleanup fails
	if cnt, err := db.CleanupDeploymentRevisionStatus(ctx, status.RevisionID); err != nil {
		log.Error("failed to cleanup old deployment revision status", zap.Error(err), zap.Reflect("status", status))
	} else if cnt > 0 {
		log.Debug("old deployment revision statuses deleted",
			zap.String("deploymentRevisionId", status.RevisionID),
			zap.Int64("count", cnt),
			zap.Duration("maxAge", *env.StatusEntriesMaxAge()))
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

func agentManifestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		deploymentTarget := internalctx.GetDeploymentTarget(ctx)
		log := internalctx.GetLogger(ctx).With(zap.String("deploymentTargetId", deploymentTarget.ID))

		if manifest, err := agentmanifest.Get(ctx, *deploymentTarget, nil); err != nil {
			log.Error("could not get agent manifest", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Header().Add("Content-Type", "application/yaml")
			if _, err := io.Copy(w, manifest); err != nil {
				log.Warn("writing to client failed", zap.Error(err))
			}
		}
	}
}

func agentAuthDeploymentTargetCtxMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		auth := middleware.Authn.Require(ctx)
		orgId := auth.CurrentOrgID()
		targetId := auth.CurrentUserID()

		if deploymentTarget, err :=
			db.GetDeploymentTarget(ctx, targetId, &orgId); errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else if err != nil {
			log.Error("failed to get DeploymentTarget", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			if ua := r.UserAgent(); strings.HasPrefix(ua, fmt.Sprintf("%v/", useragent.GlasskubeAgentUserAgent)) {
				reportedVersionName := strings.TrimPrefix(ua, fmt.Sprintf("%v/", useragent.GlasskubeAgentUserAgent))
				if reportedVersion, err := db.GetAgentVersionWithName(ctx, reportedVersionName); err != nil {
					log.Error("could not get reported agent version", zap.Error(err))
					sentry.GetHubFromContext(ctx).CaptureException(err)
				} else if deploymentTarget.ReportedAgentVersionID == nil ||
					reportedVersion.ID != *deploymentTarget.ReportedAgentVersionID {
					if err := db.UpdateDeploymentTargetReportedAgentVersionID(
						ctx, &deploymentTarget.DeploymentTarget, reportedVersion.ID); err != nil {
						log.Error("could not update reported agent version", zap.Error(err))
						sentry.GetHubFromContext(ctx).CaptureException(err)
					}
				}
			}
			ctx = internalctx.WithDeploymentTarget(ctx, deploymentTarget)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func getVerifiedDeploymentTarget(
	ctx context.Context,
	targetId string,
	targetSecret string,
) (*types.DeploymentTargetWithCreatedBy, error) {
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

var agentConnectPerTargetIdRateLimiter = httprate.NewRateLimiter(5, time.Minute)
var agentLoginPerTargetIdRateLimiter = httprate.NewRateLimiter(5, time.Minute)

var rateLimitPerAgent = httprate.Limit(
	2*15, // as long as we have 5 sec interval: 12 resources, 12 status requests
	1*time.Minute,
	httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
		if auth, err := middleware.Authn.Get(r.Context()); err != nil {
			return "", err
		} else {
			return auth.CurrentUserID(), nil
		}
	}),
)

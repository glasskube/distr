package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/cloud/api"
	"github.com/glasskube/cloud/internal/apierrors"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/env"
	"github.com/glasskube/cloud/internal/middleware"
	"github.com/glasskube/cloud/internal/security"
	"github.com/glasskube/cloud/internal/types"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func DeploymentTargetsRouter(r chi.Router) {
	r.Get("/", getDeploymentTargets)
	r.Post("/", createDeploymentTarget)
	r.Route("/{deploymentTargetId}", func(r chi.Router) {
		r.Use(deploymentTargetMiddleware)
		r.Get("/", getDeploymentTarget)
		r.Put("/", updateDeploymentTarget)
		r.Delete("/", deleteDeploymentTarget)
		r.Post("/access-request", createAccessForDeploymentTarget)
	})
}

func getDeploymentTargets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := middleware.Authn.Require(ctx)
	deploymentTargets, err := db.GetDeploymentTargets(
		ctx,
		auth.CurrentOrgID(),
		auth.CurrentUserID(),
		*auth.CurrentUserRole(),
	)
	if err != nil {
		internalctx.GetLogger(ctx).Error("failed to get DeploymentTargets", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		RespondJSON(w, deploymentTargets)
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
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := middleware.Authn.Require(ctx)
	if dt, err := JsonBody[types.DeploymentTargetWithCreatedBy](w, r); err != nil {
		return
	} else if err = dt.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if agentVersion, err := db.GetCurrentAgentVersion(ctx); err != nil {
		log.Warn("could not get current agent version", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		dt.AgentVersionID = agentVersion.ID
		if err = db.CreateDeploymentTarget(ctx, &dt, auth.CurrentOrgID(), auth.CurrentUserID()); err != nil {
			log.Warn("could not create DeploymentTarget", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			RespondJSON(w, dt)
		}
	}
}

func updateDeploymentTarget(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := middleware.Authn.Require(ctx)
	var dt types.DeploymentTargetWithCreatedBy
	if err := json.NewDecoder(r.Body).Decode(&dt); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, err)
		return
	}

	dt.AgentVersionID = dt.AgentVersion.ID

	existing := internalctx.GetDeploymentTarget(ctx)
	if dt.ID == "" {
		dt.ID = existing.ID
	} else if dt.ID != existing.ID {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "wrong id")
		return
	}

	if err := db.UpdateDeploymentTarget(ctx, &dt, auth.CurrentOrgID()); err != nil {
		log.Warn("could not update DeploymentTarget", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
	} else if err = json.NewEncoder(w).Encode(dt); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func deleteDeploymentTarget(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	dt := internalctx.GetDeploymentTarget(ctx)
	auth := middleware.Authn.Require(ctx)
	if dt.OrganizationID != auth.CurrentOrgID() {
		http.NotFound(w, r)
	} else if currentUser, err := db.GetUserAccountWithRole(
		ctx, auth.CurrentUserID(), auth.CurrentOrgID(),
	); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if currentUser.UserRole != types.UserRoleVendor && dt.CreatedByUserAccountID != currentUser.ID {
		http.Error(w, "must be vendor or creator", http.StatusForbidden)
	} else if err := db.DeleteDeploymentTargetWithID(ctx, dt.ID); err != nil {
		log.Warn("error deleting deployment target", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func createAccessForDeploymentTarget(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	deploymentTarget := internalctx.GetDeploymentTarget(ctx)
	auth := middleware.Authn.Require(ctx)

	if deploymentTarget.CurrentStatus != nil &&
		deploymentTarget.CurrentStatus.CreatedAt.Add(2*env.AgentInterval()).After(time.Now()) {
		http.Error(
			w,
			fmt.Sprintf(
				"access key cannot be regenerated because deployment target is already connected "+
					"and seems to be still running (last connection at %v)",
				deploymentTarget.CurrentStatus.CreatedAt,
			),
			http.StatusBadRequest,
		)
		return
	}

	var targetSecret string
	var err error
	if targetSecret, err = security.GenerateAccessKey(); err != nil {
		log.Error("failed to generate access key", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if salt, hash, err := security.HashAccessKey(targetSecret); err != nil {
		log.Error("failed to hash access key", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		deploymentTarget.AccessKeySalt = &salt
		deploymentTarget.AccessKeyHash = &hash
	}

	if err := db.UpdateDeploymentTargetAccess(ctx, &deploymentTarget.DeploymentTarget, auth.CurrentOrgID()); err != nil {
		log.Warn("could not update DeploymentTarget", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else if connectUrl, err := buildConnectUrl(deploymentTarget.ID, targetSecret); err != nil {
		log.Error("could not create connecturl", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else if err = json.NewEncoder(w).Encode(api.DeploymentTargetAccessTokenResponse{
		ConnectUrl:   connectUrl,
		TargetId:     deploymentTarget.ID,
		TargetSecret: targetSecret,
	}); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func buildConnectUrl(targetId string, targetSecret string) (string, error) {
	if u, err := url.Parse(env.Host()); err != nil {
		return "", err
	} else {
		query := url.Values{}
		query.Set("targetId", targetId)
		query.Set("targetSecret", targetSecret)
		u = u.JoinPath("/api/v1/connect")
		u.RawQuery = query.Encode()
		return u.String(), nil
	}
}

func deploymentTargetMiddleware(wh http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.PathValue("deploymentTargetId")
		auth := middleware.Authn.Require(ctx)
		orgId := auth.CurrentOrgID()
		if deploymentTarget, err := db.GetDeploymentTarget(ctx, id, &orgId); errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else if err != nil {
			internalctx.GetLogger(r.Context()).Error("failed to get DeploymentTarget", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			ctx = internalctx.WithDeploymentTarget(ctx, deploymentTarget)
			wh.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

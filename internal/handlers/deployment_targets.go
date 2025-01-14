package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/cloud/internal/auth"

	"github.com/glasskube/cloud/api"
	"github.com/glasskube/cloud/internal/apierrors"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/env"
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
		// r.Get("/latest-deployment", getLatestDeployment) // TODO not needed anymore?
		r.Put("/", updateDeploymentTarget)
		r.Delete("/", deleteDeploymentTarget)
		r.Post("/access-request", createAccessForDeploymentTarget)
	})
}

func getLatestDeployment(w http.ResponseWriter, r *http.Request) {
	dt := internalctx.GetDeploymentTarget(r.Context())
	if deployment, err := db.GetLatestDeploymentForDeploymentTarget(r.Context(), dt.ID); err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			internalctx.GetLogger(r.Context()).Error("failed to get latest deployment", zap.Error(err))
			sentry.GetHubFromContext(r.Context()).CaptureException(err)
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
		sentry.GetHubFromContext(r.Context()).CaptureException(err)
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
	if dt, err := JsonBody[types.DeploymentTargetWithCreatedBy](w, r); err != nil {
		return
	} else if err = dt.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if err = db.CreateDeploymentTarget(ctx, &dt); err != nil {
		log.Warn("could not create DeploymentTarget", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if err = json.NewEncoder(w).Encode(dt); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func updateDeploymentTarget(w http.ResponseWriter, r *http.Request) {
	log := internalctx.GetLogger(r.Context())
	var dt types.DeploymentTargetWithCreatedBy
	if err := json.NewDecoder(r.Body).Decode(&dt.DeploymentTarget); err != nil {
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
		sentry.GetHubFromContext(r.Context()).CaptureException(err)
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
	if orgId, err := auth.CurrentOrgId(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if dt.OrganizationID != orgId {
		http.NotFound(w, r)
	} else if currentUser, err := db.GetCurrentUserWithRole(ctx); err != nil {
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

	if deploymentTarget.CurrentStatus != nil {
		log.Warn("access key cannot be regenerated because deployment target has already been connected")
		w.WriteHeader(http.StatusBadRequest)
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

	if err := db.UpdateDeploymentTargetAccess(ctx, deploymentTarget); err != nil {
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
		if orgId, err := auth.CurrentOrgId(ctx); err != nil {
			internalctx.GetLogger(r.Context()).Error("failed to get orgId from token", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else if deploymentTarget, err := db.GetDeploymentTarget(ctx, id, &orgId); errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else if err != nil {
			internalctx.GetLogger(r.Context()).Error("failed to get DeploymentTarget", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			ctx = internalctx.WithDeploymentTarget(ctx, &deploymentTarget.DeploymentTarget)
			wh.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

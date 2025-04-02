package handlers

import (
	"errors"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func MetricsRouter(r chi.Router) {
	r.Use(middleware.RequireOrgAndRole, requireUserRoleVendor)
	r.Get("/uptime", getUptime)
}

func getUptime(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)
	log := internalctx.GetLogger(ctx)
	deploymentId, err := uuid.Parse(r.URL.Query().Get("deploymentId"))
	if err != nil {
		http.Error(w, "deploymentId is not a valid UUID", http.StatusBadRequest)
		return
	}

	if _, err := db.GetDeployment(ctx, deploymentId, auth.CurrentUserID(), *auth.CurrentOrgID(),
		*auth.CurrentUserRole()); errors.Is(err, apierrors.ErrNotFound) {
		http.Error(w, "deployment not found", http.StatusNotFound)
	} else if uptime, err := db.GetUptimeForDeployment(ctx, deploymentId); err != nil {
		log.Error("failed to get uptime metrics", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		RespondJSON(w, uptime)
	}
}

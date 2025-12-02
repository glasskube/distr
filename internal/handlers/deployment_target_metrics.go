package handlers

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func DeploymentTargetMetricsRouter(r chi.Router) {
	r.Use(middleware.RequireOrgAndRole)
	r.Get("/", getLatestDeploymentTargetMetrics)
}

func getLatestDeploymentTargetMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)

	if deploymentTargetMetrics, err := db.GetLatestDeploymentTargetMetrics(
		ctx,
		*auth.CurrentOrgID(),
		auth.CurrentCustomerOrgID(),
	); err != nil {
		internalctx.GetLogger(ctx).Error("failed to get deployment target metrics", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		RespondJSON(w, deploymentTargetMetrics)
	}
}

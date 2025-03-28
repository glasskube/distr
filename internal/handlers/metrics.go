package handlers

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func MetricsRouter(r chi.Router) {
	r.Get("/uptime", getUptime)
}

func getUptime(w http.ResponseWriter, r *http.Request) {
	// TODO org/user check etc
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	deploymentId, err := uuid.Parse(r.URL.Query().Get("deploymentId"))
	if err != nil {
		http.Error(w, "deploymentId is not a valid UUID", http.StatusBadRequest)
		return
	}

	if uptime, err := db.GetUptimeForDeployment(ctx, deploymentId); err != nil {
		log.Error("failed to get uptime metrics", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		RespondJSON(w, uptime)
	}
}

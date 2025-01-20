package handlers

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AgentVersionsRouter(r chi.Router) {
	r.Get("/", getAgentVersionsHandler())
}

func getAgentVersionsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		if agentVersions, err := db.GetAgentVersions(ctx); err != nil {
			log.Warn("could not get agent versions", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			RespondJSON(w, agentVersions)
		}
	}
}

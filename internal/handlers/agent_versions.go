package handlers

import (
	"net/http"

	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/db"
	"github.com/distr-sh/distr/internal/types"
	"github.com/getsentry/sentry-go"
	"github.com/oaswrap/spec/adapter/chiopenapi"
	"github.com/oaswrap/spec/option"
	"go.uber.org/zap"
)

func AgentVersionsRouter(r chiopenapi.Router) {
	r.WithOptions(option.GroupTags("Miscellaneous"))
	r.Get("/", getAgentVersionsHandler()).
		With(option.Description("List all agent versions")).
		With(option.Response(http.StatusOK, []types.AgentVersion{}))
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

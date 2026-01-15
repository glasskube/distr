package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/distr-sh/distr/internal/auth"
	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/db"
	"github.com/distr-sh/distr/internal/middleware"
	"github.com/distr-sh/distr/internal/types"
	"github.com/getsentry/sentry-go"
	"github.com/oaswrap/spec/adapter/chiopenapi"
	"github.com/oaswrap/spec/option"
	"go.uber.org/zap"
)

func ArtifactPullsRouter(r chiopenapi.Router) {
	r.WithOptions(option.GroupTags("Artifacts"))
	r.Use(
		middleware.RequireOrgAndRole,
		middleware.RequireVendor,
	)
	r.Get("/", getArtifactPullsHandler()).
		With(option.Description("List artifact version pulls")).
		With(option.Request(struct {
			Before *time.Time `query:"before"`
			Count  *int       `query:"count"`
		}{})).
		With(option.Response(http.StatusOK, []types.ArtifactVersionPull{}))
}

func getArtifactPullsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		auth := auth.Authentication.Require(ctx)
		before := time.Now()
		count := 50
		if s := r.FormValue("before"); s != "" {
			if t, err := time.Parse(time.RFC3339Nano, s); err != nil {
				http.Error(w, "before must be a date", http.StatusBadRequest)
				return
			} else {
				before = t
			}
		}
		if s := r.FormValue("count"); s != "" {
			if n, err := strconv.Atoi(s); err != nil {
				http.Error(w, "limit must be a date", http.StatusBadRequest)
				return
			} else {
				count = n
			}
		}
		pulls, err := db.GetArtifactVersionPulls(ctx, *auth.CurrentOrgID(), count, before)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			sentry.GetHubFromContext(ctx).CaptureException(err)
			log.Warn("could not get pulls", zap.Error(err))
			return
		}
		RespondJSON(w, pulls)
	}
}

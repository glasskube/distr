package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func ArtifactPullsRouter(r chi.Router) {
	r.Use(
		middleware.RequireOrgAndRole,
		middleware.RequireVendor,
	)
	r.Get("/", getArtifactPullsHandler())
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

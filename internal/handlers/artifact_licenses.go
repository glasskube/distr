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

func ArtifactLicensesRouter(r chi.Router) {
	r.Use(middleware.RequireOrgID, middleware.RequireUserRole, requireUserRoleVendor)
	r.Get("/", getArtifactLicenses)
}

func getArtifactLicenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)

	if licenses, err := db.GetArtifactLicenses(ctx, *auth.CurrentOrgID()); err != nil {
		log.Error("failed to get artifact licenses", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else {
		RespondJSON(w, licenses)
	}
}

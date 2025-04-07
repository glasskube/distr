package handlers

import (
	"errors"
	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

func TutorialsRouter(r chi.Router) {
	r.Use(middleware.RequireOrgAndRole, requireUserRoleVendor)
	r.Route("/{tutorial}", func(r chi.Router) {
		r.Get("/", getTutorialProgress)
		r.Put("/", saveTutorialProgress)
	})
}

func getTutorialProgress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)
	tutorial := r.PathValue("tutorial")
	if progress, err := db.GetTutorialProgress(ctx, auth.CurrentUserID(), types.Tutorial(tutorial)); errors.Is(err, apierrors.ErrNotFound) {
		http.NotFound(w, r)
	} else if err != nil {
		log.Warn("could not get tutorial progress", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		RespondJSON(w, progress)
	}
}

func saveTutorialProgress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)
	tutorial := r.PathValue("tutorial")

	progress, err := JsonBody[types.TutorialProgress](w, r)
	if err != nil {
		return
	}

	if progress.Tutorial != types.Tutorial(tutorial) {
		http.Error(w, "invalid tutorial type", http.StatusBadRequest)
		return
	}

	if err := db.SaveTutorialProgress(ctx, auth.CurrentUserID(), &progress); err != nil {
		log.Warn("could not save tutorial progress", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		RespondJSON(w, progress)
	}
}

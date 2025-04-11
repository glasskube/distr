package handlers

import (
	"errors"
	"net/http"

	"github.com/glasskube/distr/api"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func TutorialsRouter(r chi.Router) {
	r.Use(middleware.RequireOrgAndRole, requireUserRoleVendor)
	r.Get("/", getTutorialProgresses)
	r.Route("/{tutorial}", func(r chi.Router) {
		r.Get("/", getTutorialProgress)
		r.Put("/", saveTutorialProgress)
	})
}

func getTutorialProgresses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)
	if progresses, err := db.GetTutorialProgresses(ctx, auth.CurrentUserID()); err != nil {
		log.Warn("could not get tutorial progresses", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		RespondJSON(w, progresses)
	}
}

func getTutorialProgress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)
	tutorial := r.PathValue("tutorial")
	if progress, err := db.GetTutorialProgress(ctx, auth.CurrentUserID(),
		types.Tutorial(tutorial)); errors.Is(err, apierrors.ErrNotFound) {
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

	req, err := JsonBody[api.TutorialProgressRequest](w, r)
	if err != nil {
		return
	}

	if res, err := db.SaveTutorialProgress(ctx, auth.CurrentUserID(), types.Tutorial(tutorial), &req); err != nil {
		log.Warn("could not save tutorial progress", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		RespondJSON(w, res)
	}
}

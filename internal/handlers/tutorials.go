package handlers

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"github.com/glasskube/distr/internal/resources"

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
	tutorial := types.Tutorial(r.PathValue("tutorial"))

	req, err := JsonBody[api.TutorialProgressRequest](w, r)
	if err != nil {
		return
	}

	_ = db.RunTx(ctx, func(ctx context.Context) error {
		if tutorial == types.TutorialAgents && req.StepID == "welcome" && req.TaskID == "start" {
			var progress *types.TutorialProgress
			if progress, err = db.GetTutorialProgress(ctx, auth.CurrentUserID(), tutorial); err != nil {
				if !errors.Is(err, apierrors.ErrNotFound) {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return err
				}
			}
			if progress == nil || !slices.ContainsFunc(progress.Events, func(event types.TutorialProgressEvent) bool {
				return event.StepID == "welcome" && event.TaskID == "start"
			}) {
				if _, err := createHelloDistrApp(ctx); err != nil {
					return err
				}
			}
		}

		if res, err := db.SaveTutorialProgress(ctx, auth.CurrentUserID(), tutorial, &req); err != nil {
			log.Warn("could not save tutorial progress", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		} else {
			RespondJSON(w, res)
			return nil
		}
	})
}

func createHelloDistrApp(ctx context.Context) (*types.Application, error) {
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)

	application := types.Application{
		Name: "hello-distr",
		Type: types.DeploymentTypeDocker,
	}

	var composeFileData []byte
	var templateFileData []byte
	if composeFile, err := resources.Get("apps/hello-distr/docker-compose.yaml"); err != nil {
		log.Warn("failed to read hello-distr compose file", zap.Error(err))
	} else {
		composeFileData = composeFile
	}
	if templateFile, err := resources.Get("apps/hello-distr/template.env"); err != nil {
		log.Warn("failed to read hello-distr template file", zap.Error(err))
	} else {
		templateFileData = templateFile
	}

	version := types.ApplicationVersion{
		Name:             "0.1.9",
		ComposeFileData:  composeFileData,
		TemplateFileData: templateFileData,
	}

	if err := db.CreateApplication(ctx, &application, *auth.CurrentOrgID()); err != nil {
		return nil, err
	}
	version.ApplicationID = application.ID
	if err := db.CreateApplicationVersion(ctx, &version); err != nil {
		return nil, err
	}

	application.Versions = append(application.Versions, version)
	return &application, nil
}

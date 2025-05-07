package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"

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
			if dt, err := createSampleAppAndDeployment(ctx); err != nil {
				log.Warn("could not create sample app and deployment", zap.Error(err))
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, "could not create sample app and deployment", http.StatusInternalServerError)
				return err
			} else if dt != nil {
				// TODO save additional data?
				req.Value = map[string]string{
					"deploymentTargetId": dt.ID.String(),
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

func createSampleAppAndDeployment(ctx context.Context) (*types.DeploymentTargetWithCreatedBy, error) {
	if app, err := createHelloDistrApp(ctx); err != nil {
		return nil, fmt.Errorf("failed to create hello-distr app: %w", err)
	} else if dt, err := createHelloDistrDeploymentTarget(ctx); err != nil {
		return nil, fmt.Errorf("failed to create hello-distr deployment target: %w", err)
	} else if err := createHelloDistrDeploymentAndRevision(ctx, app.Versions[0].ID, dt.ID); err != nil {
		return nil, fmt.Errorf("failed to deploy hello-distr: %w", err)
	} else {
		return dt, nil
	}
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
		Name:             "0.1.10",
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

func createHelloDistrDeploymentTarget(ctx context.Context) (*types.DeploymentTargetWithCreatedBy, error) {
	auth := auth.Authentication.Require(ctx)
	dt := types.DeploymentTargetWithCreatedBy{
		DeploymentTarget: types.DeploymentTarget{
			Name: "hello-distr-tutorial",
			Type: types.DeploymentTypeDocker,
		},
	}
	if agentVersion, err := db.GetCurrentAgentVersion(ctx); err != nil {
		return nil, err
	} else {
		dt.AgentVersionID = &agentVersion.ID
		if err := db.CreateDeploymentTarget(ctx, &dt, *auth.CurrentOrgID(), auth.CurrentUserID()); err != nil {
			return nil, err
		} else {
			return &dt, nil
		}
	}
}

const helloDistrEnvironment = `
# mandatory values:
HELLO_DISTR_HOST=localhost
HELLO_DISTR_DB_NAME=hello-distr
HELLO_DISTR_DB_USER=distr
HELLO_DISTR_DB_PASSWORD=distr123
`

func createHelloDistrDeploymentAndRevision(ctx context.Context, appVersionID uuid.UUID, dtID uuid.UUID) error {
	deploymentRequest := &api.DeploymentRequest{
		ApplicationVersionID: appVersionID,
		DeploymentTargetID:   dtID,
		EnvFileData:          []byte(helloDistrEnvironment),
	}
	if err := db.CreateDeployment(ctx, deploymentRequest); err != nil {
		return err
	} else if _, err := db.CreateDeploymentRevision(ctx, deploymentRequest); err != nil {
		return err
	} else {
		return nil
	}
}

package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func DeploymentsRouter(r chi.Router) {
	r.Use(middleware.RequireOrgID, middleware.RequireUserRole)
	r.Put("/", putDeployment)
	r.Route("/{deploymentId}", func(r chi.Router) {
		r.Use(deploymentMiddleware)
		r.Get("/status", getDeploymentStatus)
	})
}

func putDeployment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)

	deploymentRequest, err := JsonBody[api.DeploymentRequest](w, r)
	if err != nil {
		return
	}

	_ = db.RunTx(ctx, pgx.TxOptions{}, func(ctx context.Context) error {
		if err := validateDeploymentRequest(ctx, w, deploymentRequest); err != nil {
			return err
		}

		if deploymentRequest.DeploymentID == nil {
			if err = db.CreateDeployment(ctx, &deploymentRequest); err != nil {
				log.Warn("could not create deployment", zap.Error(err))
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}
		}

		if _, err := db.CreateDeploymentRevision(ctx, &deploymentRequest); err != nil {
			log.Warn("could not create deployment revision", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		// TODO: We might need to send a proper deployment object back, but not sure yet what it looks like
		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}

func validateDeploymentRequest(
	ctx context.Context,
	w http.ResponseWriter,
	request api.DeploymentRequest,
) error {
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)
	orgId := *auth.CurrentOrgID()

	var license *types.ApplicationLicense
	var app *types.Application
	var version *types.ApplicationVersion
	var target *types.DeploymentTargetWithCreatedBy

	org, err := db.GetOrganizationByID(ctx, orgId)
	if err != nil {
		log.Error("failed to get org", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return err
	}

	if app, err =
		db.GetApplicationForApplicationVersionID(ctx, request.ApplicationVersionID, orgId); err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			return badRequestError(w, "Application does not exist")
		} else {
			log.Warn("could not get Application", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return err
		}
	}

	if version, err = db.GetApplicationVersion(ctx, request.ApplicationVersionID); err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			return badRequestError(w, "ApplicationVersion does not exist")
		} else {
			log.Warn("could not get ApplicationVersion", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return err
		}
	}

	if target, err = db.GetDeploymentTarget(ctx, request.DeploymentTargetID, &orgId); err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			return badRequestError(w, "DeploymentTarget does not exist")
		} else {
			log.Warn("could not get DeploymentTarget", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return err
		}
	}

	if target.Deployment != nil {
		if request.ApplicationLicenseID == nil {
			if target.Deployment.ApplicationLicenseID != nil {
				request.ApplicationLicenseID = target.Deployment.ApplicationLicenseID
			}
		} else if target.Deployment.ApplicationLicenseID == nil {
			return badRequestError(w, "can not update license")
		} else if *request.ApplicationLicenseID != *target.Deployment.ApplicationLicenseID {
			return badRequestError(w, "can not update license")
		}
	}

	if org.HasFeature(types.FeatureLicensing) {
		if request.ApplicationLicenseID != nil {
			if license, err = db.GetApplicationLicenseByID(ctx, *request.ApplicationLicenseID); err != nil {
				if errors.Is(err, apierrors.ErrNotFound) {
					return licenseNotFoundError(w)
				} else {
					log.Error("could not ApplicationLicense", zap.Error(err))
					sentry.GetHubFromContext(ctx).CaptureException(err)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return err
				}
			}

		} else if *auth.CurrentUserRole() == types.UserRoleCustomer {
			// license ID is required for customer but optional for vendor
			return badRequestError(w, "applicationLicenseId is required")
		}
	} else if request.ApplicationLicenseID != nil {
		return badRequestError(w, "unexpected applicationLicenseId")
	}

	if err = validateDeploymentRequestLicense(ctx, w, request, license, app, target); err != nil {
		return err
	} else if err = validateDeploymentRequestDeploymentType(w, target, app); err != nil {
		return err
	} else if err = validateDeploymentRequestDeploymentTarget(ctx, w, request, target); err != nil {
		return err
	} else if err = validateDeploymentRequestValues(w, request, version); err != nil {
		return err
	} else {
		return nil
	}
}

func badRequestError(w http.ResponseWriter, msg string) error {
	http.Error(w, msg, http.StatusBadRequest)
	return errors.New(msg)
}

func licenseNotFoundError(w http.ResponseWriter) error {
	return badRequestError(w, "license does not exist")
}

func invalidLicenseError(w http.ResponseWriter) error {
	return badRequestError(w, "invalid license")
}

func validateDeploymentRequestLicense(
	ctx context.Context,
	w http.ResponseWriter,
	request api.DeploymentRequest,
	license *types.ApplicationLicenseWithVersions,
	app *types.Application,
	target *types.DeploymentTargetWithCreatedBy,
) error {
	if license != nil {
		auth := auth.Authentication.Require(ctx)

		if license.OrganizationID != *auth.CurrentOrgID() {
			return licenseNotFoundError(w)
		}
		if license.OwnerUserAccountID == nil {
			return invalidLicenseError(w)
		}
		if *auth.CurrentUserRole() == types.UserRoleCustomer && *license.OwnerUserAccountID != auth.CurrentUserID() {
			return licenseNotFoundError(w)
		}
		if target.CreatedByUserAccountID != *license.OwnerUserAccountID {
			return invalidLicenseError(w)
		}
		if len(license.Versions) > 0 && !license.HasVersionWithID(request.ApplicationVersionID) {
			return invalidLicenseError(w)
		}
		if app.ID != license.ApplicationID {
			return invalidLicenseError(w)
		}
		if target.Deployment != nil && target.Deployment.ApplicationID != license.ApplicationID {
			return badRequestError(w, "license and deployment have applicationId mismatch")
		}
	}
	return nil
}

func validateDeploymentRequestDeploymentType(
	w http.ResponseWriter,
	target *types.DeploymentTargetWithCreatedBy,
	application *types.Application,
) error {
	if target.Type != application.Type {
		return badRequestError(w, "application and deployment target must have the same type")
	}
	return nil
}

func validateDeploymentRequestDeploymentTarget(
	ctx context.Context,
	w http.ResponseWriter,
	request api.DeploymentRequest,
	target *types.DeploymentTargetWithCreatedBy,
) error {
	auth := auth.Authentication.Require(ctx)

	if *auth.CurrentUserRole() == types.UserRoleCustomer &&
		target.CreatedByUserAccountID != auth.CurrentUserID() {
		http.Error(w, "DeploymentTarget not found", http.StatusBadRequest)
	}
	if request.DeploymentID == nil {
		if target.Deployment != nil {
			return badRequestError(w, "only one deployment per target is supported right now")
		}
	} else if target.Deployment == nil {
		return badRequestError(w, "given deployment is not a deployment of the given target")
	} else if target.Deployment.ID != *request.DeploymentID {
		return badRequestError(w, "given deployment does not match deployment of the given target")
	}
	return nil
}

func validateDeploymentRequestValues(
	w http.ResponseWriter,
	deploymentRequest api.DeploymentRequest,
	appVersion *types.ApplicationVersion,
) error {
	if deploymentValues, err := deploymentRequest.ParsedValuesFile(); err != nil {
		return badRequestError(w, fmt.Sprintf("invalid values: %v", err.Error()))
	} else if appVersionValues, err := appVersion.ParsedValuesFile(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	} else if _, err := util.MergeAllRecursive(appVersionValues, deploymentValues); err != nil {
		return badRequestError(w, fmt.Sprintf("values cannot be merged with base: %v", err))
	}
	return nil
}

func getDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	deployment := internalctx.GetDeployment(ctx)
	if deploymentStatus, err := db.GetDeploymentStatus(ctx, deployment.ID, 100); err != nil {
		internalctx.GetLogger(ctx).Error("failed to get deploymentstatus", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		RespondJSON(w, deploymentStatus)
	}
}

func deploymentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		deploymentId, err := uuid.Parse(r.PathValue("deploymentId"))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		deployment, err := db.GetDeployment(ctx, deploymentId)
		if errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else if err != nil {
			internalctx.GetLogger(r.Context()).Error("failed to get deployment", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			ctx = internalctx.WithDeployment(ctx, deployment)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

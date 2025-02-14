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
	deploymentRequest api.DeploymentRequest,
) error {
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)
	orgId := *auth.CurrentOrgID()

	var license *types.ApplicationLicenseWithVersions
	var application *types.Application
	var appVersion *types.ApplicationVersion
	var deploymentTarget *types.DeploymentTargetWithCreatedBy

	organization, err := db.GetOrganizationByID(ctx, orgId)
	if err != nil {
		log.Error("failed to get org", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return err
	}

	if organization.HasFeature(types.FeatureLicensing) {
		// license ID is required for customer but optional for vendor
		if deploymentRequest.ApplicationLicenseID == nil && *auth.CurrentUserRole() == types.UserRoleCustomer {
			http.Error(w, "applicationLicenseId is required", http.StatusBadRequest)
		} else {
			if license, err = db.GetApplicationLicenseWithID(ctx, *deploymentRequest.ApplicationLicenseID); err != nil {
				if errors.Is(err, apierrors.ErrNotFound) {
					return licenseNotFound(w)
				} else {
					log.Error("could not ApplicationLicense", zap.Error(err))
					sentry.GetHubFromContext(ctx).CaptureException(err)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return err
				}
			}
		}
	} else if deploymentRequest.ApplicationLicenseID != nil {
		http.Error(w, "unexpected applicationLicenseId", http.StatusBadRequest)
		return errors.New("unexpected applicationLicenseId")
	}

	if application, err =
		db.GetApplicationForApplicationVersionID(ctx, deploymentRequest.ApplicationVersionID, orgId); err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			http.Error(w, "Application does not exist", http.StatusBadRequest)
			return err
		} else {
			log.Warn("could not get Application", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return err
		}
	}

	if appVersion, err = db.GetApplicationVersion(ctx, deploymentRequest.ApplicationVersionID); err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			http.Error(w, "ApplicationVersion does not exist", http.StatusBadRequest)
			return err
		} else {
			log.Warn("could not get ApplicationVersion", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return err
		}
	}

	if deploymentTarget, err = db.GetDeploymentTarget(
		ctx, deploymentRequest.DeploymentTargetID, &orgId,
	); err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			http.Error(w, "DeploymentTarget does not exist", http.StatusBadRequest)
			return err
		} else {
			log.Warn("could not get DeploymentTarget", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return err
		}
	}

	if err := validateDeploymentRequestLicense(
		ctx, w, deploymentRequest, license, application, deploymentTarget); err != nil {
		return err
	} else if err := validateDeploymentRequestDeploymentType(w, deploymentTarget, application); err != nil {
		return err
	} else if err := validateDeploymentRequestDeploymentTarget(w, deploymentRequest, deploymentTarget); err != nil {
		return err
	} else if err := validateDeploymentRequestValues(w, deploymentRequest, appVersion); err != nil {
		return err
	} else {
		return nil
	}
}

func licenseNotFound(w http.ResponseWriter) error {
	errLicenseNotFound := errors.New("license does not exist")
	http.Error(w, errLicenseNotFound.Error(), http.StatusBadRequest)
	return errLicenseNotFound
}

func validateDeploymentRequestLicense(
	ctx context.Context,
	w http.ResponseWriter,
	deploymentRequest api.DeploymentRequest,
	license *types.ApplicationLicenseWithVersions,
	application *types.Application,
	deploymentTarget *types.DeploymentTargetWithCreatedBy,
) error {
	if license != nil {
		auth := auth.Authentication.Require(ctx)
		orgId := *auth.CurrentOrgID()

		if license.OrganizationID != orgId {
			return licenseNotFound(w)
		}
		if *auth.CurrentUserRole() == types.UserRoleCustomer &&
			(license.OwnerUserAccountID == nil || *license.OwnerUserAccountID != auth.CurrentUserID()) {
			return licenseNotFound(w)
		}
		if len(license.Versions) > 0 && !license.HasVersionWithID(deploymentRequest.ApplicationVersionID) {
			http.Error(w, "invalid license", http.StatusBadRequest)
			return errors.New("invalid license")
		}
		if application.ID != license.ApplicationID {
			http.Error(w, "invalid license", http.StatusBadRequest)
			return errors.New("invalid license")
		}
		if deploymentTarget.Deployment != nil && deploymentTarget.Deployment.ApplicationID != license.ApplicationID {
			msg := "given license does not have matching application ID for the deployment of the given target"
			http.Error(w, msg, http.StatusBadRequest)
			return errors.New(msg)
		}
	}
	return nil
}

func validateDeploymentRequestDeploymentType(
	w http.ResponseWriter,
	deploymentTarget *types.DeploymentTargetWithCreatedBy,
	application *types.Application,
) error {
	if deploymentTarget.Type != application.Type {
		msg := "application and deployment target must have the same type"
		http.Error(w, msg, http.StatusBadRequest)
		return errors.New(msg)
	}
	return nil
}

func validateDeploymentRequestDeploymentTarget(
	w http.ResponseWriter,
	deploymentRequest api.DeploymentRequest,
	deploymentTarget *types.DeploymentTargetWithCreatedBy,
) error {
	if deploymentRequest.DeploymentID == nil {
		if deploymentTarget.Deployment != nil {
			msg := "only one deployment per target is supported right now"
			http.Error(w, msg, http.StatusBadRequest)
			return errors.New(msg)
		}
	} else if deploymentTarget.Deployment == nil {
		msg := "given deployment is not a deployment of the given target"
		http.Error(w, msg, http.StatusBadRequest)
		return errors.New(msg)
	} else if deploymentTarget.Deployment.ID != *deploymentRequest.DeploymentID {
		msg := "given deployment does not match deployment of the given target"
		http.Error(w, msg, http.StatusBadRequest)
		return errors.New(msg)
	}
	return nil
}

func validateDeploymentRequestValues(
	w http.ResponseWriter,
	deploymentRequest api.DeploymentRequest,
	appVersion *types.ApplicationVersion,
) error {
	if deploymentValues, err := deploymentRequest.ParsedValuesFile(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	} else if appVersionValues, err := appVersion.ParsedValuesFile(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	} else if _, err := util.MergeAllRecursive(appVersionValues, deploymentValues); err != nil {
		http.Error(w, fmt.Sprintf("values cannot be merged with base: %v", err), http.StatusBadRequest)
		return err
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

package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/subscription"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func DeploymentsRouter(r chi.Router) {
	r.Use(middleware.RequireOrgAndRole)
	r.With(middleware.RequireReadWriteOrAdmin).Put("/", putDeployment)
	r.With(deploymentMiddleware).Route("/{deploymentId}", func(r chi.Router) {
		r.Get("/status", getDeploymentStatus)
		r.Get("/status/export", exportDeploymentStatusHandler())
		r.Get("/logs", getDeploymentLogsHandler())
		r.Get("/logs/resources", getDeploymentLogsResourcesHandler())
		r.Get("/logs/export", exportDeploymentLogsHandler())
		r.With(middleware.RequireReadWriteOrAdmin).Group(func(r chi.Router) {
			r.Patch("/", patchDeploymentHandler())
			r.Delete("/", deleteDeploymentHandler())
		})
	})
}

func putDeployment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)

	deploymentRequest, err := JsonBody[api.DeploymentRequest](w, r)
	if err != nil {
		return
	}

	_ = db.RunTx(ctx, func(ctx context.Context) error {
		if err := validateDeploymentRequest(ctx, w, deploymentRequest); err != nil {
			return err
		}

		if deploymentRequest.DeploymentID == nil {
			if err = db.CreateDeployment(ctx, &deploymentRequest); errors.Is(err, apierrors.ErrConflict) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return err
			} else if err != nil {
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

func patchDeploymentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		deployment := internalctx.GetDeployment(ctx)
		req, err := JsonBody[api.PatchDeploymentRequest](w, r)
		if err != nil {
			return
		}

		needsUpdate := false

		if req.LogsEnabled != nil && *req.LogsEnabled != deployment.LogsEnabled {
			deployment.LogsEnabled = *req.LogsEnabled
			needsUpdate = true
		}

		if needsUpdate {
			if err := db.UpdateDeployment(ctx, deployment); err != nil {
				log.Warn("deployment update failed", zap.Error(err))
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		RespondJSON(w, deployment)
	}
}

func deleteDeploymentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		auth := auth.Authentication.Require(ctx)
		orgId := *auth.CurrentOrgID()
		deployment := internalctx.GetDeployment(ctx)
		_ = db.RunTx(ctx, func(ctx context.Context) error {
			target, err := db.GetDeploymentTargetForDeploymentID(ctx, deployment.ID)
			if err != nil {
				log.Warn("could not get DeploymentTarget", zap.Error(err))
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}
			if target.OrganizationID != orgId || !isDeploymentTargetVisible(auth, target.DeploymentTarget) {
				http.NotFound(w, r)
				return apierrors.ErrNotFound
			}

			if err := db.DeleteDeploymentWithID(ctx, deployment.ID); err != nil {
				log.Warn("could not delete Deployment", zap.Error(err))
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			return nil
		})
	}
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

	org := auth.CurrentOrg()
	var err error

	if app, err = db.GetApplicationForApplicationVersionID(ctx, request.ApplicationVersionID, orgId); err != nil {
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

	var existingDeployment *types.DeploymentWithLatestRevision
	if request.DeploymentID != nil {
		for _, d := range target.Deployments {
			if d.ID == *request.DeploymentID {
				existingDeployment = &d
				break
			}
		}
		if existingDeployment == nil {
			return badRequestError(w, "DeploymentTarget doesn't have Deployment with the specified ID")
		}
	}

	if existingDeployment != nil {
		if request.ApplicationLicenseID == nil {
			if existingDeployment.ApplicationLicenseID != nil {
				request.ApplicationLicenseID = existingDeployment.ApplicationLicenseID
			}
		} else if existingDeployment.ApplicationLicenseID == nil {
			return badRequestError(w, "can not update license")
		} else if *request.ApplicationLicenseID != *existingDeployment.ApplicationLicenseID {
			return badRequestError(w, "can not update license")
		}
		if existingDeployment.ApplicationID != app.ID {
			return badRequestError(w, "can not change application of existing deployment")
		}
	}

	if org.HasFeature(types.FeatureLicensing) {
		if request.ApplicationLicenseID != nil {
			if license, err = db.GetApplicationLicenseByID(ctx, *request.ApplicationLicenseID); err != nil {
				if errors.Is(err, apierrors.ErrNotFound) {
					return licenseNotFoundError(w)
				} else {
					log.Error("could not get ApplicationLicense", zap.Error(err))
					sentry.GetHubFromContext(ctx).CaptureException(err)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return err
				}
			}
		} else if auth.CurrentCustomerOrgID() != nil {
			if licenses, err := db.GetApplicationLicensesWithOrganizationID(ctx, orgId, nil); err != nil {
				log.Error("could not get ApplicationLicense", zap.Error(err))
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return err
			} else if len(licenses) > 0 {
				// license ID is required for customer but optional for vendor
				return badRequestError(w, "applicationLicenseId is required")
			}
		}
	} else if request.ApplicationLicenseID != nil {
		return badRequestError(w, "unexpected applicationLicenseId")
	}

	if err = validateDeploymentRequestLicense(ctx, w, request, license, app, target, existingDeployment); err != nil {
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
	license *types.ApplicationLicense,
	app *types.Application,
	target *types.DeploymentTargetWithCreatedBy,
	deployment *types.DeploymentWithLatestRevision,
) error {
	if license != nil {
		auth := auth.Authentication.Require(ctx)

		if license.OrganizationID != *auth.CurrentOrgID() {
			return licenseNotFoundError(w)
		}
		if license.CustomerOrganizationID == nil {
			return invalidLicenseError(w)
		}
		if auth.CurrentCustomerOrgID() != nil && *license.CustomerOrganizationID != *auth.CurrentCustomerOrgID() {
			return licenseNotFoundError(w)
		}
		if target.CustomerOrganizationID == nil || *target.CustomerOrganizationID != *license.CustomerOrganizationID {
			return invalidLicenseError(w)
		}
		if len(license.Versions) > 0 && !license.HasVersionWithID(request.ApplicationVersionID) {
			return invalidLicenseError(w)
		}
		if app.ID != license.ApplicationID {
			return invalidLicenseError(w)
		}
		if deployment != nil && deployment.ApplicationID != license.ApplicationID {
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

	if !isDeploymentTargetVisible(auth, target.DeploymentTarget) {
		err := errors.New("DeploymentTarget not found")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	if request.DeploymentID == nil && len(target.Deployments) > 0 {
		if err := target.AgentVersion.CheckMultiDeploymentSupported(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return err
		}
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
	limit, err := QueryParam(r, "limit", strconv.Atoi, Max(100))
	if errors.Is(err, ErrParamNotDefined) {
		limit = 25
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	before, err := QueryParam(r, "before", ParseTimeFunc(time.RFC3339Nano))
	if err != nil && !errors.Is(err, ErrParamNotDefined) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	after, err := QueryParam(r, "after", ParseTimeFunc(time.RFC3339Nano))
	if err != nil && !errors.Is(err, ErrParamNotDefined) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if deploymentStatus, err := db.GetDeploymentStatus(ctx, deployment.ID, limit, before, after); err != nil {
		internalctx.GetLogger(ctx).Error("failed to get deploymentstatus", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else {
		RespondJSON(w, deploymentStatus)
	}
}

func exportDeploymentStatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)

		deployment := internalctx.GetDeployment(ctx)
		authInfo := auth.Authentication.Require(ctx)
		org := authInfo.CurrentOrg()
		limit := int(subscription.GetLogExportRowsLimit(org.SubscriptionType))

		filename := fmt.Sprintf("%s_deployment_status.log", time.Now().Format("2006-01-02"))

		SetFileDownloadHeaders(w, filename)

		err := db.GetDeploymentStatusForExport(
			ctx, deployment.ID, limit,
			func(record types.DeploymentRevisionStatus) error {
				_, err := fmt.Fprintf(w, "[%s] [%s] %s\n",
					record.CreatedAt.Format(time.RFC3339),
					record.Type,
					record.Message)
				return err
			},
		)
		if err != nil {
			log.Error("failed to export status records", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			// Note: If headers were already sent, we can't send error response
			return
		}
	}
}

func getDeploymentLogsResourcesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		deployment := internalctx.GetDeployment(ctx)
		if resources, err := db.GetDeploymentLogRecordResources(ctx, deployment.ID); err != nil {
			internalctx.GetLogger(ctx).Error("failed to get log records", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		} else {
			RespondJSON(w, resources)
		}
	}
}

func exportDeploymentLogsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)

		deployment := internalctx.GetDeployment(ctx)

		resource := r.FormValue("resource")
		if resource == "" {
			http.Error(w, "query parameter resource is required", http.StatusBadRequest)
			return
		}

		// authInfo := auth.Authentication.Require(ctx)
		// org := authInfo.CurrentOrg()

		// limit := int(subscription.GetLogExportRowsLimit(org.SubscriptionType))

		filename := fmt.Sprintf("%s_%s.log", time.Now().Format("2006-01-02"), resource)

		SetFileDownloadHeaders(w, filename)

		err := db.GetDeploymentLogRecordsForExport(
			ctx, deployment.ID, resource, 2_000_000,
			func(record types.DeploymentLogRecord) error {
				_, err := fmt.Fprintf(w, "[%s] [%s] %s\n",
					record.Timestamp.Format(time.RFC3339),
					record.Severity,
					record.Body)
				return err
			},
		)
		if err != nil {
			log.Error("failed to export log records", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			// Note: If headers were already sent, we can't send error response
			return
		}
	}
}

func getDeploymentLogsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		deployment := internalctx.GetDeployment(ctx)
		resource := r.FormValue("resource")
		if resource == "" {
			http.Error(w, "query parameter resource is required", http.StatusBadRequest)
			return
		}
		limit, err := QueryParam(r, "limit", strconv.Atoi, Max(100))
		if errors.Is(err, ErrParamNotDefined) {
			limit = 25
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		before, err := QueryParam(r, "before", ParseTimeFunc(time.RFC3339Nano))
		if err != nil && !errors.Is(err, ErrParamNotDefined) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		after, err := QueryParam(r, "after", ParseTimeFunc(time.RFC3339Nano))
		if err != nil && !errors.Is(err, ErrParamNotDefined) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if records, err := db.GetDeploymentLogRecords(ctx, deployment.ID, resource, limit, before, after); err != nil {
			internalctx.GetLogger(ctx).Error("failed to get log records", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		} else {
			response := make([]api.DeploymentLogRecord, len(records))
			for i, record := range records {
				response[i] = api.DeploymentLogRecord{
					DeploymentID:         record.DeploymentID,
					DeploymentRevisionID: record.DeploymentRevisionID,
					Resource:             record.Resource,
					Timestamp:            record.Timestamp,
					Severity:             record.Severity,
					Body:                 record.Body,
				}
			}
			RespondJSON(w, response)
		}
	}
}

func deploymentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		auth := auth.Authentication.Require(ctx)
		deploymentId, err := uuid.Parse(r.PathValue("deploymentId"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if deployment, err := db.GetDeployment(
			ctx,
			deploymentId,
			auth.CurrentUserID(),
			*auth.CurrentOrgID(),
			auth.CurrentCustomerOrgID(),
		); errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else if err != nil {
			internalctx.GetLogger(ctx).Error("failed to get deployment", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			ctx = internalctx.WithDeployment(ctx, deployment)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

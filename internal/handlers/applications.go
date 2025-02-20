package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/resources"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func ApplicationsRouter(r chi.Router) {
	r.Use(middleware.RequireOrgID, middleware.RequireUserRole)
	r.Get("/", getApplications)
	r.With(requireUserRoleVendor).Post("/", createApplication)
	r.With(requireUserRoleVendor).Post("/sample", createSampleApplication)
	r.Route("/{applicationId}", func(r chi.Router) {
		r.With(applicationMiddleware).Group(func(r chi.Router) {
			r.Get("/", getApplication)
			r.With(requireUserRoleVendor).Delete("/", deleteApplication)
			r.With(requireUserRoleVendor).Put("/", updateApplication)
		})
		r.Route("/versions", func(r chi.Router) {
			// note that it would not be necessary to use the applicationMiddleware for the versions endpoints
			// it loads the application from the db including all versions, but I guess for now this is easier
			// when performance becomes more important, we should avoid this and do the request on the database layer
			r.With(applicationMiddleware).Group(func(r chi.Router) {
				r.With(requireUserRoleVendor).Post("/", createApplicationVersion)
			})
			r.Route("/{applicationVersionId}", func(r chi.Router) {
				r.Get("/", getApplicationVersion)
				r.With(requireUserRoleVendor, applicationMiddleware).Put("/", updateApplicationVersion)
				r.Get("/compose-file", getApplicationVersionComposeFile)
				r.Get("/template-file", getApplicationVersionTemplateFile)
				r.Get("/values-file", getApplicationVersionValuesFile)
			})
		})
	})
}

func createApplication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)
	application, err := JsonBody[types.Application](w, r)
	if err != nil {
		return
	} else if application.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = db.CreateApplication(ctx, &application, *auth.CurrentOrgID()); err != nil {
		log.Warn("could not create application", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err = fmt.Fprintln(w, err); err != nil {
			log.Error("failed to write error to response", zap.Error(err))
		}
	} else if err = json.NewEncoder(w).Encode(application); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func updateApplication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)
	application, err := JsonBody[types.Application](w, r)
	if err != nil {
		return
	} else if application.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	existing := internalctx.GetApplication(ctx)
	if application.ID == uuid.Nil {
		application.ID = existing.ID
	} else if application.ID != existing.ID || application.Type != existing.Type {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := db.UpdateApplication(ctx, &application, *auth.CurrentOrgID()); err != nil {
		log.Warn("could not update application", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}
	// TODO ?
	// there surely is some way to have the update command returning the versions too, but I don't think it's worth
	// the work right now
	application.Versions = existing.Versions
	if err := json.NewEncoder(w).Encode(application); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func getApplications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)
	log := internalctx.GetLogger(ctx)

	org, err := db.GetOrganizationByID(ctx, *auth.CurrentOrgID())
	if err != nil {
		log.Error("failed to get org", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var applications []types.Application
	if org.HasFeature(types.FeatureLicensing) && *auth.CurrentUserRole() == types.UserRoleCustomer {
		applications, err = db.GetApplicationsWithLicenseOwnerID(ctx, auth.CurrentUserID())
	} else {
		applications, err = db.GetApplicationsByOrgID(ctx, *auth.CurrentOrgID())
	}

	if err != nil {
		log.Error("failed to get applications", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else {
		RespondJSON(w, applications)
	}
}

func getApplication(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, internalctx.GetApplication(r.Context()))
}

func getApplicationVersion(w http.ResponseWriter, r *http.Request) {
	applicationVersionID, err := uuid.Parse(r.PathValue("applicationVersionId"))
	if err != nil {
		http.NotFound(w, r)
	} else if applicationVersion, err := db.GetApplicationVersion(r.Context(), applicationVersionID); err != nil {
		if errors.Is(err, apierrors.ErrNotFound) {
			http.NotFound(w, r)
		} else {
			http.Error(w, "something went wrong", http.StatusInternalServerError)
		}
	} else {
		RespondJSON(w, applicationVersion)
	}
}

func createApplicationVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)

	body := r.FormValue("applicationversion")
	var applicationVersion types.ApplicationVersion
	if err := json.NewDecoder(strings.NewReader(body)).Decode(&applicationVersion); err != nil {
		log.Error("failed to deocde version", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	application := internalctx.GetApplication(ctx)
	applicationVersion.ApplicationID = application.ID

	if application.Type == types.DeploymentTypeDocker {
		if data, ok := readMultipartFile(w, r, "composefile"); !ok {
			return
		} else {
			applicationVersion.ComposeFileData = data
			if _, err := applicationVersion.ParsedComposeFile(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		if data, ok := readMultipartFile(w, r, "templatefile"); !ok {
			return
		} else {
			applicationVersion.TemplateFileData = data
		}
	} else {
		if data, ok := readMultipartFile(w, r, "valuesfile"); !ok {
			return
		} else {
			applicationVersion.ValuesFileData = data
			if _, err := applicationVersion.ParsedValuesFile(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		if data, ok := readMultipartFile(w, r, "templatefile"); !ok {
			return
		} else {
			// Template file is taken without parsing on purpose.
			// Some uses might use a non-yaml template here.
			applicationVersion.TemplateFileData = data
		}
	}

	if err := applicationVersion.Validate(application.Type); err != nil {
		log.Error("invalid application version", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.CreateApplicationVersion(ctx, &applicationVersion); err != nil {
		log.Warn("could not create applicationversion", zap.Error(err))
		sentry.GetHubFromContext(r.Context()).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err = fmt.Fprintln(w, err); err != nil {
			log.Error("failed to write error to response", zap.Error(err))
		}
	} else if err = json.NewEncoder(w).Encode(applicationVersion); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func updateApplicationVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	applicationVersion, err := JsonBody[types.ApplicationVersion](w, r)
	if err != nil {
		return
	}

	applicationVersionIdFromUrl, err := uuid.Parse(r.PathValue("applicationVersionId"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	existing := internalctx.GetApplication(ctx)
	var existingVersion *types.ApplicationVersion
	for _, version := range existing.Versions {
		if version.ID == applicationVersionIdFromUrl {
			existingVersion = &version
		}
	}
	if existingVersion == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if applicationVersion.ID == uuid.Nil {
		applicationVersion.ID = existingVersion.ID
	}

	if err := db.UpdateApplicationVersion(ctx, &applicationVersion); err != nil {
		log.Warn("could not update applicationversion", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		RespondJSON(w, applicationVersion)
	}
}

var getApplicationVersionComposeFile = getApplicationVersionFileHandler(func(av types.ApplicationVersion) []byte {
	return av.ComposeFileData
})
var getApplicationVersionValuesFile = getApplicationVersionFileHandler(func(av types.ApplicationVersion) []byte {
	return av.ValuesFileData
})
var getApplicationVersionTemplateFile = getApplicationVersionFileHandler(func(av types.ApplicationVersion) []byte {
	return av.TemplateFileData
})

func getApplicationVersionFileHandler(fileAccessor func(types.ApplicationVersion) []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		applicationVersionID, err := uuid.Parse(r.PathValue("applicationVersionId"))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if v, err := db.GetApplicationVersion(ctx, applicationVersionID); errors.Is(err, apierrors.ErrNotFound) {
			http.NotFound(w, r)
		} else if err != nil {
			log.Error("failed to get ApplicationVersion from DB", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else if data := fileAccessor(*v); data == nil {
			http.NotFound(w, r)
		} else {
			w.Header().Add("Content-Type", "application/yaml")
			w.Header().Add("Cache-Control", "max-age=300, private")
			if _, err := w.Write(data); err != nil {
				log.Warn("failed to write file to response", zap.Error(err))
			}
		}
	}
}

func createSampleApplication(w http.ResponseWriter, r *http.Request) {
	// TODO only serve request if user does not have a sample application yet
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)

	application := types.Application{
		Name: "Shiori",
		Type: types.DeploymentTypeDocker,
	}

	var composeFileData []byte
	var templateFileData []byte
	if composeFile, err := resources.Get("apps/shiori/docker-compose.yaml"); err != nil {
		log.Warn("failed to read shiori compose file", zap.Error(err))
	} else {
		composeFileData = composeFile
	}
	if templateFile, err := resources.Get("apps/shiori/template.env"); err != nil {
		log.Warn("failed to read shiori template file", zap.Error(err))
	} else {
		templateFileData = templateFile
	}

	version := types.ApplicationVersion{
		Name:             "v1.7.4",
		ComposeFileData:  composeFileData,
		TemplateFileData: templateFileData,
	}

	if err := db.RunTx(ctx, pgx.TxOptions{}, func(ctx context.Context) error {
		if err := db.CreateApplication(ctx, &application, *auth.CurrentOrgID()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		version.ApplicationID = application.ID
		if err := db.CreateApplicationVersion(ctx, &version); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		return nil
	}); err != nil {
		log.Warn("could not create sample application", zap.Error(err))
		return
	}

	application.Versions = append(application.Versions, version)
	RespondJSON(w, application)
}

func deleteApplication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	application := internalctx.GetApplication(ctx)
	auth := auth.Authentication.Require(ctx)
	if application.OrganizationID != *auth.CurrentOrgID() {
		http.NotFound(w, r)
	} else if err := db.DeleteApplicationWithID(ctx, application.ID); err != nil {
		if errors.Is(err, apierrors.ErrConflict) {
			http.Error(w, "could not delete Application because it is still in use", http.StatusBadRequest)
		} else {
			log.Warn("error deleting application", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func applicationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		applicationID, err := uuid.Parse(r.PathValue("applicationId"))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		auth := auth.Authentication.Require(ctx)
		application, err := db.GetApplication(ctx, applicationID, *auth.CurrentOrgID())
		if errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else if err != nil {
			internalctx.GetLogger(r.Context()).Error("failed to get application", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			ctx = internalctx.WithApplication(ctx, application)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/glasskube/cloud/internal/contenttype"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/cloud/internal/auth"
	"github.com/glasskube/cloud/internal/resources"
	"github.com/jackc/pgx/v5"

	"github.com/glasskube/cloud/internal/apierrors"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/types"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func ApplicationsRouter(r chi.Router) {
	r.Get("/", getApplications)
	r.With(requireUserRoleVendor).Post("/", createApplication)
	r.With(requireUserRoleVendor).Post("/sample", createSampleApplication)
	r.Route("/{applicationId}", func(r chi.Router) {
		r.Use(applicationMiddleware)
		r.Get("/", getApplication)
		r.With(requireUserRoleVendor).Delete("/", deleteApplication)
		r.With(requireUserRoleVendor).Put("/", updateApplication)
		r.Route("/versions", func(r chi.Router) {
			// note that it would not be necessary to use the applicationMiddleware for the versions endpoints
			// it loads the application from the db including all versions, but I guess for now this is easier
			// when performance becomes more important, we should avoid this and do the request on the database layer
			r.Get("/", getApplicationVersions)
			r.With(requireUserRoleVendor).Post("/", createApplicationVersion)
			r.Route("/{applicationVersionId}", func(r chi.Router) {
				r.Get("/", getApplicationVersion)
				r.With(requireUserRoleVendor).Put("/", updateApplicationVersion)
				r.Get("/compose-file", getApplicationVersionComposeFile)
			})
		})
	})
}

func createApplication(w http.ResponseWriter, r *http.Request) {
	log := internalctx.GetLogger(r.Context())
	var application types.Application
	if err := json.NewDecoder(r.Body).Decode(&application); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if err = db.CreateApplication(r.Context(), &application); err != nil {
		log.Warn("could not create application", zap.Error(err))
		sentry.GetHubFromContext(r.Context()).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err = fmt.Fprintln(w, err); err != nil {
			log.Error("failed to write error to response", zap.Error(err))
		}
	} else if err = json.NewEncoder(w).Encode(application); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func updateApplication(w http.ResponseWriter, r *http.Request) {
	log := internalctx.GetLogger(r.Context())
	var application types.Application
	if err := json.NewDecoder(r.Body).Decode(&application); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	existing := internalctx.GetApplication(r.Context())
	if application.ID == "" {
		application.ID = existing.ID
	} else if application.ID != existing.ID || application.Type != existing.Type {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := db.UpdateApplication(r.Context(), &application); err != nil {
		log.Warn("could not update application", zap.Error(err))
		sentry.GetHubFromContext(r.Context()).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}
	// there surely is some way to have the update command returning the versions too, but I don't think it's worth
	// the work right now
	application.Versions = existing.Versions
	if err := json.NewEncoder(w).Encode(application); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func getApplications(w http.ResponseWriter, r *http.Request) {
	if applications, err := db.GetApplications(r.Context()); err != nil {
		internalctx.GetLogger(r.Context()).Error("failed to get applications", zap.Error(err))
		sentry.GetHubFromContext(r.Context()).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		err := json.NewEncoder(w).Encode(applications)
		if err != nil {
			internalctx.GetLogger(r.Context()).Error("failed to encode to json", zap.Error(err))
		}
	}
}

func getApplication(w http.ResponseWriter, r *http.Request) {
	application := internalctx.GetApplication(r.Context())
	// in the future we might want to transform the application to a well-defined endpoint-type instead of passing through
	// could use the https://github.com/go-chi/render package for that or we do it ourselves
	err := json.NewEncoder(w).Encode(application)
	if err != nil {
		internalctx.GetLogger(r.Context()).Error("failed to encode to json", zap.Error(err))
	}
}

func getApplicationVersions(w http.ResponseWriter, r *http.Request) {
	application := internalctx.GetApplication(r.Context())
	err := json.NewEncoder(w).Encode(application.Versions)
	if err != nil {
		internalctx.GetLogger(r.Context()).Error("failed to encode to json", zap.Error(err))
	}
}

func getApplicationVersion(w http.ResponseWriter, r *http.Request) {
	application := internalctx.GetApplication(r.Context())
	applicationVersionId := r.PathValue("applicationVersionId")
	// once performance becomes more important, do not load the whole application but only the requested version
	for _, applicationVersion := range application.Versions {
		if applicationVersion.ID == applicationVersionId {
			err := json.NewEncoder(w).Encode(applicationVersion)
			if err != nil {
				internalctx.GetLogger(r.Context()).Error("failed to encode to json", zap.Error(err))
			}
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
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
	applicationVersion.ApplicationId = application.ID

	if application.Type == types.DeploymentTypeDocker {
		if data := readFile(w, r, "composefile", true); data == nil {
			return
		} else {
			applicationVersion.ComposeFileData = data
		}
	} else {
		if data := readFile(w, r, "valuesfile", false); data != nil {
			applicationVersion.ValuesFileData = data
		}
		if data := readFile(w, r, "templatefile", false); data != nil {
			applicationVersion.TemplateFileData = data
		}
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

func readFile(w http.ResponseWriter, r *http.Request, formKey string, required bool) *[]byte {
	log := internalctx.GetLogger(r.Context())
	if file, head, err := r.FormFile(formKey); err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			log.Error("failed to get file from upload", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return nil
		} else if required {
			log.Error("required file not given", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return nil
		}
	} else {
		log.Sugar().Debugf("got file %v with type %v and size %v", head.Filename, head.Header, head.Size)
		// max file size is 100KiB
		if head.Size > 102400 {
			log.Debug("large body was rejected")
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			fmt.Fprintln(w, "file too large (max 100 KiB)")
			return nil
		} else if err := contenttype.IsYaml(head.Header); err != nil {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			fmt.Fprint(w, err)
			return nil
		} else if data, err := io.ReadAll(file); err != nil {
			log.Error("failed to read file from upload", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return nil
		} else {
			return &data
		}
	}
	return nil
}

func updateApplicationVersion(w http.ResponseWriter, r *http.Request) {
	log := internalctx.GetLogger(r.Context())
	var applicationVersion types.ApplicationVersion
	if err := json.NewDecoder(r.Body).Decode(&applicationVersion); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	applicationVersionIdFromUrl := r.PathValue("applicationVersionId")
	existing := internalctx.GetApplication(r.Context())
	var existingVersion *types.ApplicationVersion
	for _, version := range existing.Versions {
		if version.ID == applicationVersionIdFromUrl {
			existingVersion = &version
		}
	}
	if existingVersion == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if applicationVersion.ID == "" {
		applicationVersion.ID = existingVersion.ID
	}

	if err := db.UpdateApplicationVersion(r.Context(), &applicationVersion); err != nil {
		log.Warn("could not update applicationversion", zap.Error(err))
		sentry.GetHubFromContext(r.Context()).CaptureException(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
	} else if err = json.NewEncoder(w).Encode(applicationVersion); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func getApplicationVersionComposeFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	application := internalctx.GetApplication(ctx)
	applicationVersionId := r.PathValue("applicationVersionId")
	// once performance becomes more important, do not load the whole application but only the requested version
	for _, applicationVersion := range application.Versions {
		if applicationVersion.ID == applicationVersionId {
			if data, err := db.GetApplicationVersionComposeFile(ctx, applicationVersionId); err != nil {
				log.Error("failed to get compose file from DB", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, err)
			} else if data == nil {
				break
			} else {
				w.Header().Add("Content-Type", "application/yaml")
				if _, err := w.Write(data); err != nil {
					log.Error("failed to write compose file to response", zap.Error(err))
				}
			}
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func createSampleApplication(w http.ResponseWriter, r *http.Request) {
	// TODO only serve request if user does not have a sample application yet
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)

	application := types.Application{
		Name: "Shiori",
		Type: types.DeploymentTypeDocker,
	}

	var composeFileData []byte
	if composeFile, err := resources.Get("embedded/shiori-compose.yaml"); err != nil {
		log.Warn("failed to read shiori compose file", zap.Error(err))
	} else {
		composeFileData = composeFile
	}

	version := types.ApplicationVersion{
		Name:            "v1.7.1",
		ComposeFileData: &composeFileData,
	}

	if err := db.RunTx(ctx, pgx.TxOptions{}, func(ctx context.Context) error {
		if err := db.CreateApplication(ctx, &application); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		version.ApplicationId = application.ID
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
	if orgId, err := auth.CurrentOrgId(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if application.OrganizationID != orgId {
		http.NotFound(w, r)
	} else if err := db.DeleteApplicationWithID(ctx, application.ID); err != nil {
		log.Warn("error deleting application", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func applicationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		applicationId := r.PathValue("applicationId")
		application, err := db.GetApplication(ctx, applicationId)
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

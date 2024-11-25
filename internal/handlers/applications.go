package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"go.uber.org/zap"

	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/types"
	"github.com/go-chi/chi/v5"
)

func ApplicationsRouter(r chi.Router) {
	// TODO r.Use(AuthMiddleware)
	r.Get("/", getApplications)
	r.Post("/", createApplication)
	r.Route("/{applicationId}", func(r chi.Router) {
		r.Use(applicationMiddleware)
		r.Get("/", getApplication)
		r.Put("/", updateApplication)
		r.Route("/versions", func(r chi.Router) {
			// note that it would not be necessary to use the applicationMiddleware for the versions endpoints
			// it loads the application from the db including all versions, but I guess for now this is easier
			// when performance becomes more important, we should avoid this and do the request on the database layer
			r.Get("/", getApplicationVersions)
			r.Post("/", createApplicationVersion)
			r.Route("/{applicationVersionId}", func(r chi.Router) {
				r.Get("/", getApplicationVersion)
				r.Put("/", updateApplicationVersion)
			})
		})
	})
}

func createApplication(w http.ResponseWriter, r *http.Request) {
	log := internalctx.GetLoggerOrPanic(r.Context())
	var application types.Application
	if err := json.NewDecoder(r.Body).Decode(&application); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if err = db.CreateApplication(r.Context(), &application); err != nil {
		log.Warn("could not create application", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		if _, err = fmt.Fprintln(w, err); err != nil {
			log.Error("failed to write error to response", zap.Error(err))
		}
	} else if err = json.NewEncoder(w).Encode(application); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func updateApplication(w http.ResponseWriter, r *http.Request) {
	log := internalctx.GetLoggerOrPanic(r.Context())
	var application types.Application
	if err := json.NewDecoder(r.Body).Decode(&application); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	existing := internalctx.GetApplicationOrPanic(r.Context())
	if application.ID == "" {
		application.ID = existing.ID
	} else if application.ID != existing.ID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := db.UpdateApplication(r.Context(), &application); err != nil {
		log.Warn("could not update application", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
	} else if err = json.NewEncoder(w).Encode(application); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func getApplications(w http.ResponseWriter, r *http.Request) {
	if applications, err := db.GetApplications(r.Context()); err != nil {
		internalctx.GetLoggerOrPanic(r.Context()).Error("failed to get applications", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		err := json.NewEncoder(w).Encode(applications)
		if err != nil {
			internalctx.GetLoggerOrPanic(r.Context()).Error("failed to encode to json", zap.Error(err))
		}
	}
}

func getApplication(w http.ResponseWriter, r *http.Request) {
	application := internalctx.GetApplicationOrPanic(r.Context())
	// in the future we might want to transform the application to a well-defined endpoint-type instead of passing through
	// could use the https://github.com/go-chi/render package for that or we do it ourselves
	err := json.NewEncoder(w).Encode(application)
	if err != nil {
		internalctx.GetLoggerOrPanic(r.Context()).Error("failed to encode to json", zap.Error(err))
	}
}

func getApplicationVersions(w http.ResponseWriter, r *http.Request) {
	application := internalctx.GetApplicationOrPanic(r.Context())
	err := json.NewEncoder(w).Encode(application.Versions)
	if err != nil {
		internalctx.GetLoggerOrPanic(r.Context()).Error("failed to encode to json", zap.Error(err))
	}
}

func getApplicationVersion(w http.ResponseWriter, r *http.Request) {
	application := internalctx.GetApplicationOrPanic(r.Context())
	applicationVersionId := chi.URLParam(r, "applicationVersionId")
	// once performance becomes more important, do not load the whole application but only the requested version
	for _, applicationVersion := range application.Versions {
		if applicationVersion.ID == applicationVersionId {
			err := json.NewEncoder(w).Encode(applicationVersion)
			if err != nil {
				internalctx.GetLoggerOrPanic(r.Context()).Error("failed to encode to json", zap.Error(err))
			}
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func createApplicationVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLoggerOrPanic(ctx)
	if file, _, err := r.FormFile("docker-compose"); err != nil {
		log.Error("failed to read file from upload", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		// TODO
		fmt.Fprintf(os.Stderr, "%v\n", file)
	}
	var applicationVersion types.ApplicationVersion
	if err := json.NewDecoder(r.Body).Decode(&applicationVersion); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	application := internalctx.GetApplicationOrPanic(ctx)
	applicationVersion.ApplicationId = application.ID
	if err := db.CreateApplicationVersion(ctx, &applicationVersion); err != nil {
		log.Warn("could not create applicationversion", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		if _, err = fmt.Fprintln(w, err); err != nil {
			log.Error("failed to write error to response", zap.Error(err))
		}
	} else if err = json.NewEncoder(w).Encode(applicationVersion); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func updateApplicationVersion(w http.ResponseWriter, r *http.Request) {
	log := internalctx.GetLoggerOrPanic(r.Context())
	var applicationVersion types.ApplicationVersion
	if err := json.NewDecoder(r.Body).Decode(&applicationVersion); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	applicationVersionIdFromUrl := chi.URLParam(r, "applicationVersionId")
	existing := internalctx.GetApplicationOrPanic(r.Context())
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
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
	} else if err = json.NewEncoder(w).Encode(applicationVersion); err != nil {
		log.Error("failed to encode json", zap.Error(err))
	}
}

func applicationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		applicationId := chi.URLParam(r, "applicationId")
		application, err := db.GetApplication(ctx, applicationId)
		if err != nil {
			internalctx.GetLoggerOrPanic(r.Context()).Error("failed to get application", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		} else if application == nil {
			w.WriteHeader(http.StatusNotFound)
		} else {
			ctx = internalctx.WithApplication(ctx, application)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

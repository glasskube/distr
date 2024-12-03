package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/glasskube/cloud/internal/apierrors"
	"github.com/glasskube/cloud/internal/contenttype"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/glasskube/cloud/internal/types"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func ApplicationsRouter(r chi.Router) {
	// TODO r.Use(AuthMiddleware)
	r.Get("/", getApplications)
	r.Post("/", createApplication)
	r.Post("/sample", createSampleApplication)
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
	} else if application.ID != existing.ID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := db.UpdateApplication(r.Context(), &application); err != nil {
		log.Warn("could not update application", zap.Error(err))
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

	if file, head, err := r.FormFile("file"); err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			log.Error("failed to get file from upload", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		log.Sugar().Debugf("got file %v with type %v and size %v", head.Filename, head.Header, head.Size)
		// max file size is 100KiB
		if head.Size > 102400 {
			log.Debug("large body was rejected")
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			fmt.Fprintln(w, "file too large (max 100 KiB)")
			return
		} else if err := contenttype.IsYaml(head.Header); err != nil {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			fmt.Fprint(w, err)
			return
		} else if data, err := io.ReadAll(file); err != nil {
			log.Error("failed to read file from upload", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			applicationVersion.ComposeFileData = &data
		}
	}

	application := internalctx.GetApplication(ctx)
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
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	// TODO only serve request if user does not have a sample application yet
	application := types.Application{
		Name: "Shiori",
		Type: types.DeploymentTypeDocker,
	}
	if err := db.CreateApplication(ctx, &application); err != nil {
		log.Warn("could not create sample application", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		if _, err = fmt.Fprintln(w, err); err != nil {
			log.Error("failed to write error to response", zap.Error(err))
		}
		return
	}
	version := types.ApplicationVersion{
		Name:            "v1.0.0",
		ComposeFileData: &shioriComposeFile,
		ApplicationId:   application.ID,
	}
	if err := db.CreateApplicationVersion(ctx, &version); err != nil {
		log.Warn("could not create sample applicationversion", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		if _, err = fmt.Fprintln(w, err); err != nil {
			log.Error("failed to write error to response", zap.Error(err))
		}
	} else {
		application.Versions = append(application.Versions, version)
		if err = json.NewEncoder(w).Encode(application); err != nil {
			log.Error("failed to encode json", zap.Error(err))
		}
	}
}

func applicationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		applicationId := r.PathValue("applicationId")
		application, err := db.GetApplication(ctx, applicationId)
		if errors.Is(err, apierrors.NotFound) {
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

// TODO maybe shiori is not the best sample app, because it doesn't have a "production" docker compose file
// from https://github.com/go-shiori/shiori/blob/master/docker-compose.yaml
var shioriComposeFile = []byte(`# Docker compose for development purposes only.
# Edit it to fit your current development needs.
version: "3"
services:
  shiori:
    build:
      context: .
      dockerfile: Dockerfile.compose
    container_name: shiori
    ports:
      - "8080:8080"
    volumes:
      - "./dev-data:/srv/shiori"
      - ".:/src/shiori"
    restart: unless-stopped
    links:
      - "postgres"
      - "mariadb"
    environment:
      SHIORI_DIR: /srv/shiori
      #SHIORI_DATABASE_URL: mysql://shiori:shiori@(mariadb)/shiori?charset=utf8mb4
      SHIORI_DATABASE_URL: postgres://shiori:shiori@postgres/shiori?sslmode=disable

  postgres:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD: shiori
      POSTGRES_USER: shiori
    ports:
      - "5432:5432"

  mariadb:
    image: mariadb:11
    environment:
      MYSQL_ROOT_PASSWORD: toor
      MYSQL_DATABASE: shiori
      MYSQL_USER: shiori
      MYSQL_PASSWORD: shiori
    ports:
      - "3306:3306"
`)

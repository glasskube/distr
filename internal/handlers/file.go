package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/glasskube/distr/api"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/apierrors"
	"github.com/glasskube/distr/internal/auth"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func FileRouter(r chi.Router) {
	r.With(middleware.RequireOrgAndRole).Group(func(r chi.Router) {
		r.Post("/", createFileHandler)
		r.Route("/{fileId}", func(r chi.Router) {
			r.Use(fileMiddleware)
			r.Get("/", getFileHandler)
			r.Delete("/", deleteFileHandler)
		})
	})
}

func getFileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	file := internalctx.GetFile(ctx)

	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.FileName))
	w.Header().Set("Cache-Control", "max-age=604800, private")

	// Write file data to response
	if _, err := w.Write(file.Data); err != nil {
		internalctx.GetLogger(ctx).Warn("failed to write file to response", zap.Error(err))
		http.Error(w, "failed to write file to response", http.StatusInternalServerError)
	}
}

func deleteFileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	file := internalctx.GetFile(ctx)

	if err := db.DeleteFileWithID(ctx, file.ID); err != nil {
		log.Warn("error deleting file", zap.Error(err))
		if errors.Is(err, apierrors.ErrNotFound) {
			w.WriteHeader(http.StatusNoContent)
		} else if errors.Is(err, apierrors.ErrConflict) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func createFileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)

	auth := auth.Authentication.Require(ctx)

	if file, err := getFileFromRequest(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if err := db.CreateFile(ctx, *auth.CurrentOrgID(), file); err != nil {
		log.Warn("error uploading file", zap.Error(err))
		sentry.GetHubFromContext(ctx).CaptureException(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		RespondJSON(w, file.ID)
	}
}

func getFileFromRequest(r *http.Request) (*types.File, error) {
	if err := r.ParseMultipartForm(102400); err != nil {
		return nil, fmt.Errorf("failed to parse form: %w", err)
	}

	file := types.File{}

	if multiPartFile, fileHeader, err := r.FormFile("file"); err != nil {
		return nil, errors.New("file not found")
	} else if fileHeader.Size > 102400 {
		return nil, errors.New("file too large (max 100 KiB)")
	} else if data, err := io.ReadAll(multiPartFile); err != nil {
		return nil, err
	} else {
		file.Data = data
		file.FileSize = fileHeader.Size
		file.FileName = fileHeader.Filename
		if contentType := util.PtrTo(fileHeader.Header.Get("Content-Type")); contentType != nil {
			file.ContentType = *contentType
		} else {
			file.ContentType = "application/octet-stream"
		}
	}

	return &file, nil
}

func fileMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)
		auth := auth.Authentication.Require(ctx)

		if fileID, err := uuid.Parse(r.PathValue("fileId")); err != nil {
			http.NotFound(w, r)
		} else if file, err := db.GetFileWithID(ctx, fileID); err != nil {
			if errors.Is(err, apierrors.ErrNotFound) {
				http.NotFound(w, r)
			} else {
				log.Error("error getting file", zap.Error(err))
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		} else if orgs, err := db.GetOrganizationsForUser(ctx, auth.CurrentUserID()); err != nil {
			// TODO not sure yet if its the right way to check this, or if we should make org id "nullable"
			// (client would have to say at upload if its "public" or org scoped) and it would need a migration
			log.Error("error getting orgs", zap.Error(err))
			sentry.GetHubFromContext(ctx).CaptureException(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		} else if !slices.ContainsFunc(orgs, func(role types.OrganizationWithUserRole) bool {
			return role.ID == file.OrganizationID
		}) {
			http.NotFound(w, r)
		} else {
			h.ServeHTTP(w, r.WithContext(internalctx.WithFile(ctx, file)))
		}
	})
}

func patchImageHandler(patchImage func(ctx context.Context, body api.PatchImageRequest) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := internalctx.GetLogger(ctx)

		body, err := JsonBody[api.PatchImageRequest](w, r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if body.ImageID == uuid.Nil {
			http.Error(w, "imageId can not be empty", http.StatusBadRequest)
			return
		}

		if result, err := patchImage(ctx, body); err != nil {
			log.Warn("error patching image id", zap.Error(err))
			if errors.Is(err, apierrors.ErrNotFound) {
				w.WriteHeader(http.StatusNoContent)
			} else if errors.Is(err, apierrors.ErrConflict) {
				http.Error(w, err.Error(), http.StatusBadRequest)
			} else {
				sentry.GetHubFromContext(ctx).CaptureException(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			RespondJSON(w, result)
		}
	}
}

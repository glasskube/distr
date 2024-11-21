package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/db"
	"github.com/go-chi/chi/v5"
)

func ApplicationsRouter(r chi.Router) {
	// TODO r.Use(AuthMiddleware)
	r.Get("/", getApplications)
	r.Route("/{applicationId}", func(r chi.Router) {
		r.Use(applicationMiddleware)
		r.Get("/", getApplication)
	})
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

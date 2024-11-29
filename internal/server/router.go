package server

import (
	"net/http"
	"time"

	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func ApiRouter() chi.Router {
	// TODO for all (most) routes auth middleware
	router := chi.NewRouter()
	router.Use(
		middleware.RequestID,
		loggerCtxMiddleware,
		loggingMiddleware,
		dbCtxMiddleware,
	)
	router.Route("/applications", handlers.ApplicationsRouter)
	router.Route("/deployments", handlers.DeploymentsRouter)
	router.Route("/deployment-targets", handlers.DeploymentTargetsRouter)
	return router
}

func dbCtxMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := getDbPool()
		ctx := internalctx.WithDb(r.Context(), db)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func loggerCtxMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := getLogger().
			With(zap.String("requestId", middleware.GetReqID(r.Context())))
		ctx := internalctx.WithLogger(r.Context(), logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func loggingMiddleware(wh http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		now := time.Now()
		wh.ServeHTTP(ww, r)
		elapsed := time.Since(now)
		logger := internalctx.GetLogger(r.Context())
		logger.Info("handling request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", ww.Status()),
			zap.String("time", elapsed.String()))
	}
	return http.HandlerFunc(fn)
}

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
		loggingMiddleware(getLogger()),
		middleware.Recoverer,
		loggerCtxMiddleware,
		dbCtxMiddleware,
	)
	router.Route("/applications", handlers.ApplicationsRouter)
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
		logger := getLogger()
		ctx := internalctx.WithLogger(r.Context(), logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func loggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(wh http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			now := time.Now()
			wh.ServeHTTP(ww, r)
			elapsed := time.Since(now)
			logger.Info("handling request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.Status()),
				zap.String("time", elapsed.String()))
		}
		return http.HandlerFunc(fn)
	}
}

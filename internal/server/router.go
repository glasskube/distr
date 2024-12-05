package server

import (
	"net/http"
	"time"

	"github.com/glasskube/cloud/internal/auth"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"
)

func ApiRouter() chi.Router {
	router := chi.NewRouter()
	router.Use(
		middleware.RequestID,
		loggerCtxMiddleware,
		loggingMiddleware,
		dbCtxMiddleware,
	)

	// public routes go here
	router.Group(func(r chi.Router) {
		r.Route("/auth", handlers.AuthRouter)
	})

	// authenticated routes go here
	router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(auth.JWTAuth))
		r.Use(jwtauth.Authenticator(auth.JWTAuth))
		r.Route("/applications", handlers.ApplicationsRouter)
		r.Route("/deployments", handlers.DeploymentsRouter)
		r.Route("/deployment-targets", handlers.DeploymentTargetsRouter)
	})

	// agent connect and download routes go here (authenticated but with accessKeyId and accessKeySecret)
	router.Group(func(r chi.Router) {
		r.Route("/", handlers.AgentRouter)
	})

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

package server

import (
	"net/http"
	"time"

	"github.com/glasskube/cloud/internal/auth"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/frontend"
	"github.com/glasskube/cloud/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"
)

func NewRouter(s *server) chi.Router {
	router := chi.NewRouter()
	router.Use(
		// Handles panics
		middleware.Recoverer,
		// Reject bodies larger than 1MiB
		middleware.RequestSize(1048576),
	)
	router.Mount("/api", ApiRouter(s))
	router.Mount("/", FrontendRouter())
	return router
}

func ApiRouter(s *server) http.Handler {
	router := chi.NewRouter()
	router.Use(
		middleware.RequestID,
		loggerCtxMiddleware(s),
		loggingMiddleware,
		contextInjectorMiddelware(s),
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

	return router
}

func FrontendRouter() http.Handler {
	router := chi.NewRouter()
	router.Use(
		middleware.Compress(5, "text/html", "text/css", "text/javascript"),
	)

	router.Handle("/*", handlers.StaticFileHandler(frontend.BrowserFS()))

	return router
}

func contextInjectorMiddelware(s *server) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = internalctx.WithDb(ctx, s.GetDbPool())
			ctx = internalctx.WithMailer(ctx, s.GetMailer())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func loggerCtxMiddleware(s *server) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := s.GetLogger().
				With(zap.String("requestId", middleware.GetReqID(r.Context())))
			ctx := internalctx.WithLogger(r.Context(), logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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

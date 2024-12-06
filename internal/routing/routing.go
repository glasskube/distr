package routing

import (
	"net/http"

	"github.com/glasskube/cloud/internal/auth"
	"github.com/glasskube/cloud/internal/frontend"
	"github.com/glasskube/cloud/internal/handlers"
	"github.com/glasskube/cloud/internal/mail"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewRouter(logger *zap.Logger, db *pgxpool.Pool, mailer mail.Mailer) http.Handler {
	router := chi.NewRouter()
	router.Use(
		// Handles panics
		middleware.Recoverer,
		// Reject bodies larger than 1MiB
		middleware.RequestSize(1048576),
	)
	router.Mount("/api", ApiRouter(logger, db, mailer))
	router.Mount("/", FrontendRouter())
	return router
}

func ApiRouter(logger *zap.Logger, db *pgxpool.Pool, mailer mail.Mailer) http.Handler {
	router := chi.NewRouter()
	router.Use(
		middleware.RequestID,
		loggerCtxMiddleware(logger),
		loggingMiddleware,
		contextInjectorMiddelware(db, mailer),
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

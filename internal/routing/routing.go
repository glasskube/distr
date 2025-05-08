package routing

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/glasskube/distr/internal/auth"
	"github.com/glasskube/distr/internal/frontend"
	"github.com/glasskube/distr/internal/handlers"
	"github.com/glasskube/distr/internal/mail"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

func NewRouter(logger *zap.Logger, db *pgxpool.Pool, mailer mail.Mailer, tracer *trace.TracerProvider) http.Handler {
	router := chi.NewRouter()
	router.Use(
		// Handles panics
		chimiddleware.Recoverer,
		// Reject bodies larger than 1MiB
		chimiddleware.RequestSize(1048576),
	)
	router.Mount("/api", ApiRouter(logger, db, mailer, tracer))
	router.Mount("/internal", InternalRouter())
	router.Mount("/", FrontendRouter())
	return router
}

func ApiRouter(logger *zap.Logger, db *pgxpool.Pool, mailer mail.Mailer, tracer *trace.TracerProvider) http.Handler {
	r := chi.NewRouter()
	r.Use(
		chimiddleware.RequestID,
		chimiddleware.RealIP,
		middleware.Sentry,
		otelhttp.NewMiddleware("", otelhttp.WithTracerProvider(tracer)),
		middleware.LoggerCtxMiddleware(logger),
		middleware.LoggingMiddleware,
		middleware.ContextInjectorMiddleware(db, mailer),
	)

	r.Route("/v1", func(r chi.Router) {
		// public routes go here
		r.Group(func(r chi.Router) {
			r.Route("/auth", handlers.AuthRouter)
		})

		// authenticated routes go here
		r.Group(func(r chi.Router) {
			r.Use(
				middleware.SentryUser,
				auth.Authentication.Middleware,
				httprate.Limit(30, 1*time.Second, httprate.WithKeyFuncs(middleware.RateLimitCurrentUserIdKeyFunc)),
				httprate.Limit(60, 1*time.Minute, httprate.WithKeyFuncs(middleware.RateLimitCurrentUserIdKeyFunc)),
				httprate.Limit(2000, 1*time.Hour, httprate.WithKeyFuncs(middleware.RateLimitCurrentUserIdKeyFunc)),

				// TODO (low-prio) in the future, additionally check token audience and require it to be "api"/"user",
				// such that agents cant access anything here (they also can't now, because their tokens will not
				// pass the Authentication chain (DbAuthenticator can't find the user -> 401)
			)
			r.Route("/applications", handlers.ApplicationsRouter)
			r.Route("/application-licenses", handlers.ApplicationLicensesRouter)
			r.Route("/agent-versions", handlers.AgentVersionsRouter)
			r.Route("/artifacts", handlers.ArtifactsRouter)
			r.Route("/artifact-licenses", handlers.ArtifactLicensesRouter)
			r.Route("/artifact-pulls", handlers.ArtifactPullsRouter)
			r.Route("/dashboard", handlers.DashboardRouter)
			r.Route("/deployments", handlers.DeploymentsRouter)
			r.Route("/deployment-targets", handlers.DeploymentTargetsRouter)
			r.Route("/deployment-target-metrics", handlers.DeploymentTargetMetricsRouter)
			r.Route("/files", handlers.FileRouter)
			r.Route("/organization", handlers.OrganizationRouter)
			r.Route("/settings", handlers.SettingsRouter)
			r.Route("/user-accounts", handlers.UserAccountsRouter)
			r.Route("/tutorial-progress", handlers.TutorialsRouter)
		})

		// agent connect and download routes go here (authenticated but with accessKeyId and accessKeySecret)
		r.Group(func(r chi.Router) {
			r.Route("/", handlers.AgentRouter)
		})
	})

	return r
}

func InternalRouter() http.Handler {
	router := chi.NewRouter()
	router.Route("/", handlers.InternalRouter)
	return router
}

func FrontendRouter() http.Handler {
	router := chi.NewRouter()
	router.Use(
		chimiddleware.Compress(5, "text/html", "text/css", "text/javascript"),
	)

	router.Handle("/*", handlers.StaticFileHandler(frontend.BrowserFS()))

	return router
}

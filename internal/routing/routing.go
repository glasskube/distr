package routing

import (
	"net/http"
	"time"

	"github.com/glasskube/distr/internal/auth"
	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/frontend"
	"github.com/glasskube/distr/internal/handlers"
	"github.com/glasskube/distr/internal/mail"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/oidc"
	"github.com/glasskube/distr/internal/tracers"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewRouter(
	logger *zap.Logger, db *pgxpool.Pool, mailer mail.Mailer, tracers *tracers.Tracers, oidcer *oidc.OIDCer,
) http.Handler {
	router := chi.NewRouter()
	router.Use(
		// Handles panics
		chimiddleware.Recoverer,
		// Reject bodies larger than 1MiB
		chimiddleware.RequestSize(1048576),
	)
	router.Mount("/api", ApiRouter(logger, db, mailer, tracers, oidcer))
	router.Mount("/internal", InternalRouter())
	router.Mount("/.well-known", WellKnownRouter())
	router.Mount("/", FrontendRouter())
	return router
}

func ApiRouter(
	logger *zap.Logger, db *pgxpool.Pool, mailer mail.Mailer, tracers *tracers.Tracers, oidcer *oidc.OIDCer,
) http.Handler {
	r := chi.NewRouter()
	r.Use(
		chimiddleware.RequestID,
		chimiddleware.RealIP,
		middleware.Sentry,
		middleware.LoggerCtxMiddleware(logger),
		middleware.LoggingMiddleware,
		middleware.ContextInjectorMiddleware(db, mailer, oidcer),
	)

	r.Route("/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(
				middleware.OTEL(tracers.Default()),
			)

			// public routes go here
			r.Group(func(r chi.Router) {
				r.Route("/auth", handlers.AuthRouter)
				r.Route("/webhook", handlers.WebhookRouter)
			})

			// authenticated routes go here
			r.Group(func(r chi.Router) {
				r.Use(
					middleware.SentryUser,
					auth.Authentication.Middleware,
					httprate.Limit(30, 1*time.Second, httprate.WithKeyFuncs(middleware.RateLimitUserIDKey)),
					httprate.Limit(60, 1*time.Minute, httprate.WithKeyFuncs(middleware.RateLimitUserIDKey)),
					httprate.Limit(2000, 1*time.Hour, httprate.WithKeyFuncs(middleware.RateLimitUserIDKey)),

					// TODO (low-prio) in the future, additionally check token audience and require it to be "api"/"user",
					// such that agents cant access anything here (they also can't now, because their tokens will not
					// pass the Authentication chain (DbAuthenticator can't find the user -> 401)
				)
				r.Route("/agent-versions", handlers.AgentVersionsRouter)
				r.Route("/application-licenses", handlers.ApplicationLicensesRouter)
				r.Route("/applications", handlers.ApplicationsRouter)
				r.Route("/artifact-licenses", handlers.ArtifactLicensesRouter)
				r.Route("/artifact-pulls", handlers.ArtifactPullsRouter)
				r.Route("/artifacts", handlers.ArtifactsRouter)
				r.Route("/billing", handlers.BillingRouter)
				r.Route("/context", handlers.ContextRouter)
				r.Route("/customer-organizations", handlers.CustomerOrganizationsRouter)
				r.Route("/dashboard", handlers.DashboardRouter)
				r.Route("/deployment-target-metrics", handlers.DeploymentTargetMetricsRouter)
				r.Route("/deployment-targets", handlers.DeploymentTargetsRouter)
				r.Route("/deployments", handlers.DeploymentsRouter)
				r.Route("/files", handlers.FileRouter)
				r.Route("/organization", handlers.OrganizationRouter)
				r.Route("/organizations", handlers.OrganizationsRouter)
				r.Route("/settings", handlers.SettingsRouter)
				r.Route("/subscription", handlers.SubscriptionRouter)
				r.Route("/tutorial-progress", handlers.TutorialsRouter)
				r.Route("/user-accounts", handlers.UserAccountsRouter)
			})
		})

		// agent connect and download routes go here (authenticated but with accessKeyId and accessKeySecret)
		r.Group(func(r chi.Router) {
			r.Use(
				middleware.OTEL(tracers.Agent()),
			)

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

func WellKnownRouter() http.Handler {
	router := chi.NewRouter()
	if env.WellKnownMicrosoftIdentityAssociation() != nil {
		router.Get("/microsoft-identity-association.json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(env.WellKnownMicrosoftIdentityAssociation())
		})
	}

	return router
}

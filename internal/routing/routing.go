package routing

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/distribution/reference"
	"github.com/docker/distribution"
	drouterv2 "github.com/docker/distribution/registry/api/v2"
	"github.com/go-chi/httprate"
	"github.com/gorilla/mux"

	"github.com/glasskube/distr/internal/auth"
	"github.com/glasskube/distr/internal/frontend"
	"github.com/glasskube/distr/internal/handlers"
	"github.com/glasskube/distr/internal/mail"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewRouter(logger *zap.Logger, db *pgxpool.Pool, mailer mail.Mailer, reg distribution.Namespace) http.Handler {
	router := chi.NewRouter()
	router.Use(
		// Handles panics
		chimiddleware.Recoverer,
		// Reject bodies larger than 1MiB
		chimiddleware.RequestSize(1048576),
	)
	router.Mount("/api", ApiRouter(logger, db, mailer))
	router.Mount("/internal", InternalRouter())
	router.Handle("/v2/*", OCIRouter(reg))
	router.Mount("/", FrontendRouter())
	return router
}

func OCIRouter(reg distribution.Namespace) http.Handler {
	notImplementedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not implemented", http.StatusInternalServerError)
	})
	r := drouterv2.Router()
	r.GetRoute(drouterv2.RouteNameBase).Handler(notImplementedHandler)
	r.GetRoute(drouterv2.RouteNameTags).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if name, err := reference.WithName(mux.Vars(r)["name"]); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if repo, err := reg.Repository(ctx, name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else if tags, err := repo.Tags(ctx).All(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name": name.String(),
				"tags": tags,
			})
		}
	})
	r.GetRoute(drouterv2.RouteNameBlob).Handler(notImplementedHandler)
	r.GetRoute(drouterv2.RouteNameBlobUpload).Handler(notImplementedHandler)
	r.GetRoute(drouterv2.RouteNameBlobUploadChunk).Handler(notImplementedHandler)
	r.GetRoute(drouterv2.RouteNameCatalog).Handler(notImplementedHandler)
	return r
}

func ApiRouter(logger *zap.Logger, db *pgxpool.Pool, mailer mail.Mailer) http.Handler {
	r := chi.NewRouter()
	r.Use(
		chimiddleware.RequestID,
		middleware.Sentry,
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
				httprate.Limit(10, 1*time.Second, httprate.WithKeyFuncs(middleware.RateLimitCurrentUserIdKeyFunc)),
				httprate.Limit(60, 1*time.Minute, httprate.WithKeyFuncs(middleware.RateLimitCurrentUserIdKeyFunc)),
				httprate.Limit(2000, 1*time.Hour, httprate.WithKeyFuncs(middleware.RateLimitCurrentUserIdKeyFunc)),
			)
			r.Route("/applications", handlers.ApplicationsRouter)
			r.Route("/agent-versions", handlers.AgentVersionsRouter)
			r.Route("/deployments", handlers.DeploymentsRouter)
			r.Route("/deployment-targets", handlers.DeploymentTargetsRouter)
			r.Route("/application-licenses", handlers.ApplicationLicensesRouter)
			r.Route("/metrics", handlers.MetricsRouter)
			r.Route("/organization", handlers.OrganizationRouter)
			r.Route("/settings", handlers.SettingsRouter)
			r.Route("/user-accounts", handlers.UserAccountsRouter)
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

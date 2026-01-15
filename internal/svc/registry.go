package svc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"syscall"

	"github.com/distr-sh/distr/internal/buildconfig"
	"github.com/distr-sh/distr/internal/env"
	"github.com/distr-sh/distr/internal/jobs"
	"github.com/distr-sh/distr/internal/mail"
	"github.com/distr-sh/distr/internal/migrations"
	"github.com/distr-sh/distr/internal/oidc"
	"github.com/distr-sh/distr/internal/registry"
	"github.com/distr-sh/distr/internal/routing"
	"github.com/distr-sh/distr/internal/server"
	"github.com/distr-sh/distr/internal/tracers"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Registry struct {
	dbPool            *pgxpool.Pool
	logger            *zap.Logger
	mailer            mail.Mailer
	execDbMigrations  bool
	artifactsRegistry http.Handler
	tracers           *tracers.Tracers
	jobsScheduler     *jobs.Scheduler
	oidcer            *oidc.OIDCer
}

func New(ctx context.Context, options ...RegistryOption) (*Registry, error) {
	var reg Registry
	for _, opt := range options {
		opt(&reg)
	}
	return newRegistry(ctx, &reg)
}

func NewDefault(ctx context.Context) (*Registry, error) {
	var reg Registry
	return newRegistry(ctx, &reg)
}

func newRegistry(ctx context.Context, reg *Registry) (*Registry, error) {
	reg.logger = createLogger()

	reg.logger.Info("initializing service registry",
		zap.String("version", buildconfig.Version()),
		zap.String("commit", buildconfig.Commit()),
		zap.String("edition", buildconfig.Edition()),
		zap.Bool("release", buildconfig.IsRelease()))

	if tracers, err := reg.createTracer(ctx); err != nil {
		return nil, err
	} else {
		reg.tracers = tracers
	}

	if mailer, err := createMailer(ctx); err != nil {
		return nil, err
	} else {
		reg.mailer = mailer
	}

	if reg.execDbMigrations {
		if err := migrations.Up(reg.logger); err != nil {
			return nil, err
		}
	}

	if db, err := reg.createDBPool(ctx); err != nil {
		return nil, err
	} else {
		reg.dbPool = db
	}

	if scheduler, err := reg.createJobsScheduler(); err != nil {
		return nil, err
	} else {
		reg.jobsScheduler = scheduler
	}

	reg.artifactsRegistry = reg.createArtifactsRegistry(ctx)

	if oidcer, err := reg.createOIDCer(ctx, reg.logger); err != nil {
		return nil, err
	} else {
		reg.oidcer = oidcer
	}

	return reg, nil
}

func (r *Registry) Shutdown(ctx context.Context) error {
	if err := r.jobsScheduler.Shutdown(); err != nil {
		r.logger.Warn("job scheduler shutdown failed", zap.Error(err))
	}

	r.logger.Warn("shutting down database connections")
	r.dbPool.Close()

	if err := r.tracers.Shutdown(ctx); err != nil {
		r.logger.Warn("tracer shutdown failed", zap.Error(err))
	}

	// some devices like stdout and stderr can not be synced by the OS
	if err := r.logger.Sync(); err != nil && !errors.Is(err, syscall.EINVAL) {
		return fmt.Errorf("logger sync failed: %w", err)
	}

	return nil
}

func (reg *Registry) createArtifactsRegistry(ctx context.Context) http.Handler {
	return registry.NewDefault(
		ctx,
		reg.GetLogger().With(zap.String("component", "registry")),
		reg.GetDbPool(),
		reg.GetMailer(),
		reg.GetTracers().Registry(),
	)
}

func (r *Registry) GetRouter() http.Handler {
	return routing.NewRouter(
		r.GetLogger(),
		r.GetDbPool(),
		r.GetMailer(),
		r.GetTracers(),
		r.GetOIDCer(),
	)
}

func (r *Registry) GetArtifactsRouter() http.Handler {
	return r.artifactsRegistry
}

func (r *Registry) GetServer() server.Server {
	return server.NewServer(r.GetRouter(), r.logger.With(zap.String("server", "main")))
}

func (r *Registry) GetArtifactsServer() server.Server {
	if env.RegistryEnabled() {
		return server.NewServer(r.GetArtifactsRouter(), r.logger.With(zap.String("server", "registry")))
	} else {
		return server.NewNoop()
	}
}

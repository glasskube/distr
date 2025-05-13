package svc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"syscall"

	"github.com/exaring/otelpgx"
	sentryotel "github.com/getsentry/sentry-go/otel"
	"github.com/glasskube/distr/internal/auth"
	"github.com/glasskube/distr/internal/buildconfig"
	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/mail"
	"github.com/glasskube/distr/internal/mail/noop"
	"github.com/glasskube/distr/internal/mail/ses"
	"github.com/glasskube/distr/internal/mail/smtp"
	"github.com/glasskube/distr/internal/migrations"
	"github.com/glasskube/distr/internal/registry"
	"github.com/glasskube/distr/internal/routing"
	"github.com/glasskube/distr/internal/server"
	"github.com/go-logr/zapr"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	gomail "github.com/wneessen/go-mail"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Registry struct {
	dbPool            *pgxpool.Pool
	logger            *zap.Logger
	mailer            mail.Mailer
	execDbMigrations  bool
	artifactsRegistry http.Handler
	tracer            *trace.TracerProvider
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

	reg.logger.Info("initializing server",
		zap.String("version", buildconfig.Version()),
		zap.String("commit", buildconfig.Commit()),
		zap.Bool("release", buildconfig.IsRelease()))

	if tracer, err := reg.createTracer(ctx); err != nil {
		return nil, err
	} else {
		reg.tracer = tracer
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

	reg.artifactsRegistry = reg.createArtifactsRegistry(ctx)

	return reg, nil
}

func (r *Registry) Shutdown() error {
	r.logger.Warn("shutting down database connections")
	r.dbPool.Close()
	if err := r.tracer.Shutdown(context.TODO()); err != nil {
		r.logger.Warn("tracer shutdown failed", zap.Error(err))
	}
	// some devices like stdout and stderr can not be synced by the OS
	if err := r.logger.Sync(); errors.Is(err, syscall.EINVAL) {
		return nil
	} else {
		return fmt.Errorf("logger sync failed: %w", err)
	}
}

type loggingQueryTracer struct {
	log *zap.Logger
}

var _ pgx.QueryTracer = &loggingQueryTracer{}

func (tracer *loggingQueryTracer) TraceQueryStart(
	ctx context.Context,
	_ *pgx.Conn,
	data pgx.TraceQueryStartData,
) context.Context {
	tracer.log.Debug("executing query", zap.String("sql", data.SQL), zap.Any("args", data.Args))
	return ctx
}

func (tracer *loggingQueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
}

func (reg *Registry) createDBPool(ctx context.Context) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(env.DatabaseUrl())
	if err != nil {
		return nil, err
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		typeNames := []string{"DEPLOYMENT_TYPE", "USER_ROLE", "HELM_CHART_TYPE", "DEPLOYMENT_STATUS_TYPE", "FEATURE",
			"_FEATURE", "TUTORIAL"}
		for _, typeName := range typeNames {
			if pgType, err := conn.LoadType(ctx, typeName); err != nil {
				return err
			} else {
				conn.TypeMap().RegisterType(pgType)
			}
		}
		return nil
	}
	if maxConns := env.DatabaseMaxConns(); maxConns != nil {
		config.MaxConns = int32(*maxConns)
	}
	if env.EnableQueryLogging() {
		config.ConnConfig.Tracer = &loggingQueryTracer{reg.logger}
	} else {
		config.ConnConfig.Tracer = otelpgx.NewTracer(otelpgx.WithTracerProvider(reg.tracer))
	}
	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("cannot set up db pool: %w", err)
	} else if conn, err := db.Acquire(ctx); err != nil {
		// this actually checks whether the DB can be connected to
		return nil, fmt.Errorf("cannot acquire connection: %w", err)
	} else {
		conn.Release()
		return db, nil
	}
}

func (r *Registry) GetDbPool() *pgxpool.Pool {
	return r.dbPool
}

func createMailer(ctx context.Context) (mail.Mailer, error) {
	config := env.GetMailerConfig()
	authOrgOverrideFromAddress :=
		func(ctx context.Context, mail mail.Mail) string {
			if auth, err := auth.Authentication.Get(ctx); err == nil {
				if org := auth.CurrentOrg(); org != nil && org.EmailFromAddress != nil {
					return *org.EmailFromAddress
				}
			}
			return ""
		}
	switch config.Type {
	case env.MailerTypeSMTP:
		smtpConfig := smtp.Config{
			MailerConfig: mail.MailerConfig{
				FromAddressSrc: []mail.FromAddressSrcFn{
					mail.MailOverrideFromAddress(),
					authOrgOverrideFromAddress,
					mail.StaticFromAddress(config.FromAddress.String()),
				},
			},
			Host:      config.SmtpConfig.Host,
			Port:      config.SmtpConfig.Port,
			Username:  config.SmtpConfig.Username,
			Password:  config.SmtpConfig.Password,
			TLSPolicy: gomail.TLSOpportunistic,
		}
		return smtp.New(smtpConfig)
	case env.MailerTypeSES:
		sesConfig := ses.Config{
			MailerConfig: mail.MailerConfig{
				FromAddressSrc: []mail.FromAddressSrcFn{
					mail.MailOverrideFromAddress(),
					authOrgOverrideFromAddress,
					mail.StaticFromAddress(config.FromAddress.String()),
				},
			},
		}
		return ses.NewFromContext(ctx, sesConfig)
	case env.MailerTypeUnspecified:
		return noop.New(), nil
	default:
		return nil, errors.New("invalid mailer type")
	}
}

func (reg *Registry) createArtifactsRegistry(ctx context.Context) http.Handler {
	logger := reg.logger.With(zap.String("component", "registry"))
	return registry.NewDefault(ctx, logger, reg.dbPool, reg.mailer, reg.tracer)
}

func (r *Registry) GetMailer() mail.Mailer {
	return r.mailer
}

func createLogger() *zap.Logger {
	if buildconfig.IsRelease() {
		config := zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		return zap.Must(config.Build())
	} else {
		return zap.Must(zap.NewDevelopment())
	}
}

func (reg *Registry) createTracer(ctx context.Context) (*trace.TracerProvider, error) {
	otel.SetLogger(zapr.NewLogger(reg.logger))

	var tpopts []trace.TracerProviderOption
	if env.OtelExporterOtlpEnabled() {
		if exp, err := otlptracegrpc.New(ctx); err != nil {
			return nil, err
		} else {
			tpopts = append(tpopts, trace.WithSpanProcessor(trace.NewBatchSpanProcessor(exp)))
		}
	}
	if env.OtelExporterSentryEnabled() {
		tpopts = append(tpopts, trace.WithSpanProcessor(sentryotel.NewSentrySpanProcessor()))
	}
	tp := trace.NewTracerProvider(tpopts...)
	otel.SetTracerProvider(tp)

	tmps := []propagation.TextMapPropagator{propagation.TraceContext{}, propagation.Baggage{}}
	if env.OtelExporterSentryEnabled() {
		tmps = append(tmps, sentryotel.NewSentryPropagator())
	}
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(tmps...))

	return tp, nil
}

func (r *Registry) GetLogger() *zap.Logger {
	return r.logger
}

func (r *Registry) GetRouter() http.Handler {
	return routing.NewRouter(r.logger, r.dbPool, r.mailer, r.tracer)
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

func (r *Registry) GetTracer() *trace.TracerProvider {
	return r.tracer
}

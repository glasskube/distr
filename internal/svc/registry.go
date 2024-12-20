package svc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"syscall"

	"github.com/glasskube/cloud/internal/env"
	"github.com/glasskube/cloud/internal/mail"
	"github.com/glasskube/cloud/internal/mail/ses"
	"github.com/glasskube/cloud/internal/mail/smtp"
	"github.com/glasskube/cloud/internal/routing"
	"github.com/glasskube/cloud/internal/server"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	gomail "github.com/wneessen/go-mail"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Registry struct {
	dbPool *pgxpool.Pool
	logger *zap.Logger
	mailer mail.Mailer
}

func NewDefault(ctx context.Context) (*Registry, error) {
	s := &Registry{
		logger: createLogger(),
	}

	s.logger.Info("initializing server")

	if mailer, err := createMailer(ctx); err != nil {
		return nil, err
	} else {
		s.mailer = mailer
	}

	if db, err := createDBPool(ctx, s.logger); err != nil {
		return nil, err
	} else {
		s.dbPool = db
	}

	return s, nil
}

func (r *Registry) Shutdown() error {
	r.logger.Warn("shutting down database connections")
	r.dbPool.Close()
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

func createDBPool(ctx context.Context, log *zap.Logger) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(env.DatabaseUrl())
	if err != nil {
		return nil, err
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		typeNames := []string{"DEPLOYMENT_TYPE", "USER_ROLE"}
		if pgTypes, err := conn.LoadTypes(ctx, typeNames); err != nil {
			return err
		} else {
			conn.TypeMap().RegisterTypes(pgTypes)
			return nil
		}
	}
	if env.EnableQueryLogging() {
		config.ConnConfig.Tracer = &loggingQueryTracer{log}
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
	switch config.Type {
	case env.MailerTypeSMTP:
		smtpConfig := smtp.Config{
			MailerConfig: mail.MailerConfig{
				FromAddress: config.FromAddress,
			},
			Host:      config.SmtpConfig.Host,
			Port:      config.SmtpConfig.Port,
			Username:  config.SmtpConfig.Username,
			Password:  config.SmtpConfig.Password,
			TLSPolicy: gomail.TLSOpportunistic,
		}
		return smtp.New(smtpConfig)
	case env.MailerTypeSES:
		sesConfig := ses.Config{MailerConfig: mail.MailerConfig{FromAddress: config.FromAddress}}
		return ses.NewFromContext(ctx, sesConfig)
	default:
		return nil, errors.New("invalid mailer type")
	}
}

func (r *Registry) GetMailer() mail.Mailer {
	return r.mailer
}

func createLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return zap.Must(config.Build())
}

func (r *Registry) GetLogger() *zap.Logger {
	return r.logger
}

func (r *Registry) GetRouter() http.Handler {
	return routing.NewRouter(r.logger, r.dbPool, r.mailer)
}

func (r *Registry) GetServer() server.Server {
	return *server.NewServer(r.GetRouter(), r.logger)
}

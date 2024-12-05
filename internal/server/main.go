package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/glasskube/cloud/internal/env"
	"github.com/glasskube/cloud/internal/mail"
	"github.com/glasskube/cloud/internal/mail/ses"
	"github.com/glasskube/cloud/internal/mail/smtp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	gomail "github.com/wneessen/go-mail"
	"go.uber.org/zap"

	"go.uber.org/zap/zapcore"
)

type server struct {
	dbPool *pgxpool.Pool
	logger *zap.Logger
	mailer mail.Mailer
}

func New(ctx context.Context) (*server, error) {
	s := &server{
		logger: createLogger(),
	}

	s.logger.Info("initializing server")

	if mailer, err := createMailer(ctx); err != nil {
		return nil, err
	} else {
		s.mailer = mailer
	}

	if db, err := createDBPool(ctx); err != nil {
		return nil, err
	} else {
		s.dbPool = db
	}

	return s, nil
}

func (s *server) Shutdown() error {
	s.logger.Warn("server is shutting down")
	s.dbPool.Close()
	return s.logger.Sync()
}

func createDBPool(ctx context.Context) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(env.DatabaseUrl())
	if err != nil {
		return nil, err
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		typeNames := []string{"DEPLOYMENT_TYPE"}
		if pgTypes, err := conn.LoadTypes(ctx, typeNames); err != nil {
			return err
		} else {
			conn.TypeMap().RegisterTypes(pgTypes)
			return nil
		}
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

func (s *server) GetDbPool() *pgxpool.Pool {
	return s.dbPool
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

func (s *server) GetMailer() mail.Mailer {
	return s.mailer
}

func createLogger() *zap.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
	}

	return zap.Must(config.Build())
}

func (s *server) GetLogger() *zap.Logger {
	return s.logger
}

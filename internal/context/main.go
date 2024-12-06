package context

import (
	"context"

	"github.com/glasskube/cloud/internal/mail"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type contextKey int

const (
	ctxKeyDb contextKey = iota
	ctxKeyLogger
	ctxKeyMailer
	ctxKeyApplication
	ctxKeyDeployment
	ctxKeyDeploymentTarget
)

func GetDb(ctx context.Context) *pgxpool.Pool {
	val := ctx.Value(ctxKeyDb)
	if db, ok := val.(*pgxpool.Pool); ok {
		if db != nil {
			return db
		}
	}
	panic("db not contained in context")
}

func WithDb(ctx context.Context, db *pgxpool.Pool) context.Context {
	ctx = context.WithValue(ctx, ctxKeyDb, db)
	return ctx
}

func GetLogger(ctx context.Context) *zap.Logger {
	val := ctx.Value(ctxKeyLogger)
	if logger, ok := val.(*zap.Logger); ok {
		if logger != nil {
			return logger
		}
	}
	panic("logger not contained in context")
}

func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	ctx = context.WithValue(ctx, ctxKeyLogger, logger)
	return ctx
}

func GetMailer(ctx context.Context) mail.Mailer {
	if mailer, ok := ctx.Value(ctxKeyMailer).(mail.Mailer); ok {
		if mailer != nil {
			return mailer
		}
	}
	panic("logger not contained in context")
}

func WithMailer(ctx context.Context, mailer mail.Mailer) context.Context {
	return context.WithValue(ctx, ctxKeyMailer, mailer)
}

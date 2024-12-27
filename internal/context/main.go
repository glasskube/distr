package context

import (
	"context"

	"github.com/glasskube/cloud/internal/db/queryable"
	"github.com/glasskube/cloud/internal/mail"
	"go.uber.org/zap"
)

type contextKey int

const (
	ctxKeyDb contextKey = iota
	ctxKeyLogger
	ctxKeyMailer
	ctxKeyOrgId
	ctxKeyApplication
	ctxKeyDeployment
	ctxKeyDeploymentTarget
	ctxKeyUserAccount
)

func GetDb(ctx context.Context) queryable.Queryable {
	val := ctx.Value(ctxKeyDb)
	if db, ok := val.(queryable.Queryable); ok {
		if db != nil {
			return db
		}
	}
	panic("db not contained in context")
}

func WithDb(ctx context.Context, db queryable.Queryable) context.Context {
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
	panic("mailer not contained in context")
}

func WithMailer(ctx context.Context, mailer mail.Mailer) context.Context {
	return context.WithValue(ctx, ctxKeyMailer, mailer)
}

package context

import (
	"context"

	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type contextKey int

const (
	ctxKeyDb contextKey = iota
	ctxKeyLogger
	ctxKeyApplication
)

func GetDbOrPanic(ctx context.Context) *pgxpool.Pool {
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

func GetLoggerOrPanic(ctx context.Context) *zap.Logger {
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

func GetApplicationOrPanic(ctx context.Context) *types.Application {
	val := ctx.Value(ctxKeyApplication)
	if application, ok := val.(*types.Application); ok {
		if application != nil {
			return application
		}
	}
	panic("application not contained in context")
}

func WithApplication(ctx context.Context, application *types.Application) context.Context {
	ctx = context.WithValue(ctx, ctxKeyApplication, application)
	return ctx
}

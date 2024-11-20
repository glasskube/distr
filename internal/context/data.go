package context

import (
	"context"
	"github.com/glasskube/cloud/internal/types"
)

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

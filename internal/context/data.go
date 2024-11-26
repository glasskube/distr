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

func WithDeploymentTarget(ctx context.Context, dt *types.DeploymentTarget) context.Context {
	return context.WithValue(ctx, ctxKeyDeploymentTarget, dt)
}

func GetDeploymentTarget(ctx context.Context) *types.DeploymentTarget {
	val := ctx.Value(ctxKeyDeploymentTarget)
	if dt, ok := val.(*types.DeploymentTarget); ok {
		if dt != nil {
			return dt
		}
	}
	panic("deployment target not contained in context")
}

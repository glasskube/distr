package context

import (
	"context"

	"github.com/glasskube/cloud/internal/types"
)

func GetApplication(ctx context.Context) *types.Application {
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

func GetDeployment(ctx context.Context) *types.Deployment {
	val := ctx.Value(ctxKeyDeployment)
	if deployment, ok := val.(*types.Deployment); ok {
		if deployment != nil {
			return deployment
		}
	}
	panic("deployment not contained in context")
}

func WithDeployment(ctx context.Context, deployment *types.Deployment) context.Context {
	ctx = context.WithValue(ctx, ctxKeyDeployment, deployment)
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

func WithOrgId(ctx context.Context, orgId string) context.Context {
	return context.WithValue(ctx, ctxKeyOrgId, orgId)
}

func GetOrgId(ctx context.Context) string {
	val := ctx.Value(ctxKeyOrgId)
	if dt, ok := val.(string); ok {
		return dt
	}
	panic("deployment target not contained in context")
}

func WithUserAccount(ctx context.Context, userAccount *types.UserAccount) context.Context {
	return context.WithValue(ctx, ctxKeyUserAccount, userAccount)
}

func GetUserAccount(ctx context.Context) *types.UserAccount {
	if userAccount, ok := ctx.Value(ctxKeyUserAccount).(*types.UserAccount); ok {
		return userAccount
	}
	panic("no UserAccount found in context")
}

package notification

import (
	"context"

	"github.com/distr-sh/distr/internal/types"
)

func ShouldSendDeploymentStatusNotification(
	previousStatus types.DeploymentRevisionStatus,
	currentStatus *types.DeploymentRevisionStatus,
) bool {
	if currentStatus == nil {
		return previousStatus.IsStale()
	}
	return currentStatus.Type == types.DeploymentStatusTypeError && previousStatus.Type != types.DeploymentStatusTypeError
}

func DispatchDeploymentStatusNotification(
	ctx context.Context,
	deploymentTarget types.DeploymentTargetWithCreatedBy,
	deployment types.DeploymentWithLatestRevision,
	previousStatus types.DeploymentRevisionStatus,
	currentStatus *types.DeploymentRevisionStatus,
) error {
	if !ShouldSendDeploymentStatusNotification(previousStatus, currentStatus) {
		return nil
	}

	go func(ctx context.Context) {
		// Send deployment status notification asynchronously
		// ...
	}(context.WithoutCancel(ctx))

	return nil
}

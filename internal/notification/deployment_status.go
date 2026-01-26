package notification

import (
	"context"

	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/db"
	"github.com/distr-sh/distr/internal/types"
	"go.uber.org/zap"
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

	log := internalctx.GetLogger(ctx)

	go func(ctx context.Context) {
		configs, err := db.GetDeploymentStatusNotificationConfigurationsForDeploymentTarget(ctx, deploymentTarget.ID)
		if err != nil {
			log.Error("failed to get configs", zap.Error(err))
			return
		}

		for _, config := range configs {
			if !config.Enabled {
				continue
			}

			log := log.With(zap.Stringer("config_id", config.ID))

			skip, err := db.ExistsNotificationRecord(ctx, config.ID, previousStatus.ID)
			if err != nil {
				log.Error("failed to check notification record", zap.Error(err))
				continue
			}

			if skip {
				log.Info("skip notification")
				continue
			}

			var aggErr error
			for _, user := range config.UserAccounts {
				log := log.With(zap.Stringer("user_id", user.ID))
				log.Info("send notification")

				// TODO: Send email to users
			}

			record := types.NotificationRecord{
				DeploymentStatusNotificationConfigurationID: &config.ID,
				PreviousDeploymentStatusID:                  &previousStatus.ID,
			}

			if currentStatus != nil {
				record.CurrentDeploymentStatusID = &currentStatus.ID
			}

			if aggErr != nil {
				record.Message = aggErr.Error()
			} else {
				record.Message = "ok"
			}

			if err := db.SaveNotificationRecord(ctx, &record); err != nil {
				log.Error("failed to save notification record", zap.Error(err))
			}
		}
	}(context.WithoutCancel(ctx))

	return nil
}

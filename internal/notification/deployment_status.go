package notification

import (
	"context"
	"errors"
	"fmt"

	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/db"
	"github.com/distr-sh/distr/internal/mail"
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
	mailer := internalctx.GetMailer(ctx)

	// TODO: use a background context populated with the necessary services
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
				aggErr = errors.Join(aggErr, mailer.Send(ctx, mail.New(
					mail.Subject("Deployment Status Notification"),
					mail.TextBody(fmt.Sprintf(`Deployment status has changed:
 * Deployment Target: %v
 * Application: %v
 * Timestamp: %v
 * Message: %v`,
						deploymentTarget.Name, deployment.ApplicationName, currentStatus.CreatedAt, currentStatus.Message)),
					mail.To(user.Email),
				)))
			}

			record := types.NotificationRecord{
				DeploymentStatusNotificationConfigurationID: &config.ID,
				PreviousDeploymentRevisionStatusID:          &previousStatus.ID,
			}

			if currentStatus != nil {
				record.CurrentDeploymentRevisionStatusID = &currentStatus.ID
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

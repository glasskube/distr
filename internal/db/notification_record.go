package db

import (
	"context"
	"fmt"

	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func SaveNotificationRecord(ctx context.Context, record *types.NotificationRecord) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`WITH inserted AS (
			INSERT INTO NotificationRecord (
				deployment_status_notification_configuration_id,
				previous_deployment_revision_status_id,
				current_deployment_revision_status_id,
				message
			)
			VALUES (
				@deploymentStatusNotificationConfigurationID,
				@previousDeploymentStatusID,
				@currentDeploymentStatusID,
				@message
			)
			RETURNING *
		)
		SELECT * FROM inserted`,
		pgx.NamedArgs{
			"deploymentStatusNotificationConfigurationID": record.DeploymentStatusNotificationConfigurationID,
			"previousDeploymentStatusID":                  record.PreviousDeploymentRevisionStatusID,
			"currentDeploymentStatusID":                   record.CurrentDeploymentRevisionStatusID,
			"message":                                     record.Message,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to save NotificationRecord: %w", err)
	}

	if result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.NotificationRecord]); err != nil {
		return fmt.Errorf("failed to collect NotificationRecord: %w", err)
	} else {
		*record = result
	}

	return nil
}

func ExistsNotificationRecord(ctx context.Context, configID, previousID uuid.UUID) (bool, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`SELECT exists(
			SELECT 1
			FROM NotificationRecord
			WHERE deployment_status_notification_configuration_id = @deploymentStatusNotificationConfigurationID
				AND previous_deployment_revision_status_id = @previousDeploymentStatusID
		)`,
		pgx.NamedArgs{
			"deploymentStatusNotificationConfigurationID": configID,
			"previousDeploymentStatusID":                  previousID,
		},
	)
	if err != nil {
		return false, fmt.Errorf("failed to query NotificationRecord exists: %w", err)
	}

	if exists, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[bool]); err != nil {
		return false, fmt.Errorf("failed to collect NotificationRecord exists: %w", err)
	} else {
		return exists, nil
	}
}

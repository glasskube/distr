package db

import (
	"context"
	"fmt"

	"github.com/distr-sh/distr/internal/apierrors"
	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	deploymentStatusNotificationConfigurationOutputExpr = `
	c.id,
	c.created_at,
	c.organization_id,
	c.customer_organization_id,
	c.name,
	c.enabled,
	(
		SELECT array_agg(dt.id)
		FROM DeploymentTarget dt
		WHERE exists(
			SELECT 1 FROM DeploymentStatusNotificationConfiguration_DeploymentTarget j
			WHERE j.deployment_status_notification_configuration_id = c.id
		)
	) AS deployment_target_ids,
	(
		SELECT array_agg(ua.id)
		FROM UserAccount ua
		WHERE exists(
			SELECT 1 FROM DeploymentStatusNotificationConfiguration_Organization_UserAccount j
			WHERE j.deployment_status_notification_configuration_id = c.id
				AND j.user_account_id = ua.id
		)
	) AS user_account_ids,
	(
		SELECT array_agg(row(` + userAccountOutputExpr + `))
		FROM UserAccount ua
		WHERE exists(
			SELECT 1 FROM DeploymentStatusNotificationConfiguration_Organization_UserAccount j
			WHERE j.deployment_status_notification_configuration_id = c.id
				AND j.user_account_id = ua.id
		)
	) AS user_accounts
	`
)

func GetDeploymentStatusNotificationConfigurationsForDeploymentTarget(
	ctx context.Context,
	deploymentTargetID uuid.UUID,
) ([]types.DeploymentStatusNotificationConfiguration, error) {
	db := internalctx.GetDb(ctx)

	rows, err := db.Query(
		ctx,
		`SELECT `+deploymentStatusNotificationConfigurationOutputExpr+`
		FROM DeploymentStatusNotificationConfiguration c
		WHERE exists(
			SELECT 1 FROM DeploymentStatusNotificationConfiguration_DeploymentTarget
			WHERE deployment_status_notification_configuration_id = id
				AND deployment_target_id = :deployment_target_id
		)`,
		pgx.NamedArgs{
			"deployment_target_id": deploymentTargetID,
		},
	)
	if err != nil {
		return nil, err
	}

	result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.DeploymentStatusNotificationConfiguration])
	if err != nil {
		return nil, err
	}

	return result, nil
}

func CreateDeploymentStatusNotificationConfiguration(
	ctx context.Context,
	config *types.DeploymentStatusNotificationConfiguration,
) error {
	panic("not implemented")
}

func UpdateDeploymentStatusNotificationConfiguration(
	ctx context.Context,
	config *types.DeploymentStatusNotificationConfiguration,
) error {
	panic("not implemented")
}

func DeleteDeploymentStatusNotificationConfiguration(ctx context.Context, id uuid.UUID) error {
	db := internalctx.GetDb(ctx)

	cmd, err := db.Exec(
		ctx,
		`DELETE FROM DeploymentStatusNotificationConfiguration WHERE id = :id`,
		pgx.NamedArgs{"id": id},
	)

	if err == nil && cmd.RowsAffected() == 0 {
		err = apierrors.ErrNotFound
	}

	if err != nil {
		return fmt.Errorf("failed to delete DeploymentStatusNotificationConfiguration: %w", err)
	}

	return nil
}

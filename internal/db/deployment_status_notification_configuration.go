package db

import (
	"context"

	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func GetDeploymentStatusNotificationConfigurationsForDeploymentTarget(
	ctx context.Context,
	deploymentTargetID uuid.UUID,
) ([]types.DeploymentStatusNotificationConfiguration, error) {
	db := internalctx.GetDb(ctx)

	rows, err := db.Query(
		ctx,
		`SELECT
			id,
			created_at,
			organization_id,
			customer_organization_id,
			name,
			enabled,
			(
				SELECT array_agg(row(`+userAccountOutputExpr+`))
				FROM UserAccount ua
				WHERE exists(
					SELECT 1 FROM DeploymentStatusNotificationConfiguration_Organization_UserAccount j
					WHERE j.deployment_status_notification_configuration_id = c.id
						AND j.user_account_id = ua.id
				)
			) AS user_accounts
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
	panic("not implemented")
}

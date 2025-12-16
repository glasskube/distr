package db

import (
	"context"
	"fmt"

	"github.com/glasskube/distr/internal/apierrors"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const secretOutputExpr = `
	s.id,
	s.created_at,
	s.updated_at,
	s.updated_by_useraccount_id,
	s.organization_id,
	s.customer_organization_id,
	s.key,
	s.value`

const secretWithUpdatedByOutputExpr = secretOutputExpr + `,
	CASE WHEN u.id IS NULL
		THEN NULL
		ELSE (` + userAccountOutputExpr + `)
	END AS updated_by`

func GetSecrets(
	ctx context.Context,
	organizationID uuid.UUID,
	customerOrganizationID *uuid.UUID,
) ([]types.SecretWithUpdatedBy, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`SELECT `+secretWithUpdatedByOutputExpr+` FROM Secret s
		LEFT JOIN UserAccount u
			ON s.updated_by_useraccount_id = u.id
		WHERE s.organization_id = @organization_id
			AND (@is_vendor OR s.customer_organization_id = @customer_organization_id)`,
		pgx.NamedArgs{
			"organization_id":          organizationID,
			"customer_organization_id": customerOrganizationID,
			"is_vendor":                customerOrganizationID == nil,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query Secret: %w", err)
	}

	if secrets, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.SecretWithUpdatedBy]); err != nil {
		return nil, fmt.Errorf("failed to collect Secret: %w", err)
	} else {
		return secrets, nil
	}
}

func CreateOrUpdateSecret(
	ctx context.Context,
	key, value string,
	organizationID uuid.UUID,
	customerOrganizationID *uuid.UUID,
	updatedByUserAccountID uuid.UUID,
) (*types.SecretWithUpdatedBy, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`WITH inserted AS (
			INSERT INTO Secret (key, value, organization_id, customer_organization_id, updated_by_useraccount_id)
			VALUES (@key, @value, @organization_id, @customer_organization_id, @updated_by_useraccount_id)
			ON CONFLICT (organization_id, customer_organization_id, key) DO UPDATE SET
				value = @value,
				updated_at = now(),
				updated_by_useraccount_id = @updated_by_useraccount_id
			RETURNING *
		)
		SELECT `+secretWithUpdatedByOutputExpr+` FROM inserted s
		LEFT JOIN UserAccount u
			ON s.updated_by_useraccount_id = u.id
		`,
		pgx.NamedArgs{
			"key":                       key,
			"value":                     value,
			"organization_id":           organizationID,
			"customer_organization_id":  customerOrganizationID,
			"updated_by_useraccount_id": updatedByUserAccountID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query Secret: %w", err)
	}

	if secret, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.SecretWithUpdatedBy]); err != nil {
		return nil, fmt.Errorf("failed to collect Secret: %w", err)
	} else {
		return &secret, nil
	}
}

func DeleteSecret(ctx context.Context, key string, organizationID uuid.UUID, customerOrganizationID *uuid.UUID) error {
	db := internalctx.GetDb(ctx)
	cmd, err := db.Exec(
		ctx,
		`DELETE FROM Secret
		WHERE key = @key
			AND organization_id = @organization_id
			AND (@is_vendor OR  customer_organization_id = @customer_organization_id)`,
		pgx.NamedArgs{
			"key":                      key,
			"organization_id":          organizationID,
			"customer_organization_id": customerOrganizationID,
			"is_vendor":                customerOrganizationID == nil,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to delete Secret: %w", err)
	} else if cmd.RowsAffected() == 0 {
		return fmt.Errorf("failed to delete Secret: %w", apierrors.ErrNotFound)
	} else {
		return nil
	}
}

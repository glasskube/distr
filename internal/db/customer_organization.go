package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/glasskube/distr/internal/apierrors"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	customerOrganizationOutputExpr = `
		co.id,
		co.created_at,
		co.organization_id,
		co.image_id,
		co.name
	`
)

func CreateCustomerOrganization(ctx context.Context, customerOrg *types.CustomerOrganization) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"INSERT INTO CustomerOrganization AS co (organization_id, image_id, name) "+
			"VALUES (@organizationId, @imageId, @name) RETURNING "+customerOrganizationOutputExpr,
		pgx.NamedArgs{
			"organizationId": customerOrg.OrganizationID,
			"imageId":        customerOrg.ImageID,
			"name":           customerOrg.Name,
		},
	)
	if err != nil {
		return fmt.Errorf("could not insert CustomerOrganization: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.CustomerOrganization])
	if err != nil {
		return err
	} else {
		*customerOrg = result
		return nil
	}
}

func GetCustomerOrganizationByID(
	ctx context.Context,
	id uuid.UUID,
) (*types.CustomerOrganization, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"SELECT "+customerOrganizationOutputExpr+
			" FROM CustomerOrganization co WHERE co.id = @id ORDER BY co.name",
		pgx.NamedArgs{"id": id},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query CustomerOrganization: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.CustomerOrganization])
	if err != nil {
		return nil, fmt.Errorf("could not collect CustomerOrganization: %w", err)
	}
	return &result, nil
}

func GetCustomerOrganizationsByOrganizationID(
	ctx context.Context,
	orgID uuid.UUID,
) ([]types.CustomerOrganizationWithUserCount, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		fmt.Sprintf(
			`SELECT %v, count(oua.user_account_id) as user_count
			FROM CustomerOrganization co
			LEFT JOIN Organization_UserAccount oua ON co.id = oua.customer_organization_id
			WHERE co.organization_id = @orgId
			GROUP BY %v
			ORDER BY co.name`,
			customerOrganizationOutputExpr, customerOrganizationOutputExpr,
		),
		pgx.NamedArgs{"orgId": orgID},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query CustomerOrganization: %w", err)
	}
	result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.CustomerOrganizationWithUserCount])
	if err != nil {
		return nil, fmt.Errorf("could not collect CustomerOrganization: %w", err)
	}
	return result, nil
}

func UpdateCustomerOrganization(ctx context.Context, customerOrg *types.CustomerOrganization) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"UPDATE CustomerOrganization AS co SET name = @name, image_id = @imageId "+
			"WHERE co.id = @id AND co.organization_id = @organizationId RETURNING "+customerOrganizationOutputExpr,
		pgx.NamedArgs{
			"id":             customerOrg.ID,
			"organizationId": customerOrg.OrganizationID,
			"name":           customerOrg.Name,
			"imageId":        customerOrg.ImageID,
		},
	)
	if err != nil {
		return fmt.Errorf("could not update CustomerOrganization: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.CustomerOrganization])
	if errors.Is(err, pgx.ErrNoRows) {
		return apierrors.ErrNotFound
	} else if err != nil {
		return err
	} else {
		*customerOrg = result
		return nil
	}
}

func DeleteCustomerOrganizationWithID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	db := internalctx.GetDb(ctx)
	cmd, err := db.Exec(
		ctx,
		`DELETE FROM CustomerOrganization WHERE id = @id AND organization_id = @organizationId`,
		pgx.NamedArgs{"id": id, "organizationId": organizationID},
	)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == pgerrcode.ForeignKeyViolation {
			err = fmt.Errorf("%w: %w", apierrors.ErrConflict, err)
		}
		return err
	} else if cmd.RowsAffected() == 0 {
		return apierrors.ErrNotFound
	}
	return nil
}

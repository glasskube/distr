package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/glasskube/distr/internal/apierrors"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/types"
	"github.com/jackc/pgx/v5"
)

func CreateOrganization(ctx context.Context, org *types.Organization) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"INSERT INTO Organization (name) VALUES (@name) RETURNING id, created_at, name, features",
		pgx.NamedArgs{"name": org.Name},
	)
	if err != nil {
		return err
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[types.Organization])
	if err != nil {
		return err
	} else {
		*org = *result
		return nil
	}
}

func GetOrganizationsForUser(ctx context.Context, userId string) ([]*types.OrganizationWithUserRole, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, `
		SELECT o.id, o.created_at, o.name, o.features, j.user_role
			FROM UserAccount u
			INNER JOIN Organization_UserAccount j ON u.id = j.user_account_id
			INNER JOIN Organization o ON o.id = j.organization_id
			WHERE u.id = @id
	`, pgx.NamedArgs{"id": userId})
	if err != nil {
		return nil, err
	}
	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[types.OrganizationWithUserRole])
	if err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func GetOrganizationByID(ctx context.Context, orgId string) (*types.Organization, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"SELECT id, created_at, name, features FROM Organization WHERE id = @id",
		pgx.NamedArgs{"id": orgId},
	)
	if err != nil {
		return nil, err
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.Organization])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	} else if err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func GetOrganizationWithBranding(ctx context.Context, orgId string) (*types.OrganizationWithBranding, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		fmt.Sprintf(
			`SELECT
				o.id, o.created_at, o.name, o.features,
				CASE WHEN b.id IS NOT NULL THEN (%v) END AS branding
			FROM Organization o
			LEFT JOIN OrganizationBranding b ON b.organization_id = o.id
			WHERE o.id = @id`,
			organizationBrandingOutputExpr,
		),
		pgx.NamedArgs{"id": orgId},
	)
	if err != nil {
		return nil, err
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[types.OrganizationWithBranding])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = apierrors.ErrNotFound
		}
		return nil, fmt.Errorf("could not get organization: %w", err)
	} else {
		return result, nil
	}
}

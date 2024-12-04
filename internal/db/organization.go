package db

import (
	"context"
	"errors"

	"github.com/glasskube/cloud/internal/apierrors"
	"github.com/glasskube/cloud/internal/auth"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgx/v5"
)

func CreateOrganization(ctx context.Context, org *types.Organization) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"INSERT INTO Organization (name) VALUES (@name) RETURNING id, created_at, name",
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

func GetOrganizationsForUser(ctx context.Context, userId string) ([]*types.Organization, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, `
		SELECT o.id, o.created_at, o.name
			FROM UserAccount u
			INNER JOIN Organization_UserAccount j ON u.id = j.user_account_id
			INNER JOIN Organization o ON o.id = j.organization_id
			WHERE u.id = @id
	`, pgx.NamedArgs{"id": userId})
	if err != nil {
		return nil, err
	}
	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[types.Organization])
	if err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func GetOrganizationWithID(ctx context.Context, orgId string) (*types.Organization, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"SELECT id, created_at, name FROM Organization WHERE id = @id",
		pgx.NamedArgs{"id": orgId},
	)
	if err != nil {
		return nil, err
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[types.Organization])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	} else if err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

// GetCurrentOrg retrieves the organization_id from the context auth token and returns the corresponding Organization
//
// TODO: this function should probably be moved to another module and maybe support some kind of result caching.
func GetCurrentOrg(ctx context.Context) (*types.Organization, error) {
	if orgId, err := auth.CurrentOrgId(ctx); err != nil {
		return nil, err
	} else if org, err := GetOrganizationWithID(ctx, orgId); err != nil {
		return nil, err
	} else {
		return org, nil
	}
}

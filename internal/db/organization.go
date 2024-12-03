package db

import (
	"context"
	"errors"

	"github.com/glasskube/cloud/internal/auth"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgx/v5"
)

func CreateOrganization(ctx context.Context, org *types.Organization) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, "INSERT INTO Organization (name) VALUES (@name)", pgx.NamedArgs{"name": org.Name})
	if err != nil {
		return err
	}
	if result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[types.Organization]); err != nil {
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
			LEFT JOIN Organization_UserAccount j ON u.id = j.user_account_id
			LEFT JOIN Organization o ON o.id = j.organization_id
			WHERE u.id = @id
	`, pgx.NamedArgs{"id": userId})
	if err != nil {
		return nil, err
	} else if result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[types.Organization]); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func GetOrganizationWithID(ctx context.Context, orgId string) (*types.Organization, error) {
	// TODO: Implement
	return nil, errors.New("not implemented")
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

package db

import (
	"context"
	"errors"

	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgx/v5"
)

func CreateOrganization(ctx context.Context, org *types.Organization) error {
	// TODO: Implement
	return errors.New("not implemented")
}

func GetOrganizationsForUser(ctx context.Context, userId string) ([]*types.Organization, error) {
	// TODO: Implement
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

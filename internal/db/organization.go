package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/glasskube/distr/internal/apierrors"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	organizationOutputExpr = `
		o.id,
		o.created_at,
		o.name,
		o.slug,
		o.features,
		o.app_domain,
		o.registry_domain,
		o.email_from_address
	`
	organizationWithUserRoleOutputExpr = organizationOutputExpr + ", j.user_role, j.created_at as joined_org_at "
)

func CreateOrganization(ctx context.Context, org *types.Organization) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"INSERT INTO Organization AS o (name) VALUES (@name) RETURNING "+organizationOutputExpr,
		pgx.NamedArgs{"name": org.Name},
	)
	if err != nil {
		return fmt.Errorf("could not create orgnization: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[types.Organization])
	if err != nil {
		return err
	} else {
		*org = *result
		return nil
	}
}

func UpdateOrganization(ctx context.Context, org *types.Organization) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"UPDATE Organization AS o SET name = @name, slug = @slug WHERE id = @id RETURNING "+organizationOutputExpr,
		pgx.NamedArgs{"id": org.ID, "name": org.Name, "slug": org.Slug},
	)
	if err != nil {
		return err
	}
	if result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[types.Organization]); err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation {
			err = fmt.Errorf("%w: %w", apierrors.ErrConflict, err)
		}
		return err
	} else {
		*org = *result
		return nil
	}
}

func GetOrganizationsForUser(ctx context.Context, userID uuid.UUID) ([]types.OrganizationWithUserRole, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, `
		SELECT`+organizationWithUserRoleOutputExpr+`
			FROM UserAccount u
			INNER JOIN Organization_UserAccount j ON u.id = j.user_account_id
			INNER JOIN Organization o ON o.id = j.organization_id
			WHERE u.id = @id
			ORDER BY o.created_at
	`, pgx.NamedArgs{"id": userID})
	if err != nil {
		return nil, err
	}
	result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.OrganizationWithUserRole])
	if err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func GetOrganizationByID(ctx context.Context, orgID uuid.UUID) (*types.Organization, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"SELECT "+organizationOutputExpr+" FROM Organization o WHERE id = @id",
		pgx.NamedArgs{"id": orgID},
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

func GetOrganizationWithBranding(ctx context.Context, orgID uuid.UUID) (*types.OrganizationWithBranding, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		fmt.Sprintf(
			`SELECT `+organizationOutputExpr+`,
				CASE WHEN b.id IS NOT NULL THEN (%v) END AS branding
			FROM Organization o
			LEFT JOIN OrganizationBranding b ON b.organization_id = o.id
			WHERE o.id = @id`,
			organizationBrandingOutputExpr,
		),
		pgx.NamedArgs{"id": orgID},
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

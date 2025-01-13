package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/glasskube/cloud/internal/apierrors"

	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgx/v5"
)

const (
	organizationBrandingOutputExpr = `
		b.id, b.created_at, b.organization_id, b.updated_at, b.updated_by_user_account_id, b.title, b.description,
		b.logo, b.logo_file_name, b.logo_content_type
	`
)

func GetOrganizationBranding(ctx context.Context, organizationId string) (*types.OrganizationBranding, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"SELECT"+organizationBrandingOutputExpr+
			"FROM OrganizationBranding b "+
			"WHERE b.organization_id = @organizationId",
		pgx.NamedArgs{"organizationId": organizationId})
	if err != nil {
		return nil, fmt.Errorf("failed to query OrganizationBranding: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.OrganizationBranding])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get OrganizationBranding: %w", err)
	} else {
		return &result, nil
	}
}

func CreateOrganizationBranding(ctx context.Context, b *types.OrganizationBranding) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`INSERT INTO OrganizationBranding AS b
			(organization_id, updated_at, updated_by_user_account_id, title, description,
			 logo, logo_file_name, logo_content_type)
			VALUES (@organization_id, @updated_at, @updated_by_user_account_id, @title, @description,
			        @logo, @logo_file_name, @logo_content_type)
			RETURNING `+organizationBrandingOutputExpr,
		pgx.NamedArgs{
			"organization_id":            b.OrganizationID,
			"updated_at":                 b.UpdatedAt,
			"updated_by_user_account_id": b.UpdatedByUserAccountID,
			"title":                      b.Title,
			"description":                b.Description,
			"logo":                       b.Logo,
			"logo_file_name":             b.LogoFileName,
			"logo_content_type":          b.LogoContentType,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create OrganizationBranding: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.OrganizationBranding])
	if err != nil {
		return fmt.Errorf("could not save OrganizationBranding: %w", err)
	} else {
		*b = result
		return nil
	}
}

func UpdateOrganizationBranding(ctx context.Context, b *types.OrganizationBranding) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`UPDATE OrganizationBranding AS b
		SET updated_at = @updated_at,
			updated_by_user_account_id = @updated_by_user_account_id,
			title = @title,
			description = @description,
			logo = @logo,
			logo_file_name = @logo_file_name,
			logo_content_type = @logo_content_type
		WHERE id = @id
		RETURNING `+organizationBrandingOutputExpr,
		pgx.NamedArgs{
			"id":                         b.ID,
			"updated_at":                 b.UpdatedAt,
			"updated_by_user_account_id": b.UpdatedByUserAccountID,
			"title":                      b.Title,
			"description":                b.Description,
			"logo":                       b.Logo,
			"logo_file_name":             b.LogoFileName,
			"logo_content_type":          b.LogoContentType,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update OrganizationBranding: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.OrganizationBranding])
	if err != nil {
		return fmt.Errorf("could not save OrganizationBranding: %w", err)
	} else {
		*b = result
		return nil
	}
}

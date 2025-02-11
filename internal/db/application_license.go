package db

import (
	"context"
	"fmt"

	"github.com/glasskube/distr/internal/apierrors"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/types"
	"github.com/jackc/pgx/v5"
)

const (
	applicationLicenseOutputExpr = `
		al.id, al.created_at, al.name, al.expires_at, al.application_id, al.organization_id,
		al.owner_useraccount_id, al.registry_url, al.registry_username, al.registry_password
	`
	applicationLicenseWithVersionsOutputExpr = applicationLicenseOutputExpr + `,
		coalesce((
		   	SELECT array_agg(
				row(av.id, av.created_at, av.name, av.chart_type, av.chart_name, av.chart_url, av.chart_version)
				ORDER BY av.created_at ASC
			)
		   	FROM ApplicationLicense_ApplicationVersion alav
				LEFT JOIN applicationversion av ON alav.application_version_id = av.id
		   	WHERE alav.application_license_id = al.id
		   ), array[]::record[]
		) as versions
	`
)

func CreateApplicationLicense(ctx context.Context, license *types.ApplicationLicense) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`INSERT INTO ApplicationLicense AS al (
			name, expires_at, application_id, organization_id, owner_useraccount_id, registry_url, registry_username,
			registry_password
		) VALUES (
			@name, @expiresAt, @applicationId, @organizationId, @ownerUserAccountId, @registryUrl, @registryUsername,
			@registryPassword
		) RETURNING`+applicationLicenseOutputExpr,
		pgx.NamedArgs{
			"name":               license.Name,
			"expiresAt":          license.ExpiresAt,
			"applicationId":      license.ApplicationID,
			"organizationId":     license.OrganizationID,
			"ownerUserAccountId": license.OwnerUserAccountID,
			"registryUrl":        license.RegistryURL,
			"registryUsername":   license.RegistryUsername,
			"registryPassword":   license.RegistryPassword,
		},
	)

	if err == nil {
		return fmt.Errorf("could not insert ApplicationLicense: %w", err)
	}

	if result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.ApplicationLicense]); err != nil {
		return err
	} else {
		*license = result
		return nil
	}
}

func RevokeApplicationLicenseWithID(ctx context.Context, id string) error {
	db := internalctx.GetDb(ctx)
	cmd, err := db.Exec(ctx, "UPDATE ApplicationLicense SET expires_at = now() WHERE id = @id", pgx.NamedArgs{"id": id})
	if err == nil && cmd.RowsAffected() < 1 {
		err = apierrors.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("could not update ApplicationLicense: %w", err)
	} else {
		return nil
	}
}

func AddVersionToApplicationLicense(
	ctx context.Context,
	license *types.ApplicationLicense,
	id string,
) error {
	db := internalctx.GetDb(ctx)
	_, err := db.Exec(
		ctx,
		`INSERT INTO ApplicationLicense_ApplicationVersion (application_version_id, application_license_id)
		VALUES (@applicationVersionId, @applicationLicenseId)
		ON CONFLICT (application_version_id, application_license_id) DO NOTHING`,
		pgx.NamedArgs{
			"applicationVersionId": id,
			"applicationLicenseId": license.ID,
		},
	)
	if err != nil {
		return fmt.Errorf("could not insert relation: %w", err)
	}
	return nil
}

func RemoveVersionFromApplicationLicense(
	ctx context.Context,
	license *types.ApplicationLicense,
	id string,
) error {
	db := internalctx.GetDb(ctx)
	_, err := db.Exec(
		ctx,
		`DELETE FROM ApplicationLicense_ApplicationVersion
		WHERE application_license_id = @applicationLicenseId
		AND application_version_id = @applicationVersionId`,
		pgx.NamedArgs{
			"applicationLicenseId": license.ID,
			"applicationVersionId": id,
		},
	)
	if err != nil {
		return fmt.Errorf("could not delete relation: %w", err)
	} else {
		return nil
	}
}

func GetApplicationLicensesWithOrganizationID(
	ctx context.Context,
	id string,
) ([]types.ApplicationLicenseWithVersions, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		fmt.Sprintf(
			"SELECT %v FROM ApplicationLicense al WHERE al.organization_id = @id",
			applicationLicenseWithVersionsOutputExpr,
		),
		pgx.NamedArgs{"id": id},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query ApplicationLicense: %w", err)
	}

	if result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.ApplicationLicenseWithVersions]); err != nil {
		return nil, fmt.Errorf("could not collect ApplicationLicense: %w", err)
	} else {
		return result, nil
	}
}

func GetApplicationLicensesWithOwnerID(ctx context.Context, id string) ([]types.ApplicationLicenseWithVersions, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		fmt.Sprintf(
			"SELECT %v FROM ApplicationLicense al WHERE al.owner_useraccount_id = @id",
			applicationLicenseWithVersionsOutputExpr,
		),
		pgx.NamedArgs{"id": id},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query ApplicationLicense: %w", err)
	}

	if result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.ApplicationLicenseWithVersions]); err != nil {
		return nil, fmt.Errorf("could not collect ApplicationLicense: %w", err)
	} else {
		return result, nil
	}
}

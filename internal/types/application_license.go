package types

import "time"

type ApplicationLicense struct {
	ID                 string     `db:"id"`
	CreatedAt          time.Time  `db:"created_at"`
	Name               string     `db:"name"`
	ExpiresAt          *time.Time `db:"expires_at"`
	ApplicationID      string     `db:"application_id"`
	OrganizationID     string     `db:"organization_id"`
	OwnerUserAccountID *string    `db:"owner_useraccount_id"`
	RegistryURL        *string    `db:"registry_url"`
	RegistryUsername   *string    `db:"registry_username"`
	RegistryPassword   *string    `db:"registry_password"`
}

type ApplicationLicenseWithVersions struct {
	ApplicationLicense
	Versions []ApplicationVersion `db:"versions"`
}

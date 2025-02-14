package types

import (
	"time"

	"github.com/google/uuid"
)

type ApplicationLicense struct {
	ID                 uuid.UUID  `db:"id"`
	CreatedAt          time.Time  `db:"created_at"`
	Name               string     `db:"name"`
	ExpiresAt          *time.Time `db:"expires_at"`
	ApplicationID      uuid.UUID  `db:"application_id"`
	OrganizationID     uuid.UUID  `db:"organization_id"`
	OwnerUserAccountID *uuid.UUID `db:"owner_useraccount_id"`
	RegistryURL        *string    `db:"registry_url"`
	RegistryUsername   *string    `db:"registry_username"`
	RegistryPassword   *string    `db:"registry_password"`
}

type ApplicationLicenseWithVersions struct {
	ApplicationLicense
	Versions []ApplicationVersion `db:"versions"`
}

func (license *ApplicationLicenseWithVersions) HasVersionWithID(id uuid.UUID) bool {
	for _, v := range license.Versions {
		if v.ID == id {
			return true
		}
	}
	return false
}

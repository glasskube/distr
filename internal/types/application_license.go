package types

import (
	"time"

	"github.com/google/uuid"
)

type ApplicationLicense struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	CreatedAt          time.Time  `db:"created_at" json:"createdAt"`
	Name               string     `db:"name" json:"name"`
	ExpiresAt          *time.Time `db:"expires_at" json:"expiresAt,omitempty"`
	ApplicationID      uuid.UUID  `db:"application_id" json:"applicationId"`
	OrganizationID     uuid.UUID  `db:"organization_id" json:"-"`
	OwnerUserAccountID *uuid.UUID `db:"owner_useraccount_id" json:"ownerUserAccountId,omitempty"`
	RegistryURL        *string    `db:"registry_url" json:"registryUrl,omitempty"`
	RegistryUsername   *string    `db:"registry_username" json:"registryUsername,omitempty"`
	RegistryPassword   *string    `db:"registry_password" json:"registryPassword,omitempty"`
}

type ApplicationLicenseWithVersions struct {
	ApplicationLicense
	Versions []ApplicationVersion `db:"versions" json:"versions"`
}

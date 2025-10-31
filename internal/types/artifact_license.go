package types

import (
	"time"

	"github.com/google/uuid"
)

type ArtifactLicenseBase struct {
	ID                     uuid.UUID  `db:"id" json:"id"`
	CreatedAt              time.Time  `db:"created_at" json:"createdAt"`
	Name                   string     `db:"name" json:"name"`
	ExpiresAt              *time.Time `db:"expires_at" json:"expiresAt,omitempty"`
	OrganizationID         uuid.UUID  `db:"organization_id" json:"-"`
	CustomerOrganizationID *uuid.UUID `db:"customer_organization_id" json:"customerOrganizationId,omitempty"`
}

type ArtifactLicenseSelection struct {
	ArtifactID uuid.UUID   `db:"artifact_id" json:"artifactId"`
	VersionIDs []uuid.UUID `db:"versions" json:"versionIds"`
}

type ArtifactLicense struct {
	ArtifactLicenseBase
	Artifacts []ArtifactLicenseSelection `db:"artifacts" json:"artifacts,omitempty"`
}

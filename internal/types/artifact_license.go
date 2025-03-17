package types

import (
	"time"

	"github.com/google/uuid"
)

type ArtifactLicenseBase struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	CreatedAt          time.Time  `db:"created_at" json:"createdAt"`
	Name               string     `db:"name" json:"name"`
	ExpiresAt          *time.Time `db:"expires_at" json:"expiresAt,omitempty"`
	OrganizationID     uuid.UUID  `db:"organization_id" json:"-"`
	OwnerUserAccountID *uuid.UUID `db:"owner_useraccount_id" json:"ownerUserAccountId,omitempty"`
}

type ArtifactLicenseSelection struct {
	Artifact Artifact                `db:"artifact" json:"artifact"`
	Versions []TaggedArtifactVersion `db:"versions" json:"versions"`
}

type ArtifactLicense struct {
	ArtifactLicenseBase
	Artifacts []ArtifactLicenseSelection `db:"artifacts" json:"artifacts,omitempty"`
	Owner     *UserAccount               `db:"owner" json:"owner,omitempty"`
}

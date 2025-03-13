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
	OwnerUserAccountID *uuid.UUID `db:"owner_useraccount_id" json:"ownerUserAccountID,omitempty"`
}

type ArtifactLicenseSelection struct {
	artifact Artifact                `db:"artifact"`
	tags     []TaggedArtifactVersion `db:"tags"`
}

type ArtifactLicense struct {
	ArtifactLicenseBase
	artifacts []ArtifactLicenseSelection
	Owner     *UserAccount `db:"owner" json:"owner,omitempty"`
}

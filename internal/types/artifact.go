package types

import (
	"time"

	"github.com/google/uuid"
)

type Artifact struct {
	ID             uuid.UUID `db:"id" json:"id"`
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
	OrganizationID uuid.UUID `db:"organization_id" json:"-"`
	Name           string    `db:"name" json:"name"`
}

type ArtifactVersionTag struct {
	ID   uuid.UUID `db:"id" json:"id"`
	Name string    `db:"name" json:"name"`
}

type TaggedArtifactVersion struct {
	ID        uuid.UUID            `db:"id" json:"id"`
	CreatedAt time.Time            `db:"created_at" json:"createdAt"`
	Digest    string               `db:"-" json:"digest"`
	Tags      []ArtifactVersionTag `db:"-" json:"tags"`
}

type ArtifactWithTaggedVersion struct {
	Artifact
	Versions []TaggedArtifactVersion `db:"versions" json:"versions"`
}

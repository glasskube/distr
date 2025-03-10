package types

import (
	"github.com/google/uuid"
	"time"
)

type ArtifactTag struct {
	ID                     uuid.UUID `db:"id" json:"id"`
	CreatedAt              time.Time `db:"created_at" json:"createdAt"`
	CreatedByUserAccountID uuid.UUID `db:"organization_id" json:"-"`
	Hash                   string    `db:"hash" json:"hash"`
	Labels                 []string  `db:"labels" json:"labels"`

	ArtifactID uuid.UUID `db:"artifact_id" json:"artifactId"`
}

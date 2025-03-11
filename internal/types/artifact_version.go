package types

import (
	"github.com/google/uuid"
	"time"
)

type ArtifactVersion struct {
	ID                     uuid.UUID  `db:"id" json:"id"`
	CreatedAt              time.Time  `db:"created_at" json:"createdAt"`
	CreatedByUserAccountID *uuid.UUID `db:"created_by_useraccount_id" json:"-"`
	Name                   string     `db:"name" json:"name"`

	ArtifactID uuid.UUID `db:"artifact_id" json:"artifactId"`
}

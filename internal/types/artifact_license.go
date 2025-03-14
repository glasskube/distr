package types

import (
	"time"

	"github.com/google/uuid"
)

type ArtifactLicense struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`
	Name      string     `db:"name" json:"name"`
	ExpiresAt *time.Time `db:"expires_at" json:"expiresAt,omitempty"`
}

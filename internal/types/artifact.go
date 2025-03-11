package types

import (
	"github.com/google/uuid"
	"time"
)

type Artifact struct {
	ID             uuid.UUID `db:"id" json:"id"`
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
	OrganizationID uuid.UUID `db:"organization_id" json:"-"`
	Name           string    `db:"name" json:"name"`
}

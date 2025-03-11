package types

import (
	"github.com/google/uuid"
	"time"
)

type ArtifactBlob struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	Name      string    `db:"name" json:"name"`
	IsLead    bool      `db:"is_lead" json:"isLead"`
}

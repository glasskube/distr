package types

import (
	"time"

	"github.com/google/uuid"
)

type CustomerOrganization struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	CreatedAt      time.Time  `db:"created_at" json:"createdAt"`
	OrganizationID uuid.UUID  `db:"organization_id" json:"organizationId"`
	ImageID        *uuid.UUID `db:"image_id" json:"imageId,omitempty"`
	Name           string     `db:"name" json:"name"`
}

type CustomerOrganizationWithUserCount struct {
	CustomerOrganization
	UserCount int `db:"user_count" json:"userCount"`
}

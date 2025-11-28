package api

import (
	"time"

	"github.com/google/uuid"
)

type CreateUpdateCustomerOrganizationRequest struct {
	Name    string     `json:"name"`
	ImageID *uuid.UUID `json:"imageId,omitempty"`
}

type CustomerOrganization struct {
	ID        uuid.UUID  `json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	Name      string     `json:"name"`
	ImageID   *uuid.UUID `json:"imageId,omitempty"`
	ImageURL  string     `json:"imageUrl,omitempty"`
}

type CustomerOrganizationWithUserCount struct {
	CustomerOrganization
	UserCount int64 `json:"userCount"`
}

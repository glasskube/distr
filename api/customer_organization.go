package api

import "github.com/google/uuid"

type CreateUpdateCustomerOrganizationRequest struct {
	Name    string     `json:"name"`
	ImageID *uuid.UUID `json:"imageId,omitempty"`
}

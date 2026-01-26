package api

import (
	"time"

	"github.com/distr-sh/distr/internal/types"
	"github.com/google/uuid"
)

type CreateUpdateCustomerOrganizationRequest struct {
	Name     string                              `json:"name"`
	ImageID  *uuid.UUID                          `json:"imageId,omitempty"`
	Features []types.CustomerOrganizationFeature `json:"features,omitempty"`
}

type CustomerOrganization struct {
	ID        uuid.UUID                           `json:"id"`
	CreatedAt time.Time                           `json:"createdAt"`
	Name      string                              `json:"name"`
	ImageID   *uuid.UUID                          `json:"imageId,omitempty"`
	ImageURL  *string                             `json:"imageUrl,omitempty"`
	Features  []types.CustomerOrganizationFeature `json:"features"`
}

type CustomerOrganizationWithUsage struct {
	CustomerOrganization
	UserCount             int64 `json:"userCount"`
	DeploymentTargetCount int64 `json:"deploymentTargetCount"`
}

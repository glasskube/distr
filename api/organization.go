package api

import (
	"github.com/glasskube/distr/internal/types"
)

type CreateUpdateOrganizationRequest struct {
	Name string  `json:"name"`
	Slug *string `json:"slug"`
}

type OrganizationResponse struct {
	Organization       types.Organization
	SubscriptionLimits SubscriptionLimits
}

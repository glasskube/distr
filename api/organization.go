package api

import (
	"github.com/glasskube/distr/internal/types"
)

type CreateUpdateOrganizationRequest struct {
	Name string  `json:"name"`
	Slug *string `json:"slug"`
}

type OrganizationResponse struct {
	types.Organization
	SubscriptionLimits SubscriptionLimits `json:"subscriptionLimits"`
}

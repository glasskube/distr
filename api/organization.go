package api

import (
	"github.com/glasskube/distr/internal/types"
)

type OrganizationResponse struct {
	types.Organization
	SubscriptionLimits SubscriptionLimits `json:"subscriptionLimits"`
}

package api

import (
	"github.com/glasskube/distr/internal/types"
)

type OrganizationResponse struct {
	Organization       types.Organization
	SubscriptionLimits SubscriptionLimits
}

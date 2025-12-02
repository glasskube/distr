package mapping

import (
	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/subscription"
	"github.com/glasskube/distr/internal/types"
)

func OrganizationToAPI(o types.Organization) api.OrganizationResponse {
	return api.OrganizationResponse{
		Organization:       o,
		SubscriptionLimits: subscription.GetSubscriptionLimits(o.SubscriptionType),
	}
}

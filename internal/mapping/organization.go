package mapping

import (
	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/internal/subscription"
	"github.com/distr-sh/distr/internal/types"
)

func OrganizationToAPI(o types.Organization) api.OrganizationResponse {
	return api.OrganizationResponse{
		Organization:       o,
		SubscriptionLimits: subscription.GetSubscriptionLimits(o.SubscriptionType),
	}
}

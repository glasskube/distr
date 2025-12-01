package api

import (
	"time"

	"github.com/glasskube/distr/internal/types"
)

type SubscriptionLimits struct {
	MaxCustomerOrganizations        int64 `json:"maxCustomerOrganizations"`
	MaxUsersPerCustomerOrganization int64 `json:"maxUsersPerCustomerOrganization"`
	MaxDeploymentsPerCustomerOrg    int64 `json:"maxDeploymentsPerCustomerOrganization"`
}

type SubscriptionInfo struct {
	SubscriptionType                       types.SubscriptionType `json:"subscriptionType"`
	SubscriptionEndsAt                     time.Time              `json:"subscriptionEndsAt"`
	SubscriptionExternalID                 *string                `json:"subscriptionExternalId"`
	SubscriptionCustomerOrganizationQty    *int64                 `json:"subscriptionCustomerOrganizationQuantity"`
	SubscriptionUserAccountQty             *int64                 `json:"subscriptionUserAccountQuantity"`
	CurrentUserAccountCount                int64                  `json:"currentUserAccountCount"`
	CurrentCustomerOrganizationCount       int64                  `json:"currentCustomerOrganizationCount"`
	CurrentMaxUsersPerCustomer             int64                  `json:"currentMaxUsersPerCustomer"`
	CurrentMaxDeploymentTargetsPerCustomer int64                  `json:"currentMaxDeploymentTargetsPerCustomer"`
	TrialLimits                            SubscriptionLimits     `json:"trialLimits"`
	StarterLimits                          SubscriptionLimits     `json:"starterLimits"`
	ProLimits                              SubscriptionLimits     `json:"proLimits"`
	EnterpriseLimits                       SubscriptionLimits     `json:"enterpriseLimits"`
}

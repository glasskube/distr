package subscription

import (
	"fmt"

	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/types"
)

type Limit int64

const (
	Unlimited Limit = -1

	MaxCustomersPerOrganizationStarter   Limit = 3
	MaxCustomersPerOrganizationPro       Limit = 100
	MaxCustomersPerOrganizationUnlimited Limit = Unlimited

	MaxUsersPerCustomerOrganizationStarter   Limit = 1
	MaxUsersPerCustomerOrganizationPro       Limit = 10
	MaxUsersPerCustomerOrganizationUnlimited Limit = Unlimited

	MaxDeploymentTargetsPerCustomerOrganizationStarter   Limit = 1
	MaxDeploymentTargetsPerCustomerOrganizationPro       Limit = 3
	MaxDeploymentTargetsPerCustomerOrganizationUnlimited Limit = Unlimited
)

func (l Limit) IsReached(other int64) bool {
	return l != Unlimited && int64(l) <= other
}

func (l Limit) IsExceeded(other int64) bool {
	return l != Unlimited && int64(l) < other
}

func GetCustomersPerOrganizationLimit(st types.SubscriptionType) Limit {
	switch st {
	case types.SubscriptionTypeCommunity:
		return MaxCustomersPerOrganizationUnlimited
	case types.SubscriptionTypeTrial:
		return MaxCustomersPerOrganizationUnlimited
	case types.SubscriptionTypeStarter:
		return MaxCustomersPerOrganizationStarter
	case types.SubscriptionTypePro:
		return MaxCustomersPerOrganizationPro
	case types.SubscriptionTypeEnterprise:
		return MaxCustomersPerOrganizationUnlimited
	default:
		panic(fmt.Sprintf("invalid subscription type: %v", st))
	}
}

func GetUsersPerCustomerOrganizationLimit(st types.SubscriptionType) Limit {
	switch st {
	case types.SubscriptionTypeCommunity:
		return MaxUsersPerCustomerOrganizationStarter
	case types.SubscriptionTypeTrial:
		return MaxUsersPerCustomerOrganizationUnlimited
	case types.SubscriptionTypeStarter:
		return MaxUsersPerCustomerOrganizationStarter
	case types.SubscriptionTypePro:
		return MaxUsersPerCustomerOrganizationPro
	case types.SubscriptionTypeEnterprise:
		return MaxUsersPerCustomerOrganizationUnlimited
	default:
		panic(fmt.Sprintf("invalid subscription type: %v", st))
	}
}

func GetDeploymentTargetsPerCustomerOrganizationLimit(st types.SubscriptionType) Limit {
	switch st {
	case types.SubscriptionTypeCommunity:
		return MaxDeploymentTargetsPerCustomerOrganizationStarter
	case types.SubscriptionTypeTrial:
		return MaxDeploymentTargetsPerCustomerOrganizationUnlimited
	case types.SubscriptionTypeStarter:
		return MaxDeploymentTargetsPerCustomerOrganizationStarter
	case types.SubscriptionTypePro:
		return MaxDeploymentTargetsPerCustomerOrganizationPro
	case types.SubscriptionTypeEnterprise:
		return MaxDeploymentTargetsPerCustomerOrganizationUnlimited
	default:
		panic(fmt.Sprintf("invalid subscription type: %v", st))
	}
}

func GetSubscriptionLimits(st types.SubscriptionType) api.SubscriptionLimits {
	return api.SubscriptionLimits{
		MaxCustomerOrganizations:        int64(GetCustomersPerOrganizationLimit(st)),
		MaxUsersPerCustomerOrganization: int64(GetUsersPerCustomerOrganizationLimit(st)),
		MaxDeploymentsPerCustomerOrg:    int64(GetDeploymentTargetsPerCustomerOrganizationLimit(st)),
	}
}

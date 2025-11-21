package subscription

import (
	"fmt"

	"github.com/glasskube/distr/internal/types"
)

type Limit int

const (
	Unlimited Limit = -1

	MaxCustomersPerOrganizationStarter   Limit = 1
	MaxCustomersPerOrganizationPro       Limit = 50
	MaxCustomersPerOrganizationUnlimited Limit = Unlimited

	MaxUsersPerCustomerOrganizationStarter   Limit = 1
	MaxUsersPerCustomerOrganizationPro       Limit = 10
	MaxUsersPerCustomerOrganizationUnlimited Limit = Unlimited

	MaxDeploymentTargetsPerCustomerOrganizationStarter   Limit = 1
	MaxDeploymentTargetsPerCustomerOrganizationPro       Limit = 3
	MaxDeploymentTargetsPerCustomerOrganizationUnlimited Limit = Unlimited
)

func (l Limit) Check(other int64) bool {
	return l == Unlimited || other <= int64(l)
}

func GetCustomersPerOrganizationLimit(st types.SubscriptionType) Limit {
	switch st {
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

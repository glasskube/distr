package subscription

import (
	"context"
	"fmt"

	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
)

func IsVendorUserAccountLimitReached(ctx context.Context, org types.Organization) (bool, error) {
	if !org.HasActiveSubscription() {
		return true, nil
	} else if org.HasActiveSubscriptionWithType(types.SubscriptionTypeTrial) {
		return false, nil
	} else if org.SubscriptionUserAccountQty == nil {
		return true, nil
	} else if vendorCount, err := db.CountUserAccountsByOrgIDAndRole(ctx, org.ID, types.UserRoleVendor); err != nil {
		return true, err
	} else {
		return vendorCount >= *org.SubscriptionUserAccountQty, nil
	}
}

func IsCustomerUserAccountLimitReached(
	ctx context.Context,
	org types.Organization,
	customerOrganization types.CustomerOrganizationWithUserCount,
) (bool, error) {
	if !org.HasActiveSubscription() {
		return true, nil
	} else {
		return !GetUsersPerCustomerOrganizationLimit(org.SubscriptionType).Check(customerOrganization.UserCount), nil
	}
}

func IsCustomerOrganizationLimitReached(ctx context.Context, org types.Organization) (bool, error) {
	if !org.HasActiveSubscription() {
		return true, nil
	} else if org.HasActiveSubscriptionWithType(types.SubscriptionTypeTrial) {
		return false, nil
	} else if org.SubscriptionCustomerOrganizationQty == nil {
		return true, nil
	} else {
		if customerOrgCount, err := db.CountCustomerOrganizationsByOrganizationID(ctx, org.ID); err != nil {
			return true, fmt.Errorf("could not query CustomerOrganization: %w", err)
		} else {
			return customerOrgCount >= *org.SubscriptionCustomerOrganizationQty,
				nil
		}
	}
}

func IsDeploymentTargetLimitReached(
	ctx context.Context,
	org types.Organization,
	customerOrgID *uuid.UUID,
) (bool, error) {
	if !org.HasActiveSubscription() {
		return true, nil
	} else if org.HasActiveSubscriptionWithType(types.SubscriptionTypeTrial) {
		return false, nil
	} else if count, err := db.CountDeploymentTargets(ctx, org.ID, customerOrgID); err != nil {
		return true, fmt.Errorf("could not query DeploymentTarget: %w", err)
	} else {
		return !GetDeploymentTargetsPerCustomerOrganizationLimit(org.SubscriptionType).Check(count), nil
	}
}

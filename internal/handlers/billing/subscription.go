package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/glasskube/distr/internal/api"
	"github.com/glasskube/distr/internal/auth"
	"github.com/glasskube/distr/internal/billing"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/handlerutil"
	"github.com/glasskube/distr/internal/subscription"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func GetSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)
	org := auth.CurrentOrg()

	info, err := buildSubscriptionInfo(ctx, org)
	if err != nil {
		http.Error(w, "failed to build subscription info", http.StatusInternalServerError)
		return
	}

	respondJSON(w, info)
}

func CreateSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)

	var body struct {
		SubscriptionType        types.SubscriptionType `json:"subscriptionType"`
		BillingMode             billing.BillingMode    `json:"billingMode"`
		CustomerOrganizationQty int64                  `json:"subscriptionCustomerOrganizationQuantity"`
		UserAccountQty          int64                  `json:"subscriptionUserAccountQuantity"`
		Currency                string                 `json:"currency"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Debug("bad json payload", zap.Error(err))
		http.Error(w, "bad json payload", http.StatusBadRequest)
		return
	}

	// Default to USD if no currency is provided
	if body.Currency == "" {
		body.Currency = "usd"
	}

	// Get current usage counts
	usage, err := getCurrentUsageCounts(ctx, *auth.CurrentOrgID())
	if err != nil {
		log.Error("failed to get current usage counts", zap.Error(err))
		http.Error(w, "failed to get current usage counts", http.StatusInternalServerError)
		return
	}

	// Validate subscription quantities
	if err := validateSubscriptionQuantities(
		body.SubscriptionType,
		body.CustomerOrganizationQty,
		body.UserAccountQty,
		usage,
	); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	session, err := billing.CreateCheckoutSession(ctx, billing.CheckoutSessionParams{
		OrganizationID:          auth.CurrentOrgID().String(),
		SubscriptionType:        body.SubscriptionType,
		BillingMode:             body.BillingMode,
		CustomerOrganizationQty: body.CustomerOrganizationQty,
		UserAccountQty:          body.UserAccountQty,
		Currency:                body.Currency,
		SuccessURL:              fmt.Sprintf("%v/subscription/callback", handlerutil.GetRequestSchemeAndHost(r)),
	})
	if err != nil {
		log.Error("failed to create checkout session", zap.Error(err))
		http.Error(w, "failed to create checkout session", http.StatusInternalServerError)
		return
	}

	respondJSON(w, api.CheckoutResponse{
		SessionID: session.ID,
		URL:       session.URL,
	})
}

func UpdateSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := internalctx.GetLogger(ctx)
	auth := auth.Authentication.Require(ctx)
	org := auth.CurrentOrg()

	// Check if organization has an active subscription
	if org.SubscriptionType == types.SubscriptionTypeTrial {
		http.Error(w, "cannot update trial subscription, please create a new subscription", http.StatusBadRequest)
		return
	}

	if org.StripeSubscriptionID == nil || *org.StripeSubscriptionID == "" {
		http.Error(w, "no active subscription found", http.StatusBadRequest)
		return
	}

	var body struct {
		CustomerOrganizationQty int64 `json:"subscriptionCustomerOrganizationQuantity"`
		UserAccountQty          int64 `json:"subscriptionUserAccountQuantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Debug("bad json payload", zap.Error(err))
		http.Error(w, "bad json payload", http.StatusBadRequest)
		return
	}

	// Get current usage counts
	usage, err := getCurrentUsageCounts(ctx, org.ID)
	if err != nil {
		log.Error("failed to get current usage counts", zap.Error(err))
		http.Error(w, "failed to get current usage counts", http.StatusInternalServerError)
		return
	}

	// Validate subscription quantities with current subscription type
	if err := validateSubscriptionQuantities(
		org.SubscriptionType,
		body.CustomerOrganizationQty,
		body.UserAccountQty,
		usage,
	); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update the subscription via Stripe
	updatedSub, err := billing.UpdateSubscription(ctx, billing.SubscriptionUpdateParams{
		SubscriptionID:          *org.StripeSubscriptionID,
		CustomerOrganizationQty: body.CustomerOrganizationQty,
		UserAccountQty:          body.UserAccountQty,
		ReturnURL:               fmt.Sprintf("%v/subscription", handlerutil.GetRequestSchemeAndHost(r)),
	})
	if err != nil {
		log.Error("failed to update subscription", zap.Error(err))
		http.Error(w, "failed to update subscription", http.StatusInternalServerError)
		return
	}

	// Extract quantities from the updated subscription
	customerOrgQty, err := billing.GetCustomerOrganizationQty(*updatedSub)
	if err != nil {
		log.Error("failed to get customer organization quantity from updated subscription", zap.Error(err))
		http.Error(w, "failed to get customer organization quantity", http.StatusInternalServerError)
		return
	}

	userAccountQty, err := billing.GetUserAccountQty(*updatedSub)
	if err != nil {
		log.Error("failed to get user account quantity from updated subscription", zap.Error(err))
		http.Error(w, "failed to get user account quantity", http.StatusInternalServerError)
		return
	}

	// Update the organization in the database with new quantities
	org.SubscriptionCustomerOrganizationQty = &customerOrgQty
	org.SubscriptionUserAccountQty = &userAccountQty

	if err := db.UpdateOrganization(ctx, org); err != nil {
		log.Error("failed to update organization", zap.Error(err))
		http.Error(w, "failed to update organization", http.StatusInternalServerError)
		return
	}

	// Reload the organization to get the latest data
	updatedOrg, err := db.GetOrganizationByID(ctx, org.ID)
	if err != nil {
		log.Error("failed to reload organization", zap.Error(err))
		http.Error(w, "failed to reload organization", http.StatusInternalServerError)
		return
	}

	// Build and return the full subscription info
	info, err := buildSubscriptionInfo(ctx, updatedOrg)
	if err != nil {
		log.Error("failed to build subscription info", zap.Error(err))
		http.Error(w, "failed to build subscription info", http.StatusInternalServerError)
		return
	}

	respondJSON(w, info)
}

// validateSubscriptionQuantities validates that the requested quantities meet all requirements
func validateSubscriptionQuantities(
	subscriptionType types.SubscriptionType,
	customerOrgQty int64,
	userAccountQty int64,
	usage *currentUsageCounts,
) error {
	// Validate that requested quantities meet current usage minimums
	if customerOrgQty < int64(usage.customerOrganizationCount) {
		return fmt.Errorf(
			"customer organization quantity (%d) cannot be less than current count (%d)",
			customerOrgQty,
			usage.customerOrganizationCount,
		)
	}

	if userAccountQty < usage.userAccountCount {
		return fmt.Errorf(
			"user account quantity (%d) cannot be less than current count (%d)",
			userAccountQty,
			usage.userAccountCount,
		)
	}

	// Validate that the subscription type limits can accommodate the requested quantities
	customerOrgLimit := subscription.GetCustomersPerOrganizationLimit(subscriptionType)
	if customerOrgLimit != subscription.Unlimited && customerOrgQty > int64(customerOrgLimit) {
		return fmt.Errorf(
			"subscription type %v can have at most %v customer organizations, but %v were requested",
			subscriptionType,
			customerOrgLimit,
			customerOrgQty,
		)
	}

	// Validate that the subscription type can accommodate current max users per customer
	usersPerCustomerLimit := subscription.GetUsersPerCustomerOrganizationLimit(subscriptionType)
	if usersPerCustomerLimit != subscription.Unlimited &&
		usage.maxUsersPerCustomer > 0 &&
		usage.maxUsersPerCustomer > int64(usersPerCustomerLimit) {
		return fmt.Errorf(
			"subscription type %v allows at most %v users per customer organization, "+
				"but you currently have a customer with %v users",
			subscriptionType,
			usersPerCustomerLimit,
			usage.maxUsersPerCustomer,
		)
	}

	// Validate that the subscription type can accommodate current max deployments per customer
	deploymentsPerCustomerLimit := subscription.GetDeploymentTargetsPerCustomerOrganizationLimit(subscriptionType)
	if deploymentsPerCustomerLimit != subscription.Unlimited &&
		usage.maxDeploymentTargetsPerCustomer > 0 &&
		usage.maxDeploymentTargetsPerCustomer > int64(deploymentsPerCustomerLimit) {
		return fmt.Errorf(
			"subscription type %v allows at most %v deployment targets per customer organization, "+
				"but you currently have a customer with %v deployment targets",
			subscriptionType,
			deploymentsPerCustomerLimit,
			usage.maxDeploymentTargetsPerCustomer,
		)
	}

	return nil
}

// buildSubscriptionInfo builds the full subscription info response for an organization
func buildSubscriptionInfo(ctx context.Context, org *types.Organization) (*api.SubscriptionInfo, error) {
	// Get current usage counts
	usage, err := getCurrentUsageCounts(ctx, org.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current usage counts: %w", err)
	}

	// Build limits for all subscription types
	trialLimits := api.SubscriptionLimits{
		MaxCustomerOrganizations: int64(
			subscription.GetCustomersPerOrganizationLimit(types.SubscriptionTypeTrial),
		),
		MaxUsersPerCustomerOrganization: int64(
			subscription.GetUsersPerCustomerOrganizationLimit(types.SubscriptionTypeTrial),
		),
		MaxDeploymentsPerCustomerOrg: int64(
			subscription.GetDeploymentTargetsPerCustomerOrganizationLimit(types.SubscriptionTypeTrial),
		),
	}

	starterLimits := api.SubscriptionLimits{
		MaxCustomerOrganizations: int64(
			subscription.GetCustomersPerOrganizationLimit(types.SubscriptionTypeStarter),
		),
		MaxUsersPerCustomerOrganization: int64(
			subscription.GetUsersPerCustomerOrganizationLimit(types.SubscriptionTypeStarter),
		),
		MaxDeploymentsPerCustomerOrg: int64(
			subscription.GetDeploymentTargetsPerCustomerOrganizationLimit(types.SubscriptionTypeStarter),
		),
	}

	proLimits := api.SubscriptionLimits{
		MaxCustomerOrganizations: int64(
			subscription.GetCustomersPerOrganizationLimit(types.SubscriptionTypePro),
		),
		MaxUsersPerCustomerOrganization: int64(
			subscription.GetUsersPerCustomerOrganizationLimit(types.SubscriptionTypePro),
		),
		MaxDeploymentsPerCustomerOrg: int64(
			subscription.GetDeploymentTargetsPerCustomerOrganizationLimit(types.SubscriptionTypePro),
		),
	}

	enterpriseLimits := api.SubscriptionLimits{
		MaxCustomerOrganizations: int64(
			subscription.GetCustomersPerOrganizationLimit(types.SubscriptionTypeEnterprise),
		),
		MaxUsersPerCustomerOrganization: int64(
			subscription.GetUsersPerCustomerOrganizationLimit(types.SubscriptionTypeEnterprise),
		),
		MaxDeploymentsPerCustomerOrg: int64(
			subscription.GetDeploymentTargetsPerCustomerOrganizationLimit(types.SubscriptionTypeEnterprise),
		),
	}

	info := &api.SubscriptionInfo{
		SubscriptionType:                       org.SubscriptionType,
		SubscriptionEndsAt:                     org.SubscriptionEndsAt.Format("2006-01-02"),
		SubscriptionExternalID:                 org.StripeSubscriptionID,
		SubscriptionCustomerOrganizationQty:    org.SubscriptionCustomerOrganizationQty,
		SubscriptionUserAccountQty:             org.SubscriptionUserAccountQty,
		CurrentUserAccountCount:                usage.userAccountCount,
		CurrentCustomerOrganizationCount:       usage.customerOrganizationCount,
		CurrentMaxUsersPerCustomer:             usage.maxUsersPerCustomer,
		CurrentMaxDeploymentTargetsPerCustomer: usage.maxDeploymentTargetsPerCustomer,
		TrialLimits:                            trialLimits,
		StarterLimits:                          starterLimits,
		ProLimits:                              proLimits,
		EnterpriseLimits:                       enterpriseLimits,
	}

	return info, nil
}

// currentUsageCounts represents the current usage counts for an organization
type currentUsageCounts struct {
	userAccountCount                int64
	customerOrganizationCount       int64
	maxUsersPerCustomer             int64
	maxDeploymentTargetsPerCustomer int64
}

// getCurrentUsageCounts retrieves the current usage counts for the given organization
func getCurrentUsageCounts(ctx context.Context, orgID uuid.UUID) (*currentUsageCounts, error) {
	// Get current user account count
	userAccountCount, err := db.CountUserAccountsByOrgIDAndRole(ctx, orgID, types.UserRoleVendor)
	if err != nil {
		return nil, fmt.Errorf("failed to get user accounts: %w", err)
	}

	// Get current customer organization count
	customerOrgs, err := db.GetCustomerOrganizationsByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer organizations: %w", err)
	}

	// Find the maximum user count and deployment target count across all customer organizations
	var maxUsersPerCustomer int64
	var maxDeploymentTargetsPerCustomer int64
	for _, customerOrg := range customerOrgs {
		if customerOrg.UserCount > maxUsersPerCustomer {
			maxUsersPerCustomer = customerOrg.UserCount
		}
		if customerOrg.DeploymentTargetCount > maxDeploymentTargetsPerCustomer {
			maxDeploymentTargetsPerCustomer = customerOrg.DeploymentTargetCount
		}
	}

	return &currentUsageCounts{
		userAccountCount:                userAccountCount,
		customerOrganizationCount:       int64(len(customerOrgs)),
		maxUsersPerCustomer:             maxUsersPerCustomer,
		maxDeploymentTargetsPerCustomer: maxDeploymentTargetsPerCustomer,
	}, nil
}

// respondJSON is a helper function to send JSON responses
func respondJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

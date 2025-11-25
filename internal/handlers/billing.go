package handlers

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
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/subscription"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func BillingRouter(r chi.Router) {
	r.Use(middleware.RequireOrgAndRole, requireUserRoleVendor)

	r.Get("/subscription", getSubscriptionHandler)
	r.Post("/subscription", postSubscriptionHandler)
}

func getSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)
	org := auth.CurrentOrg()

	// Get current usage counts
	usage, err := getCurrentUsageCounts(ctx, org.ID)
	if err != nil {
		http.Error(w, "failed to get current usage counts", http.StatusInternalServerError)
		return
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

	info := api.SubscriptionInfo{
		SubscriptionType:                       org.SubscriptionType,
		SubscriptionEndsAt:                     org.SubscriptionEndsAt.Format("2006-01-02"),
		SubscriptionExternalID:                 org.SubscriptionExternalID,
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

	RespondJSON(w, info)
}

func postSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
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

	// Validate that requested quantities meet current usage minimums
	if body.CustomerOrganizationQty < int64(usage.customerOrganizationCount) {
		http.Error(
			w,
			fmt.Sprintf(
				"customer organization quantity (%d) cannot be less than current count (%d)",
				body.CustomerOrganizationQty,
				usage.customerOrganizationCount,
			),
			http.StatusBadRequest,
		)
		return
	}

	if body.UserAccountQty < usage.userAccountCount {
		http.Error(
			w,
			fmt.Sprintf(
				"user account quantity (%d) cannot be less than current count (%d)",
				body.UserAccountQty,
				usage.userAccountCount,
			),
			http.StatusBadRequest,
		)
		return
	}

	// Validate that the subscription type limits can accommodate the requested quantities
	customerOrgLimit := subscription.GetCustomersPerOrganizationLimit(body.SubscriptionType)
	if customerOrgLimit != subscription.Unlimited && body.CustomerOrganizationQty > int64(customerOrgLimit) {
		http.Error(
			w,
			fmt.Sprintf(
				"subscription type %v can have at most %v customer organizations, but %v were requested",
				body.SubscriptionType,
				customerOrgLimit,
				body.CustomerOrganizationQty,
			),
			http.StatusBadRequest,
		)
		return
	}

	// Validate that the subscription type can accommodate current max users per customer
	usersPerCustomerLimit := subscription.GetUsersPerCustomerOrganizationLimit(body.SubscriptionType)
	if usersPerCustomerLimit != subscription.Unlimited && usage.maxUsersPerCustomer > 0 && usage.maxUsersPerCustomer > int64(usersPerCustomerLimit) {
		http.Error(
			w,
			fmt.Sprintf(
				"subscription type %v allows at most %v users per customer organization, but you currently have a customer with %v users",
				body.SubscriptionType,
				usersPerCustomerLimit,
				usage.maxUsersPerCustomer,
			),
			http.StatusBadRequest,
		)
		return
	}

	// Validate that the subscription type can accommodate current max deployments per customer
	deploymentsPerCustomerLimit := subscription.GetDeploymentTargetsPerCustomerOrganizationLimit(body.SubscriptionType)
	if deploymentsPerCustomerLimit != subscription.Unlimited && usage.maxDeploymentTargetsPerCustomer > 0 && usage.maxDeploymentTargetsPerCustomer > int64(deploymentsPerCustomerLimit) {
		http.Error(
			w,
			fmt.Sprintf(
				"subscription type %v allows at most %v deployment targets per customer organization, but you currently have a customer with %v deployment targets",
				body.SubscriptionType,
				deploymentsPerCustomerLimit,
				usage.maxDeploymentTargetsPerCustomer,
			),
			http.StatusBadRequest,
		)
		return
	}

	session, err := billing.CreateCheckoutSession(ctx, billing.CheckoutSessionParams{
		OrganizationID:          auth.CurrentOrgID().String(),
		SubscriptionType:        body.SubscriptionType,
		BillingMode:             body.BillingMode,
		CustomerOrganizationQty: body.CustomerOrganizationQty,
		UserAccountQty:          body.UserAccountQty,
		Currency:                body.Currency,
		SuccessURL:              fmt.Sprintf("%v/billing/success", handlerutil.GetRequestSchemeAndHost(r)),
	})
	if err != nil {
		log.Error("failed to create checkout session", zap.Error(err))
		http.Error(w, "failed to create checkout session", http.StatusInternalServerError)
		return
	}

	RespondJSON(w, api.CheckoutResponse{
		SessionID: session.ID,
		URL:       session.URL,
	})
}

// currentUsageCounts represents the current usage counts for an organization
type currentUsageCounts struct {
	userAccountCount                int64
	customerOrganizationCount       int
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
		customerOrganizationCount:       len(customerOrgs),
		maxUsersPerCustomer:             maxUsersPerCustomer,
		maxDeploymentTargetsPerCustomer: maxDeploymentTargetsPerCustomer,
	}, nil
}

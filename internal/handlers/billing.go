package handlers

import (
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

	// Get current user account count
	userAccountCount, err := db.CountUserAccountsByOrgIDAndRole(ctx, org.ID, types.UserRoleVendor)
	if err != nil {
		http.Error(w, "failed to get user accounts", http.StatusInternalServerError)
		return
	}

	// Get current customer organization count
	customerOrgs, err := db.GetCustomerOrganizationsByOrganizationID(ctx, *auth.CurrentOrgID())
	if err != nil {
		http.Error(w, "failed to get customer organizations", http.StatusInternalServerError)
		return
	}

	// Find the maximum user count and deployment target count across all customer organizations
	var maxCurrentUsersPerCustomer int64
	var maxCurrentDeploymentsPerCustomer int64
	for _, customerOrg := range customerOrgs {
		if customerOrg.UserCount > maxCurrentUsersPerCustomer {
			maxCurrentUsersPerCustomer = customerOrg.UserCount
		}
		if customerOrg.DeploymentTargetCount > maxCurrentDeploymentsPerCustomer {
			maxCurrentDeploymentsPerCustomer = customerOrg.DeploymentTargetCount
		}
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
		CurrentUserAccountCount:                userAccountCount,
		CurrentCustomerOrganizationCount:       len(customerOrgs),
		CurrentMaxUsersPerCustomer:             maxCurrentUsersPerCustomer,
		CurrentMaxDeploymentTargetsPerCustomer: maxCurrentDeploymentsPerCustomer,
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

	limit := subscription.GetCustomersPerOrganizationLimit(body.SubscriptionType)
	if limit.IsReached(body.CustomerOrganizationQty) {
		http.Error(
			w,
			fmt.Sprintf("subscription with typ %v can have at most %v customers", body.SubscriptionType, limit),
			http.StatusBadRequest,
		)
		return
	}

	// Default to USD if no currency is provided
	if body.Currency == "" {
		body.Currency = "usd"
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

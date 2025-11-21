package handlers

import (
	"net/http"

	"github.com/glasskube/distr/internal/auth"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
)

type SubscriptionInfo struct {
	SubscriptionType                    types.SubscriptionType `json:"subscriptionType"`
	SubscriptionEndsAt                  string                 `json:"subscriptionEndsAt"`
	SubscriptionExternalID              *string                `json:"subscriptionExternalId"`
	SubscriptionCustomerOrganizationQty *int64                 `json:"subscriptionCustomerOrganizationQuantity"`
	SubscriptionUserAccountQty          *int64                 `json:"subscriptionUserAccountQuantity"`
	CurrentUserAccountCount             int                    `json:"currentUserAccountCount"`
	CurrentCustomerOrganizationCount    int                    `json:"currentCustomerOrganizationCount"`
}

func SubscriptionRouter(r chi.Router) {
	r.Use(middleware.RequireOrgAndRole, requireUserRoleVendor)
	r.Get("/", getSubscriptionHandler)
}

func getSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := auth.Authentication.Require(ctx)
	org := auth.CurrentOrg()

	// Get current user account count
	userAccounts, err := db.GetUserAccountsByOrgID(ctx, *auth.CurrentOrgID())
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

	info := SubscriptionInfo{
		SubscriptionType:                    org.SubscriptionType,
		SubscriptionEndsAt:                  org.SubscriptionEndsAt.Format("2006-01-02T15:04:05Z07:00"),
		SubscriptionExternalID:              org.SubscriptionExternalID,
		SubscriptionCustomerOrganizationQty: org.SubscriptionCustomerOrganizationQty,
		SubscriptionUserAccountQty:          org.SubscriptionUserAccountQty,
		CurrentUserAccountCount:             len(userAccounts),
		CurrentCustomerOrganizationCount:    len(customerOrgs),
	}

	RespondJSON(w, info)
}

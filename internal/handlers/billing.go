package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/glasskube/distr/internal/auth"
	"github.com/glasskube/distr/internal/billing"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/handlerutil"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/glasskube/distr/internal/subscription"
	"github.com/glasskube/distr/internal/types"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func BillingRouter(r chi.Router) {
	r.Use(middleware.RequireOrgAndRole, requireUserRoleVendor)

	r.Post("/checkout", postCheckoutHandler())
}

func postCheckoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"sessionId": session.ID,
			"url":       session.URL,
		}); err != nil {
			log.Error("failed to encode response", zap.Error(err))
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}

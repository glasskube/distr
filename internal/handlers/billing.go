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
	"github.com/glasskube/distr/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
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
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			log.Debug("bad json payload", zap.Error(err))
			http.Error(w, "bad json payload", http.StatusBadRequest)
			return
		}

		if limit := subscription.GetCustomersPerOrganizationLimit(body.SubscriptionType); !limit.Check(body.CustomerOrganizationQty) {
			http.Error(
				w,
				fmt.Sprintf("subscription with typ %v can have at most %v customers", body.SubscriptionType, limit),
				http.StatusBadRequest,
			)
			return
		}

		prices, err := billing.GetStripePrices(ctx, body.SubscriptionType, body.BillingMode)
		if err != nil {
			log.Warn("failed to get stripe prices", zap.Error(err))
			http.Error(w, "failed to get stripe prices", http.StatusInternalServerError)
			return
		}

		params := stripe.CheckoutSessionParams{
			Mode:       util.PtrTo(string(stripe.CheckoutSessionModeSubscription)),
			SuccessURL: util.PtrTo(fmt.Sprintf("%v/billing/success", handlerutil.GetRequestSchemeAndHost(r))),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{Price: &prices.CustomerPriceID, Quantity: util.PtrTo(body.CustomerOrganizationQty)},
				{Price: &prices.UserPriceID, Quantity: util.PtrTo(body.UserAccountQty)},
			},
			SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
				Metadata: map[string]string{
					"organizationId": auth.CurrentOrgID().String(),
				},
			},
		}

		session, err := session.New(&params)
		if err != nil {
			log.Error("failed to create checkout session", zap.Error(err))
			http.Error(w, "failed to create checkout session", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, session.URL, http.StatusFound)
	}
}

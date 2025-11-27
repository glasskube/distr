package billing

import (
	"github.com/glasskube/distr/internal/handlers"
	"github.com/glasskube/distr/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func Router(r chi.Router) {
	r.Use(middleware.RequireOrgAndRole, handlers.RequireUserRoleVendor)

	r.Route("/billing", func(r chi.Router) {
		// Subscription management
		r.Get("/subscription", GetSubscriptionHandler)
		r.Post("/subscription", CreateSubscriptionHandler)
		r.Put("/subscription", UpdateSubscriptionHandler)

		// Billing portal
		r.Post("/billing-portal", CreateBillingPortalSessionHandler)
	})
}

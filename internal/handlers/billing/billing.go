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
		r.Route("/subscription", func(r chi.Router) {
			r.Get("/", GetSubscriptionHandler)
			r.Post("/", CreateSubscriptionHandler)
			r.Put("/", UpdateSubscriptionHandler)
		})
		// Billing portal
		r.Post("/billing-portal", CreateBillingPortalSessionHandler)
	})
}

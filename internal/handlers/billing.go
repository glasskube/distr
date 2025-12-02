package handlers

import (
	"github.com/glasskube/distr/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func BillingRouter(r chi.Router) {
	r.Use(middleware.RequireOrgAndRole, RequireUserRoleVendor)
	r.Route("/subscription", func(r chi.Router) {
		r.Get("/", GetSubscriptionHandler)
		r.Post("/", CreateSubscriptionHandler)
		r.Put("/", UpdateSubscriptionHandler)
	})
	r.Post("/portal", CreateBillingPortalSessionHandler)
}

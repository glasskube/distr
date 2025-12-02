package handlers

import (
	"github.com/glasskube/distr/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func BillingRouter(r chi.Router) {
	r.Use(middleware.RequireOrgAndRole, middleware.RequireVendor)
	r.Route("/subscription", func(r chi.Router) {
		r.Get("/", GetSubscriptionHandler)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAdmin)
			r.Post("/", CreateSubscriptionHandler)
			r.Put("/", UpdateSubscriptionHandler)
		})
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAdmin)
		r.Post("/portal", CreateBillingPortalSessionHandler)
	})
}

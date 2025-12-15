package handlers

import (
	"github.com/glasskube/distr/internal/middleware"
	"github.com/oaswrap/spec/adapters/chiopenapi"
)

func BillingRouter(r chiopenapi.Router) {
	r.Use(middleware.RequireOrgAndRole, middleware.RequireVendor)
	r.Route("/subscription", func(r chiopenapi.Router) {
		r.Get("/", GetSubscriptionHandler)
		r.Group(func(r chiopenapi.Router) {
			r.Use(middleware.RequireAdmin)
			r.Post("/", CreateSubscriptionHandler)
			r.Put("/", UpdateSubscriptionHandler)
		})
	})
	r.Group(func(r chiopenapi.Router) {
		r.Use(middleware.RequireAdmin)
		r.Post("/portal", CreateBillingPortalSessionHandler)
	})
}

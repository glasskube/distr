package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func WebhookRouter(r chi.Router) {
	const MaxBodyBytes = int64(65536)

	r.With(middleware.RequestSize(MaxBodyBytes)).
		Post("/stripe", stripeWebhookHandler())
}

package handlers

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/oaswrap/spec/adapters/chiopenapi"
)

func WebhookRouter(r chiopenapi.Router) {
	const MaxBodyBytes = int64(65536)

	r.With(middleware.RequestSize(MaxBodyBytes)).
		Post("/stripe", stripeWebhookHandler())
}

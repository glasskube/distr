package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/glasskube/distr/internal/billing"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/env"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/webhook"
	"go.uber.org/zap"
)

func stripeWebhookHandler() http.HandlerFunc {
	endpointSecret := env.StripeWebhookSecret()

	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		log := internalctx.GetLogger(ctx)

		if endpointSecret == nil {
			log.Warn("stripe endpoint secret not set")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		payload, err := io.ReadAll(req.Body)
		if err != nil {
			log.Warn("error reading request body", zap.Error(err))
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"), *endpointSecret)
		if err != nil {
			log.Warn("webhook signature verification failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log = log.With(zap.String("stripeEventId", event.ID), zap.String("stripeEventType", string(event.Type)))
		ctx = internalctx.WithLogger(ctx, log)

		switch event.Type {
		case stripe.EventTypeCustomerSubscriptionCreated:
			var subscription stripe.Subscription
			err := json.Unmarshal(event.Data.Raw, &subscription)
			if err != nil {
				log.Info("Error parsing webhook JSON", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			log.Info("stripe customer subscription created")

			if err := handleStripeSubscription(ctx, subscription); err != nil {
				log.Error("Error handling stripe subscription", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		case stripe.EventTypeCustomerSubscriptionUpdated:
			var subscription stripe.Subscription
			err := json.Unmarshal(event.Data.Raw, &subscription)
			if err != nil {
				log.Info("Error parsing webhook JSON", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			log.Info("stripe customer subscription updated")

			if err := handleStripeSubscription(ctx, subscription); err != nil {
				log.Error("Error handling stripe subscription", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		case stripe.EventTypeCustomerSubscriptionDeleted:
			var subscription stripe.Subscription
			err := json.Unmarshal(event.Data.Raw, &subscription)
			if err != nil {
				log.Info("Error parsing webhook JSON", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			log.Info("stripe customer subscription deleted", zap.Any("subscription", subscription))

			if err := handleStripeSubscription(ctx, subscription); err != nil {
				log.Error("Error handling stripe subscription", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		default:
			log.Info("unhandled stripe event")
		}

		w.WriteHeader(http.StatusOK)
	}
}

func handleStripeSubscription(ctx context.Context, subscription stripe.Subscription) error {
	log := internalctx.GetLogger(ctx)

	orgId, err := uuid.Parse(subscription.Metadata["organizationId"])
	if err != nil {
		log.Warn("subscription event with missing or invalid organizationId", zap.Error(err))
	}

	org, err := db.GetOrganizationByID(ctx, orgId)
	if err != nil {
		return err
	}

	org.StripeSubscriptionId = &subscription.ID
	org.StripeCustomerId = &subscription.Customer.ID

	if subscription.Status == stripe.SubscriptionStatusCanceled {
		org.SubscriptionEndsAt = time.Now()
	} else if currentPeriodEnd, err := billing.GetCurrentPeriodEnd(subscription); err != nil {
		return err
	} else {
		org.SubscriptionEndsAt = *currentPeriodEnd
	}

	if subscriptionType, err := billing.GetSubscriptionType(subscription); err != nil {
		return err
	} else {
		org.SubscriptionType = *subscriptionType
	}

	if qty, err := billing.GetCustomerOrganizationQty(subscription); err != nil {
		return err
	} else {
		org.SubscriptionCustomerOrganizationQty = &qty
	}

	if qty, err := billing.GetUserAccountQty(subscription); err != nil {
		return err
	} else {
		org.SubscriptionUserAccountQty = &qty
	}

	log.Info("updated organization subscription",
		zap.Stringer("organizationId", org.ID),
		zap.String("subscriptionType", string(org.SubscriptionType)),
		zap.Time("subscriptionEndsAt", org.SubscriptionEndsAt),
		zap.Int64p("userAccountQty", org.SubscriptionUserAccountQty),
		zap.Int64p("customerOrganizationQty", org.SubscriptionCustomerOrganizationQty))

	if err := db.UpdateOrganization(ctx, org); err != nil {
		return err
	}

	return nil
}

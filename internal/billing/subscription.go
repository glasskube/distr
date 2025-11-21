package billing

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
)

func GetSubscriptionType(subscription stripe.Subscription) (*types.SubscriptionType, error) {
	var result *types.SubscriptionType

	for _, item := range subscription.Items.Data {
		if item.Price != nil {
			if slices.Contains(ProPriceKeys, item.Price.LookupKey) {
				if result == nil {
					result = util.PtrTo(types.SubscriptionTypePro)
				} else if *result != types.SubscriptionTypePro {
					return nil, fmt.Errorf("multiple subscription types found")
				}
			} else if slices.Contains(StarterPriceKeys, item.Price.LookupKey) {
				if result == nil {
					result = util.PtrTo(types.SubscriptionTypeStarter)
				} else if *result != types.SubscriptionTypeStarter {
					return nil, fmt.Errorf("multiple subscription types found")
				}
			}
		}
	}

	if result == nil {
		return nil, fmt.Errorf("no subscription type found")
	}

	return result, nil
}

func GetCurrentPeriodEnd(subscription stripe.Subscription) (*time.Time, error) {
	var result *time.Time
	for _, item := range subscription.Items.Data {
		if t := time.Unix(item.CurrentPeriodEnd, 0); !t.IsZero() {
			if result == nil {
				result = &t
			} else if !t.Equal(*result) {
				return nil, fmt.Errorf("multiple current period ends found")
			}
		}
	}
	if result == nil {
		return nil, fmt.Errorf("no current period end found")
	}
	return result, nil
}

func GetUserAccountQty(subscription stripe.Subscription) (int64, error) {
	for _, item := range subscription.Items.Data {
		if item.Price != nil && slices.Contains(UserPriceKeys, item.Price.LookupKey) {
			return item.Quantity, nil
		}
	}
	return 0, fmt.Errorf("no unit price for UserAccount found")
}

func GetCustomerOrganizationQty(subscription stripe.Subscription) (int64, error) {
	for _, item := range subscription.Items.Data {
		if item.Price != nil && slices.Contains(CustomerPriceKeys, item.Price.LookupKey) {
			return item.Quantity, nil
		}
	}
	return 0, fmt.Errorf("no unit price for CustomerOrganization found")
}

type CheckoutSessionParams struct {
	OrganizationID          string
	SubscriptionType        types.SubscriptionType
	BillingMode             BillingMode
	CustomerOrganizationQty int64
	UserAccountQty          int64
	Currency                string
	SuccessURL              string
}

func CreateCheckoutSession(ctx context.Context, params CheckoutSessionParams) (*stripe.CheckoutSession, error) {
	prices, err := GetStripePrices(ctx, params.SubscriptionType, params.BillingMode)
	if err != nil {
		return nil, fmt.Errorf("failed to get stripe prices: %w", err)
	}

	sessionParams := &stripe.CheckoutSessionParams{
		Params:     stripe.Params{Context: ctx},
		Mode:       util.PtrTo(string(stripe.CheckoutSessionModeSubscription)),
		Currency:   util.PtrTo(params.Currency),
		SuccessURL: util.PtrTo(params.SuccessURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{Price: &prices.CustomerPriceID, Quantity: util.PtrTo(params.CustomerOrganizationQty)},
			{Price: &prices.UserPriceID, Quantity: util.PtrTo(params.UserAccountQty)},
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"organizationId": params.OrganizationID,
			},
		},
	}

	return session.New(sessionParams)
}

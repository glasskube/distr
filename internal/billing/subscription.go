package billing

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
	"github.com/stripe/stripe-go/v84"
	checkoutsession "github.com/stripe/stripe-go/v84/checkout/session"
	"github.com/stripe/stripe-go/v84/subscription"
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

func GetSubscriptionPeriode(subscription stripe.Subscription) (types.SubscriptionPeriode, error) {

	for _, item := range subscription.Items.Data {
		if item.Price != nil {
			lookupKey := item.Price.LookupKey
			if slices.Contains(MonthlyPriceKeys, lookupKey) {
				return types.SubscriptionPeriodeMonthly, nil
			}
			if slices.Contains(YearlyPriceKeys, lookupKey) {
				return types.SubscriptionPeriodeYearly, nil
			}
		}
	}
	return types.SubscriptionPeriodeMonthly, fmt.Errorf("no subscription periode found in subscription prices")
}

type CheckoutSessionParams struct {
	OrganizationID          string
	SubscriptionType        types.SubscriptionType
	SubscriptionPeriode     types.SubscriptionPeriode
	CustomerOrganizationQty int64
	UserAccountQty          int64
	Currency                string
	SuccessURL              string
}

func CreateCheckoutSession(ctx context.Context, params CheckoutSessionParams) (*stripe.CheckoutSession, error) {
	prices, err := GetStripePrices(ctx, params.SubscriptionType, params.SubscriptionPeriode)
	if err != nil {
		return nil, fmt.Errorf("failed to get stripe prices: %w", err)
	}

	sessionParams := &stripe.CheckoutSessionParams{
		Params:     stripe.Params{Context: ctx},
		Mode:       util.PtrTo(string(stripe.CheckoutSessionModeSubscription)),
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

	return checkoutsession.New(sessionParams)
}

type SubscriptionUpdateParams struct {
	SubscriptionID          string
	CustomerOrganizationQty int64
	UserAccountQty          int64
	ReturnURL               string
}

func UpdateSubscription(ctx context.Context, params SubscriptionUpdateParams) (*stripe.Subscription, error) {
	// Get the existing subscription to find the price IDs
	sub, err := subscription.Get(params.SubscriptionID, &stripe.SubscriptionParams{
		Params: stripe.Params{Context: ctx},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	// Find the existing subscription items with their price IDs
	var customerPriceID, userPriceID string
	var customerItemID, userItemID string

	for _, item := range sub.Items.Data {
		if item.Price != nil && item.Price.LookupKey != "" {
			if slices.Contains(CustomerPriceKeys, item.Price.LookupKey) {
				customerPriceID = item.Price.ID
				customerItemID = item.ID
			} else if slices.Contains(UserPriceKeys, item.Price.LookupKey) {
				userPriceID = item.Price.ID
				userItemID = item.ID
			}
		}
	}

	if customerPriceID == "" || userPriceID == "" {
		return nil, fmt.Errorf("could not find price IDs in subscription")
	}

	// Update the subscription with new quantities
	// Stripe will automatically prorate the charges
	updateParams := &stripe.SubscriptionParams{
		Params: stripe.Params{Context: ctx},
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:       util.PtrTo(customerItemID),
				Price:    util.PtrTo(customerPriceID),
				Quantity: util.PtrTo(params.CustomerOrganizationQty),
			},
			{
				ID:       util.PtrTo(userItemID),
				Price:    util.PtrTo(userPriceID),
				Quantity: util.PtrTo(params.UserAccountQty),
			},
		},
	}

	updatedSub, err := subscription.Update(params.SubscriptionID, updateParams)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	return updatedSub, nil
}

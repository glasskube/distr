package billing

import (
	"context"
	"fmt"

	"github.com/glasskube/distr/internal/types"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/price"
)

type BillingMode string

const (
	BillingModeMonthly BillingMode = "monthly"
	BillingModeYearly  BillingMode = "yearly"
)

const (
	PriceKeyStarterCustomerMonthly = "distr_starter_customer_monthly"
	PriceKeyStarterCustomerYearly  = "distr_starter_customer_yearly"
	PriceKeyStarterUserMonthly     = "distr_starter_user_monthly"
	PriceKeyStarterUserYearly      = "distr_starter_user_yearly"
	PriceKeyProCustomerMonthly     = "distr_pro_customer_monthly"
	PriceKeyProCustomerYearly      = "distr_pro_customer_yearly"
	PriceKeyProUserMonthly         = "distr_pro_user_monthly"
	PriceKeyProUserYearly          = "distr_pro_user_yearly"
)

var (
	CustomerPriceKeys = []string{
		PriceKeyStarterCustomerMonthly,
		PriceKeyStarterCustomerYearly,
		PriceKeyProCustomerMonthly,
		PriceKeyProCustomerYearly,
	}
	UserPriceKeys = []string{
		PriceKeyStarterUserMonthly,
		PriceKeyStarterUserYearly,
		PriceKeyProUserMonthly,
		PriceKeyProUserYearly,
	}
	StarterPriceKeys = []string{
		PriceKeyStarterCustomerMonthly,
		PriceKeyStarterCustomerYearly,
		PriceKeyStarterUserMonthly,
		PriceKeyStarterUserYearly,
	}
	ProPriceKeys = []string{
		PriceKeyProCustomerMonthly,
		PriceKeyProCustomerYearly,
		PriceKeyProUserMonthly,
		PriceKeyProUserYearly,
	}
)

type PriceIDs struct {
	CustomerPriceID string
	UserPriceID     string
}

func GetStripePrices(
	ctx context.Context,
	subscriptionType types.SubscriptionType,
	mode BillingMode,
) (*PriceIDs, error) {
	var customerPriceLookupKey string
	var userPriceLookupKey string

	switch subscriptionType {
	case types.SubscriptionTypeStarter:
		switch mode {
		case BillingModeMonthly:
			customerPriceLookupKey = PriceKeyStarterCustomerMonthly
			userPriceLookupKey = PriceKeyStarterUserMonthly
		case BillingModeYearly:
			customerPriceLookupKey = PriceKeyStarterCustomerYearly
			userPriceLookupKey = PriceKeyStarterUserYearly
		default:
			return nil, fmt.Errorf("invalid billing mode: %v", mode)
		}
	case types.SubscriptionTypePro:
		switch mode {
		case BillingModeMonthly:
			customerPriceLookupKey = PriceKeyProCustomerMonthly
			userPriceLookupKey = PriceKeyProUserMonthly
		case BillingModeYearly:
			customerPriceLookupKey = PriceKeyProCustomerYearly
			userPriceLookupKey = PriceKeyProUserYearly
		default:
			return nil, fmt.Errorf("invalid billing mode: %v", mode)
		}
	default:
		return nil, fmt.Errorf("invalid subscription type: %v", subscriptionType)
	}

	listPriceResult := price.List(&stripe.PriceListParams{
		ListParams: stripe.ListParams{Context: ctx},
		LookupKeys: stripe.StringSlice([]string{customerPriceLookupKey, userPriceLookupKey}),
	})

	var result PriceIDs
	for listPriceResult.Next() {
		price := listPriceResult.Price()
		switch price.LookupKey {
		case customerPriceLookupKey:
			result.CustomerPriceID = price.ID
		case userPriceLookupKey:
			result.UserPriceID = price.ID
		}
	}

	if err := listPriceResult.Err(); err != nil {
		return nil, err
	}

	if result.CustomerPriceID == "" || result.UserPriceID == "" {
		return nil, fmt.Errorf("failed to find prices")
	}

	return &result, nil
}

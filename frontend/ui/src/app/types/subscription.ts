export type SubscriptionType = 'starter' | 'pro' | 'enterprise' | 'trial';

export interface SubscriptionInfo {
  subscriptionType: SubscriptionType;
  subscriptionEndsAt: string;
  subscriptionExternalId?: string;
  subscriptionCustomerOrganizationQuantity?: number;
  subscriptionUserAccountQuantity?: number;
  currentUserAccountCount: number;
  currentCustomerOrganizationCount: number;
}

export interface CheckoutRequest {
  subscriptionType: SubscriptionType;
  billingMode: 'monthly' | 'yearly';
  subscriptionUserAccountQuantity: number;
  subscriptionCustomerOrganizationQuantity: number;
  currency: string;
}

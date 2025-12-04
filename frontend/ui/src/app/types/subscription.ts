export type SubscriptionType = 'starter' | 'pro' | 'enterprise' | 'trial';

export interface SubscriptionLimits {
  maxCustomerOrganizations: number;
  maxUsersPerCustomerOrganization: number;
  maxDeploymentsPerCustomerOrganization: number;
}

export interface SubscriptionInfo {
  subscriptionType: SubscriptionType;
  subscriptionEndsAt: string;
  subscriptionExternalId?: string;
  subscriptionSubscriptionPeriode?: 'monthly' | 'yearly';
  subscriptionCustomerOrganizationQuantity?: number;
  subscriptionUserAccountQuantity?: number;
  currentUserAccountCount: number;
  currentCustomerOrganizationCount: number;
  currentMaxUsersPerCustomer: number;
  currentMaxDeploymentTargetsPerCustomer: number;
  trialLimits: SubscriptionLimits;
  starterLimits: SubscriptionLimits;
  proLimits: SubscriptionLimits;
  enterpriseLimits: SubscriptionLimits;
}

export interface CheckoutRequest {
  subscriptionType: SubscriptionType;
  subscriptionPeriode: 'monthly' | 'yearly';
  subscriptionUserAccountQuantity: number;
  subscriptionCustomerOrganizationQuantity: number;
}

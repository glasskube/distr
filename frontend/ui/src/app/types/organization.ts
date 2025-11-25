import {BaseModel, Named, UserRole} from '@glasskube/distr-sdk';

export type Feature = 'licensing';

export type SubscriptionType = 'trial' | 'starter' | 'pro' | 'enterprise';

export interface Organization extends BaseModel, Named {
  slug?: string;
  features: Feature[];
  appDomain?: string;
  registryDomain?: string;
  emailFromAddress?: string;
  subscriptionType?: SubscriptionType;
  subscriptionEndsAt?: string;
}

export interface OrganizationWithUserRole extends Organization {
  userRole: UserRole;
  joinedOrgAt: string;
}

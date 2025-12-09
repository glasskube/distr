import {BaseModel, Named, UserRole} from '@glasskube/distr-sdk';
import {SubscriptionType} from './subscription';

export type Feature = 'licensing';

export interface SubscriptionLimits {
  maxCustomerOrganizations: number;
  maxUsersPerCustomerOrganization: number;
  maxDeploymentsPerCustomerOrganization: number;
}

export interface Organization extends BaseModel, Named {
  slug?: string;
  features: Feature[];
  appDomain?: string;
  registryDomain?: string;
  emailFromAddress?: string;
  subscriptionType: SubscriptionType;
  subscriptionLimits: SubscriptionLimits;
  subscriptionEndsAt?: string;
  subscriptionCustomerOrganizationQuantity: number;
  subscriptionUserAccountQuantity: number;
}

export interface OrganizationWithUserRole extends Organization {
  userRole: UserRole;
  customerOrganizationId?: string;
  customerOrganizationName?: string;
  joinedOrgAt: string;
}

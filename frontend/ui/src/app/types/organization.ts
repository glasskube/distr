import {BaseModel, Named, UserRole} from '@glasskube/distr-sdk';

export type Feature = 'licensing';

export interface Organization extends BaseModel, Named {
  slug?: string;
  features: Feature[];
  appDomain?: string;
  registryDomain?: string;
  emailFromAddress?: string;
}

export interface OrganizationWithUserRole extends Organization {
  userRole: UserRole;
}

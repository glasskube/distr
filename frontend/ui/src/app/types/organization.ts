import {BaseModel, Named} from '@glasskube/distr-sdk';

export type Feature = 'licensing';

export interface Organization extends BaseModel, Named {
  slug?: string;
  features: Feature[];
  appDomain?: string;
  registryDomain?: string;
  emailFromAddress?: string;
}

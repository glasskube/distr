import {BaseModel, Named} from '@glasskube/distr-sdk';

export type Feature = 'licensing' | 'registry';

export interface Organization extends BaseModel, Named {
  slug?: string;
  features: Feature[];
}

import {BaseModel, Named} from '@glasskube/distr-sdk';

export type Feature = 'licensing';

export interface Organization extends BaseModel, Named {
  features: Feature[];
}

import {BaseModel, Named} from '@glasskube/distr-sdk';

export interface Organization extends BaseModel, Named {
  licensingEnabled: boolean;
}

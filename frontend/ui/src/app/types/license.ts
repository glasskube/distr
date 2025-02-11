import {ApplicationVersion, BaseModel} from '@glasskube/distr-sdk';

export interface License extends BaseModel {
  expiresAt?: string;
  name: string;
  versions: ApplicationVersion[];
}

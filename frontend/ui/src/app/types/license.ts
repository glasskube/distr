import {ApplicationVersion, BaseModel} from '@glasskube/distr-sdk';

export interface License extends BaseModel {
  expiresAt?: string;
  name: string;
  applicationId: string;
  versions: ApplicationVersion[];
}

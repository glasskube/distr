import {ApplicationVersion, BaseModel, Named} from '@glasskube/distr-sdk';

export interface ApplicationLicense extends BaseModel, Named {
  expiresAt?: Date;
  applicationId?: string;
  versions?: ApplicationVersion[];
  ownerUserAccountId?: string;

  registryUrl?: string;
  registryUsername?: string;
  registryPassword?: string;
}

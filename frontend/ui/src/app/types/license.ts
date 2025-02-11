import {ApplicationVersion, BaseModel, Named} from '@glasskube/distr-sdk';

export interface License extends BaseModel, Named {
  applicationId?: string;
  versions?: ApplicationVersion[];
  ownerUserAccountId?: string;

  registryUrl?: string;
  registryUsername?: string;
  registryPassword?: string;
}

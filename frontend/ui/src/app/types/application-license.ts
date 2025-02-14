import {Application, ApplicationVersion, BaseModel, Named, UserAccount} from '@glasskube/distr-sdk';

export interface ApplicationLicense extends BaseModel, Named {
  expiresAt?: Date;
  applicationId?: string;
  application?: Application;
  versions?: ApplicationVersion[];
  ownerUserAccountId?: string;
  owner?: UserAccount;

  registryUrl?: string;
  registryUsername?: string;
  registryPassword?: string;
}

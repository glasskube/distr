import {Application, ApplicationVersion, BaseModel, Named} from '@glasskube/distr-sdk';
import {CustomerOrganization} from '../../../../../sdk/js/dist';

export interface ApplicationLicense extends BaseModel, Named {
  expiresAt?: Date;
  applicationId?: string;
  application?: Application;
  versions?: ApplicationVersion[];
  customerOrganizationId?: string;
  customerOrganization?: CustomerOrganization;

  registryUrl?: string;
  registryUsername?: string;
  registryPassword?: string;
}

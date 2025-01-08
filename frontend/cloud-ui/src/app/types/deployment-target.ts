import {BaseModel, Named} from './base';
import {DeploymentType} from './deployment';
import {Geolocation} from './geolocation';
import {UserAccountWithRole} from './user-account';

export interface DeploymentTarget extends BaseModel, Named {
  name: string;
  type: DeploymentType;
  namespace?: string;
  geolocation?: Geolocation;
  createdBy?: UserAccountWithRole;

  currentStatus?: DeploymentTargetStatus;
}

export interface DeploymentTargetStatus extends BaseModel {
  message: string;
}

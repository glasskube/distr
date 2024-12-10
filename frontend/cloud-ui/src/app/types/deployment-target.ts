import {BaseModel, Named} from './base';
import {Geolocation} from './geolocation';
import {UserAccount} from './user-account';

export interface DeploymentTarget extends BaseModel, Named {
  name: string;
  type: string;
  geolocation?: Geolocation;
  createdBy?: UserAccount;

  currentStatus?: DeploymentTargetStatus;
}

export interface DeploymentTargetStatus extends BaseModel {
  message: string;
}

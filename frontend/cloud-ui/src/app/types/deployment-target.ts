import {BaseModel, Named} from './base';
import {Geolocation} from './geolocation';
import {DeploymentWithData} from './deployment';
import {Observable} from 'rxjs';

export interface DeploymentTarget extends BaseModel, Named {
  name: string;
  type: string;
  geolocation?: Geolocation;
  latestDeployment?: Observable<DeploymentWithData>;
  currentStatus?: DeploymentTargetStatus;
}

export interface DeploymentTargetStatus extends BaseModel {
  message: string;
}

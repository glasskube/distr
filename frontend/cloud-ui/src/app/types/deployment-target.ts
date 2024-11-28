import {BaseModel, Named} from './base';
import {Geolocation} from './geolocation';
import {Deployment, DeploymentWithData} from './deployment';
import {Observable} from 'rxjs';

export interface DeploymentTarget extends BaseModel, Named {
  name: string;
  type: string;
  geolocation?: Geolocation;
  latestDeployment?: Observable<DeploymentWithData>;
}

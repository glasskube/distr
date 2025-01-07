import {DeploymentTarget} from '../types/deployment-target';
import {Observable} from 'rxjs';
import {DeploymentWithData} from '../types/deployment';

export interface DeploymentTargetViewModel extends DeploymentTarget {
  id: string;
  latestDeployment?: Observable<DeploymentWithData>;
}

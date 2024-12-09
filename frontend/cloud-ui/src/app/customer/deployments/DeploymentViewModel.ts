import {Observable} from 'rxjs';
import {DeploymentWithData} from '../../types/deployment';
import {DeploymentTarget} from '../../types/deployment-target';

export interface DeploymentViewModel extends DeploymentTarget {
  latestDeployment?: Observable<DeploymentWithData>;
}

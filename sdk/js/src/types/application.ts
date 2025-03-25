import {BaseModel, Named} from './base';
import {DeploymentType, HelmChartType, DockerType} from './deployment';

export interface Application extends BaseModel, Named {
  type: DeploymentType;
  versions?: ApplicationVersion[];
}

export interface ApplicationVersion {
  id?: string;
  name: string;
  createdAt?: string;
  archivedAt?: string;
  applicationId?: string;
  chartType?: HelmChartType;
  dockerType?: DockerType;
  chartName?: string;
  chartUrl?: string;
  chartVersion?: string;
}

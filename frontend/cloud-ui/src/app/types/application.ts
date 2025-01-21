import {BaseModel, Named} from './base';
import {DeploymentType, HelmChartType} from './deployment';

export interface Application extends BaseModel, Named {
  type: DeploymentType;
  versions?: ApplicationVersion[];
}

export interface ApplicationVersion {
  id?: string;
  name?: string;
  createdAt?: string;
  applicationId?: string;
  chartType?: HelmChartType;
  chartName?: string;
  chartUrl?: string;
  chartVersion?: string;
}

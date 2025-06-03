import {BaseModel, Named} from './base';
import {DeploymentType, HelmChartType, DockerType} from './deployment';

export interface Application extends BaseModel, Named {
  type: DeploymentType;
  imageUrl?: string;
  versions?: ApplicationVersion[];
}

export interface ApplicationVersion {
  id?: string;
  name: string;
  createdAt?: string;
  archivedAt?: string;
  applicationId?: string;
  chartType?: HelmChartType;
  chartName?: string;
  chartUrl?: string;
  chartVersion?: string;
}

export interface PatchApplicationRequest {
  name?: string;
  versions?: {id: string; archivedAt?: string}[];
}

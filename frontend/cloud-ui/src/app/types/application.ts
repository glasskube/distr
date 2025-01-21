import {BaseModel, Named} from './base';
import {DeploymentType, HelmChartType} from './deployment';
import {ApplicationVersion} from '@glasskube/cloud-sdk';

export interface Application extends BaseModel, Named {
  type: DeploymentType;
  versions?: ApplicationVersion[];
}

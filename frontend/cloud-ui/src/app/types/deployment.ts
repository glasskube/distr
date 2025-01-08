import {BaseModel} from './base';

export interface Deployment extends BaseModel {
  applicationVersionId: string;
  deploymentTargetId: string;
  releaseName?: string;
  valuesYaml?: string;
  note?: string;
}

export interface DeploymentWithData extends Deployment {
  applicationId: string;
  applicationName: string;
  applicationVersionName: string;
}

export type DeploymentType = 'docker' | 'kubernetes';

export type HelmChartType = 'repository' | 'oci';

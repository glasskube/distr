import {BaseModel} from './base';

export interface Deployment extends BaseModel {
  deploymentTargetId: string;
  releaseName?: string;
  dockerType?: DockerType;
  logsEnabled: boolean;
}

export interface DeploymentRequest {
  deploymentTargetId: string;
  applicationVersionId: string;
  deploymentId?: string;
  applicationLicenseId?: string;
  releaseName?: string;
  dockerType?: DockerType;
  valuesYaml?: string;
  envFileData?: string;
  logsEnabled?: boolean;
  forceRestart?: boolean;
  ignoreRevisionSkew?: boolean;
}

export interface PatchDeploymentRequest {
  logsEnabled?: boolean;
}

export interface DeploymentWithLatestRevision extends Deployment {
  applicationId: string;
  applicationName: string;
  applicationVersionId: string;
  applicationVersionName: string;
  applicationLink: string;
  applicationLicenseId?: string;
  valuesYaml?: string;
  envFileData?: string;
  deploymentRevisionId?: string;
  deploymentRevisionCreatedAt?: string;
  latestStatus?: DeploymentRevisionStatus;
}

export interface DeploymentRevisionStatus extends BaseModel {
  type: DeploymentStatusType;
  message: string;
}

export type DeploymentType = 'docker' | 'kubernetes';

export type HelmChartType = 'repository' | 'oci';

export type DockerType = 'compose' | 'swarm';

export type DeploymentStatusType = 'ok' | 'progressing' | 'error';

export type DeploymentTargetScope = 'cluster' | 'namespace';

import {BaseModel, Named} from './base';
import {Geolocation} from './geolocation';

export interface Deployment extends BaseModel {
  applicationVersionId: string;
  deploymentTargetId: string;
  note?: string;
}

export interface DeploymentWithData extends Deployment {
  applicationId: string;
  applicationName: string;
  applicationVersionName: string;
}

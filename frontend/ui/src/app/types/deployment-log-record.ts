import {BaseModel} from '@glasskube/distr-sdk';

export interface DeploymentLogRecord extends BaseModel {
  deploymentId: string;
  deploymentRevisionId: string;
  resource: string;
  timestamp: string;
  severity: string;
  body: string;
}

import {AgentVersion} from './agent-version';
import {BaseModel, Named} from './base';
import {DeploymentTargetScope, DeploymentType, DeploymentWithLatestRevision} from './deployment';
import {Geolocation} from './geolocation';
import {UserAccountWithRole} from './user-account';

export interface DeploymentTarget extends BaseModel, Named {
  name: string;
  type: DeploymentType;
  namespace?: string;
  scope?: DeploymentTargetScope;
  geolocation?: Geolocation;
  createdBy?: UserAccountWithRole;
  currentStatus?: DeploymentTargetStatus;
  deployment?: DeploymentWithLatestRevision;
  agentVersion?: AgentVersion;
  reportedAgentVersionId?: string;
}

export interface DeploymentTargetStatus extends BaseModel {
  message: string;
}

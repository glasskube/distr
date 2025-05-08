import {AgentVersion} from './agent-version';
import {BaseModel, Named} from './base';
import {DeploymentTargetScope, DeploymentType, DeploymentWithLatestRevision} from './deployment';
import {Geolocation} from './geolocation';
import {UserAccountWithRole} from './user-account';

export interface DeploymentTargetBase extends BaseModel, Named {
  name: string;
  type: DeploymentType;
  namespace?: string;
  scope?: DeploymentTargetScope;
  geolocation?: Geolocation;
  createdBy?: UserAccountWithRole;
  currentStatus?: DeploymentTargetStatus;
  agentVersion?: AgentVersion;
  reportedAgentVersionId?: string;
}

export interface DeploymentTarget extends DeploymentTargetBase {
  /**
   * @deprecated This property will be removed in v2. Please consider using `deployments` instead.
   */
  deployment?: DeploymentWithLatestRevision;
  deployments: DeploymentWithLatestRevision[];
}

export interface DeploymentTargetStatus extends BaseModel {
  message: string;
}

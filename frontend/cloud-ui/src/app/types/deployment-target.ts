import {AgentVersion} from './agent-version';
import {BaseModel, Named} from './base';
import {DeploymentType, DeploymentWithData} from './deployment';
import {Geolocation} from './geolocation';
import {UserAccountWithRole} from './user-account';

export interface DeploymentTarget extends BaseModel, Named {
  name: string;
  type: DeploymentType;
  namespace?: string;
  geolocation?: Geolocation;
  createdBy?: UserAccountWithRole;
  currentStatus?: DeploymentTargetStatus;
  latestDeployment?: DeploymentWithData;
  agentVersion?: AgentVersion;
  reportedAgentVersionId?: string;
}

export interface DeploymentTargetStatus extends BaseModel {
  message: string;
}

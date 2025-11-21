import {AgentVersion} from './agent-version';
import {BaseModel, Named} from './base';
import {CustomerOrganization} from './customer-organization';
import {DeploymentTargetScope, DeploymentType, DeploymentWithLatestRevision} from './deployment';
import {UserAccountWithRole} from './user-account';

export interface DeploymentTarget extends BaseModel, Named {
  name: string;
  type: DeploymentType;
  namespace?: string;
  scope?: DeploymentTargetScope;
  createdBy?: UserAccountWithRole;
  customerOrganization?: CustomerOrganization;
  currentStatus?: DeploymentTargetStatus;
  /**
   * @deprecated This property will be removed in v2. Please consider using `deployments` instead.
   */
  deployment?: DeploymentWithLatestRevision;
  deployments: DeploymentWithLatestRevision[];
  agentVersion?: AgentVersion;
  reportedAgentVersionId?: string;
  metricsEnabled: boolean;
}

export interface DeploymentTargetStatus extends BaseModel {
  message: string;
}

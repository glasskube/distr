import {Client, ClientConfig} from './client';
import {
  Application,
  ApplicationVersion,
  Deployment,
  DeploymentTarget,
  DeploymentTargetAccessResponse,
  DeploymentType,
  DeploymentWithLatestRevision,
} from '../types';

export type LatestVersionStrategy = 'semver' | 'chronological';

export type CreateDeploymentResult = {
  deploymentTarget: DeploymentTarget;
  access: DeploymentTargetAccessResponse;
};

export type IsOutdatedResult = {
  deploymentTarget: DeploymentTarget;
  application: Application;
  newerVersions: ApplicationVersion[];
  outdated: boolean;
};

export class CloudService {
  private readonly client: Client;
  constructor(
    clientConfig: ClientConfig,
    private readonly latestVersionStrategy: LatestVersionStrategy = 'chronological'
  ) {
    this.client = new Client(clientConfig);
  }

  public async createDockerApplicationVersion(
    applicationId: string,
    name: string,
    composeFile: string
  ): Promise<ApplicationVersion> {
    return this.client.createApplicationVersion(applicationId, {name}, {composeFile});
  }

  public async createKubernetesApplicationVersion(
    applicationId: string,
    name: string,
    baseValuesFile?: string,
    templateFile?: string
  ): Promise<ApplicationVersion> {
    return this.client.createApplicationVersion(applicationId, {name}, {baseValuesFile, templateFile});
  }

  public async createDeployment(
    target: {deploymentName: string; type: DeploymentType; namespace?: string},
    application: {id: string; versionId?: string}
  ): Promise<CreateDeploymentResult> {
    // TODO support kubernetes

    const deploymentTarget = await this.client.createDeploymentTarget({
      name: target.deploymentName,
      type: target.type,
      namespace: target.namespace,
    });
    let versionId = application.versionId;
    if (!application.versionId) {
      // TODO properly implement latest version strategies
      const app = await this.client.getApplication(application.id);
      versionId = app.versions?.[0].id;
    }
    await this.client.createOrUpdateDeployment({
      deploymentTargetId: deploymentTarget.id!,
      applicationVersionId: versionId!,
    });
    const access = await this.client.createAccessForDeploymentTarget(deploymentTarget.id!);
    return {
      deploymentTarget: await this.client.getDeploymentTarget(deploymentTarget.id!),
      access,
    };
  }

  public async updateDeployment(deploymentTargetId: string, applicationVersionId?: string): Promise<any> {
    // TODO support kubernetes

    const existing = await this.client.getDeploymentTarget(deploymentTargetId);
    if (!existing.deployment) {
      throw new Error('cannot update deployment, because nothing deployed yet');
    }
    let versionId = applicationVersionId;
    if (!versionId) {
      const app = await this.client.getApplication(existing.deployment.applicationId!);
      // TODO properly implement latest version strategies
      versionId = app.versions?.[0].id;
    }
    return this.client.createOrUpdateDeployment({
      deploymentTargetId,
      deploymentId: existing.deployment.id,
      applicationVersionId: versionId!,
    });
  }

  public async isOutdated(deploymentTargetId: string): Promise<IsOutdatedResult> {
    const existing = await this.client.getDeploymentTarget(deploymentTargetId);
    if (!existing.deployment) {
      throw new Error('nothing deployed yet');
    }
    const app = await this.client.getApplication(existing.deployment.applicationId!);
    return {
      deploymentTarget: existing,
      application: app,
      newerVersions: [], // TODO
      outdated: app.versions?.[0].id !== existing.deployment.applicationVersionId, // TODO properly implement latest version strategies
    };
  }
}

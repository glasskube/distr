import {Client, ClientConfig} from './client';
import {
  Application,
  ApplicationVersion,
  DeploymentTarget,
  DeploymentTargetAccessResponse,
  DeploymentTargetScope,
  DeploymentType,
  HelmChartType,
} from '../types';
import semver from 'semver/preload';

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
    private readonly latestVersionStrategy: LatestVersionStrategy = 'semver'
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
    versionName: string,
    data: {
      chartName: string;
      chartVersion: string;
      chartType: HelmChartType;
      chartUrl: string;
      baseValuesFile?: string;
      templateFile?: string;
    }
  ): Promise<ApplicationVersion> {
    return this.client.createApplicationVersion(
      applicationId,
      {
        name: versionName,
        chartName: data.chartName,
        chartVersion: data.chartVersion,
        chartType: data.chartType,
        chartUrl: data.chartUrl,
      },
      {
        baseValuesFile: data.baseValuesFile,
        templateFile: data.templateFile,
      }
    );
  }

  public async createDeployment(
    target: {deploymentName: string; type: DeploymentType; namespace?: string; scope?: DeploymentTargetScope},
    application: {id: string; versionId?: string}
  ): Promise<CreateDeploymentResult> {
    // TODO support kubernetes

    const deploymentTarget = await this.client.createDeploymentTarget({
      name: target.deploymentName,
      type: target.type,
      namespace: target.namespace,
      scope: target.scope,
    });
    let versionId = application.versionId;
    if (!versionId) {
      const latest = await this.getLatestVersion(application.id);
      if (!latest) {
        throw new Error('no versions available');
      }
      versionId = latest.id!;
    }
    await this.client.createOrUpdateDeployment({
      deploymentTargetId: deploymentTarget.id!,
      applicationVersionId: versionId,
    });
    return {
      deploymentTarget: await this.client.getDeploymentTarget(deploymentTarget.id!),
      access: await this.client.createAccessForDeploymentTarget(deploymentTarget.id!),
    };
  }

  public async updateDeployment(deploymentTargetId: string, applicationVersionId?: string): Promise<any> {
    // TODO support kubernetes

    const existing = await this.client.getDeploymentTarget(deploymentTargetId);
    if (!existing.deployment) {
      // TODO or should we create a new one here? (deploying to an existing target isn't possible yet)
      throw new Error('cannot update deployment, because nothing deployed yet');
    }
    let versionId = applicationVersionId;
    if (!versionId) {
      const res = await this.isOutdated(existing.id!);
      if (res.outdated && res.newerVersions.length > 0) {
        versionId = res.newerVersions[res.newerVersions.length - 1].id;
      } else {
        throw new Error('cannot update deployment, there seems to be no newer version');
      }
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
    const {app, newerVersions} = await this.getNewerVersions(
      existing.deployment.applicationId!,
      existing.deployment.applicationVersionId!
    );
    return {
      deploymentTarget: existing,
      application: app,
      newerVersions: newerVersions,
      outdated: newerVersions.length > 0,
    };
  }

  private async getLatestVersion(appId: string): Promise<ApplicationVersion | undefined> {
    const {newerVersions} = await this.getNewerVersions(appId);
    return newerVersions.length > 0 ? newerVersions[newerVersions.length - 1] : undefined;
  }

  private async getNewerVersions(
    appId: string,
    currentVersionId?: string
  ): Promise<{app: Application; newerVersions: ApplicationVersion[]}> {
    const app = await this.client.getApplication(appId);
    const currentVersion = (app.versions || []).find((it) => it.id === currentVersionId);
    if (!currentVersion && currentVersionId) {
      throw new Error('given version ID does not exist in this application');
    }
    const newerVersions = (app.versions || [])
      .filter((it) => {
        if (!currentVersion) {
          return true;
        }
        // surely there are fancier ways to deal with strategies but that's it for now
        switch (this.latestVersionStrategy) {
          case 'semver':
            return semver.gt(it.name!, currentVersion.name!, {loose: true});
          case 'chronological':
            return it.createdAt! > currentVersion.createdAt!; // TODO proper date handling maybe
        }
      })
      .sort((a, b) => {
        switch (this.latestVersionStrategy) {
          case 'semver':
            return semver.compare(a.name!, b.name!, {loose: true});
          case 'chronological':
            return a.createdAt?.localeCompare(b.createdAt!) ?? 0; // TODO proper date handling maybe
        }
      });
    return {app, newerVersions};
  }
}

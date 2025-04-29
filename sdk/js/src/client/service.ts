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
import {ConditionalPartial, defaultClientConfig} from './config';

/**
 * The strategy for determining the latest version of an application.
 * * 'semver' uses semantic versioning to determine the latest version.
 * * 'chronological' uses the creation date of the versions to determine the latest version.
 */
export type LatestVersionStrategy = 'semver' | 'chronological';

export type CreateDeploymentParams = {
  target: {
    name: string;
    type: DeploymentType;
    kubernetes?: {
      namespace: string;
      scope: DeploymentTargetScope;
    };
  };
  application: {
    id?: string;
    versionId?: string;
  };
  kubernetesDeployment?: {
    releaseName: string;
    valuesYaml?: string;
  };
};

export type CreateDeploymentResult = {
  deploymentTarget: DeploymentTarget;
  access: DeploymentTargetAccessResponse;
};

export type UpdateDeploymentParams = {
  deploymentTargetId: string;
  applicationVersionId?: string;
  kubernetesDeployment?: {
    valuesYaml?: string;
  };
};

export type IsOutdatedResult = {
  deploymentTarget: DeploymentTarget;
  application: Application;
  newerVersions: ApplicationVersion[];
  outdated: boolean;
};

/**
 * The DistrService provides a higher-level API for the Distr API. It allows to create and update deployments, check
 * if a deployment is outdated, and get the latest version of an application according to a specified strategy.
 * Under the hood it uses the low-level {@link Client}.
 */
export class DistrService {
  private readonly client: Client;

  /**
   * Creates a new DistrService instance. A client config containing an API key must be provided, optionally the API
   * base URL can be set. Optionally, a strategy for determining the latest version of an application can be specified –
   * the default is semantic versioning.
   * @param config ClientConfig containing at least an API key and optionally an API base URL
   * @param latestVersionStrategy Strategy for determining the latest version of an application (default: 'semver')
   */
  constructor(
    config: ConditionalPartial<ClientConfig, keyof typeof defaultClientConfig>,
    private readonly latestVersionStrategy: LatestVersionStrategy = 'semver'
  ) {
    this.client = new Client(config);
  }

  /**
   * Creates a new application version for the given docker application using a Docker Compose file and an
   * optional template file.
   * @param applicationId
   * @param name Name of the new version
   * @param data
   */
  public async createDockerApplicationVersion(
    applicationId: string,
    name: string,
    data: {
      composeFile: string;
      templateFile?: string;
    }
  ): Promise<ApplicationVersion> {
    return this.client.createApplicationVersion(
      applicationId,
      {name},
      {
        composeFile: data.composeFile,
        templateFile: data.templateFile,
      }
    );
  }

  /**
   * Creates a new application version for the given Kubernetes application using a Helm chart.
   * @param applicationId
   * @param versionName
   * @param data
   */
  public async createKubernetesApplicationVersion(
    applicationId: string,
    versionName: string,
    data: {
      chartName?: string;
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

  /**
   * Creates a new deployment target and deploys the given application version to it.
   * * If deployment type is 'kubernetes', the namespace and scope must be provided.
   * * If deployment type is 'kubernetes', the helm release name and values YAML can be provided.
   * * If no application version ID is given, the latest version of the application will be deployed.
   * @param params
   */
  public async createDeployment(params: CreateDeploymentParams): Promise<CreateDeploymentResult> {
    const {target, application, kubernetesDeployment} = params;

    let versionId = application.versionId;
    if (!versionId) {
      if (!application.id) {
        throw new Error('application ID or version ID must be provided');
      }
      const latest = await this.getLatestVersion(application.id);
      if (!latest) {
        throw new Error('no version available for this application');
      }
      versionId = latest.id!;
    }

    const deploymentTarget = await this.client.createDeploymentTarget({
      name: target.name,
      type: target.type,
      namespace: target.kubernetes?.namespace,
      scope: target.kubernetes?.scope,
      deployments: [],
    });
    await this.client.createOrUpdateDeployment({
      deploymentTargetId: deploymentTarget.id!,
      applicationVersionId: versionId,
      releaseName: kubernetesDeployment?.releaseName,
      valuesYaml: kubernetesDeployment?.valuesYaml ? btoa(kubernetesDeployment?.valuesYaml) : undefined,
    });
    return {
      deploymentTarget: await this.client.getDeploymentTarget(deploymentTarget.id!),
      access: await this.client.createAccessForDeploymentTarget(deploymentTarget.id!),
    };
  }

  /**
   * Updates the deployment of an existing deployment target. If no application version ID is given, the latest version
   * of the already deployed application will be deployed.
   * @param params
   */
  public async updateDeployment(params: UpdateDeploymentParams): Promise<void> {
    const {deploymentTargetId, applicationVersionId, kubernetesDeployment} = params;

    const existing = await this.client.getDeploymentTarget(deploymentTargetId);
    if (!existing.deployment && !applicationVersionId) {
      throw new Error('cannot update deployment, because nothing deployed yet');
    }
    let versionId = applicationVersionId;
    if (!versionId) {
      const res = await this.isOutdated(existing.id!);
      if (res.outdated && res.newerVersions.length > 0) {
        versionId = res.newerVersions[res.newerVersions.length - 1].id!;
      } else if (existing.deployment) {
        // version stays the same, other params might have changed
        versionId = existing.deployment.applicationVersionId;
      } else {
        throw new Error('cannot update deployment, because nothing deployed yet');
      }
    }
    await this.client.createOrUpdateDeployment({
      deploymentTargetId,
      deploymentId: existing.deployment?.id,
      applicationVersionId: versionId,
      valuesYaml: kubernetesDeployment?.valuesYaml ? btoa(kubernetesDeployment?.valuesYaml) : undefined,
    });
  }

  /**
   * Checks if the given deployment target is outdated, i.e. if there is a newer version of the application available.
   * The result additionally contains versions that are newer than the currently deployed one, ordered ascending.
   * @param deploymentTargetId
   */
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

  /**
   * Returns the latest version of the given application according to the specified strategy.
   * @param appId
   */
  public async getLatestVersion(appId: string): Promise<ApplicationVersion | undefined> {
    const {newerVersions} = await this.getNewerVersions(appId);
    return newerVersions.length > 0 ? newerVersions[newerVersions.length - 1] : undefined;
  }

  /**
   * Returns the application and all versions that are newer than the given version ID. If no version ID is given,
   * all versions are considered. The versions are ordered ascending according to the given strategy.
   * @param appId
   * @param currentVersionId
   */
  public async getNewerVersions(
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

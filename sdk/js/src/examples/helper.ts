import {Client, ClientConfig} from '../client';
import {DeploymentType} from '../types';

export const clientConfig = {
  apiBase: 'http://localhost:8080/api/v1',
  apiKey: 'gkc-973b942c4c6cc579e0b46390afb400c4',
};

export async function getSomeDockerAppId(clientConfig: ClientConfig): Promise<string> {
  return await firstAppOfType(clientConfig, 'docker');
}

export async function getSomeKubernetesAppId(clientConfig: ClientConfig): Promise<string> {
  return await firstAppOfType(clientConfig, 'kubernetes');
}

export async function getSomeDockerDeploymentTargetId(clientConfig: ClientConfig): Promise<string> {
  return await firstDeploymentTargetOfType(clientConfig, 'docker');
}

export async function getSomeKubernetesDeploymentTargetId(clientConfig: ClientConfig): Promise<string> {
  return await firstDeploymentTargetOfType(clientConfig, 'kubernetes');
}

async function firstAppOfType(clientConfig: ClientConfig, type: DeploymentType): Promise<string> {
  const lowLevelClient = new Client(clientConfig);
  const apps = await lowLevelClient.getApplications();
  const firstApp = apps.find((a) => a.type === type);
  return firstApp?.id ?? '';
}

async function firstDeploymentTargetOfType(clientConfig: ClientConfig, type: DeploymentType): Promise<string> {
  const lowLevelClient = new Client(clientConfig);
  const dts = await lowLevelClient.getDeploymentTargets();
  const firstDeploymentTarget = dts.find((a) => a.type === type);
  return firstDeploymentTarget?.id ?? '';
}

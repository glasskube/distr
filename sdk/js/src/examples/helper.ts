import {Client, ClientConfig} from '../client';

export async function getSomeDockerAppId(clientConfig: ClientConfig): Promise<string> {
  const lowLevelClient = new Client(clientConfig);
  const apps = await lowLevelClient.getApplications();
  const firstDockerApp = apps.find((a) => a.type === 'docker');
  return firstDockerApp?.id ?? '';
}

export async function getSomeKubernetesAppId(clientConfig: ClientConfig): Promise<string> {
  const lowLevelClient = new Client(clientConfig);
  const apps = await lowLevelClient.getApplications();
  const firstKubernetesApp = apps.find((a) => a.type === 'kubernetes');
  return firstKubernetesApp?.id ?? '';
}

export async function getSomeDeploymentTargetId(): Promise<string> {
  const lowLevelClient = new Client(clientConfig); // just needed to provide some testing ids
  const dts = await lowLevelClient.getDeploymentTargets();
  const firstDocker = dts.find((a) => a.type === 'docker');
  return firstDocker?.id ?? '';
}

// this would be injected via ENV
export const clientConfig = {
  apiBase: 'http://localhost:8080/api/v1',
  apiKey:
    'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAdGVzdC5jb20yIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImV4cCI6MTczNzY1MTAyMSwiaWF0IjoxNzM3NTY0NjIxLCJuYW1lIjoiIiwibmJmIjoxNzM3NTY0NjIxLCJvcmciOiJiZjdmNGFkZi00YjVjLTQ0YjQtOTJiYS05ZmNhNzVkN2E4NTQiLCJyb2xlIjoidmVuZG9yIiwic3ViIjoiZDJjZDYyNmUtMTA5OC00YWExLTk2NmMtYWJiMjkwZTJjZDRjIn0.dw2NcP_o6u7fdthDMfyepb-iAa5Ad-8crQeaXdgV7L4',
};

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

// this would be injected via ENV
export const clientConfig = {
  apiBase: 'http://localhost:8080/api/v1',
  apiKey:
    'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImhhaGFAaGFoYS5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZXhwIjoxNzM3NjI2MDMyLCJpYXQiOjE3Mzc1Mzk2MzIsIm5hbWUiOiIiLCJuYmYiOjE3Mzc1Mzk2MzIsIm9yZyI6IjkxYmZmMDcxLTRjZjMtNGQ2Ny1hMDMyLWU3YTkzZDRjNGYzMSIsInJvbGUiOiJ2ZW5kb3IiLCJzdWIiOiI1NjI4NTJmZi0xNWFiLTQwMjctOTFjNi1kYTczMmMyNjA2ZGEifQ.ZBDa8UlmsRGkrbjaF7DlYi352pom9ramYWdDrETulr0',
};

import {CloudService} from '../client/service';
import {Client} from '../client';

// this would be injected via ENV
const clientConfig = {
  apiBase: 'http://localhost:8080/api/v1',
  apiKey:
    'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImhhaGFAaGFoYS5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZXhwIjoxNzM3NjI2MDMyLCJpYXQiOjE3Mzc1Mzk2MzIsIm5hbWUiOiIiLCJuYmYiOjE3Mzc1Mzk2MzIsIm9yZyI6IjkxYmZmMDcxLTRjZjMtNGQ2Ny1hMDMyLWU3YTkzZDRjNGYzMSIsInJvbGUiOiJ2ZW5kb3IiLCJzdWIiOiI1NjI4NTJmZi0xNWFiLTQwMjctOTFjNi1kYTczMmMyNjA2ZGEifQ.ZBDa8UlmsRGkrbjaF7DlYi352pom9ramYWdDrETulr0',
};

const gc = new CloudService(clientConfig, 'semver');

try {
  const deploymentTargetId = 'c555ea88-409a-4ecd-9912-76ff031d39cc'; // await getSomeDeploymentTargetId(); // this would be replaced by something injected via ENV
  const outdatedRes = await gc.isOutdated(deploymentTargetId);
  console.log(outdatedRes);
} catch (e) {
  console.error(e);
}

async function getSomeDeploymentTargetId(): Promise<string> {
  const lowLevelClient = new Client(clientConfig); // just needed to provide some testing ids
  const dts = await lowLevelClient.getDeploymentTargets();
  const firstDocker = dts.find((a) => a.type === 'docker');
  return firstDocker?.id ?? '';
}

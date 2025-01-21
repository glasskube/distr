import {CloudService} from '../client/service';
import {Client} from '../client';

// TODO this would be injected via ENV
const clientConfig = {
  apiBase: 'http://localhost:8080/api/v1',
  apiKey:
    'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImhhaGFAaGFoYS5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZXhwIjoxNzM3NTMwMzcwLCJpYXQiOjE3Mzc0NDM5NzAsIm5hbWUiOiIiLCJuYmYiOjE3Mzc0NDM5NzAsIm9yZyI6IjkxYmZmMDcxLTRjZjMtNGQ2Ny1hMDMyLWU3YTkzZDRjNGYzMSIsInJvbGUiOiJ2ZW5kb3IiLCJzdWIiOiI1NjI4NTJmZi0xNWFiLTQwMjctOTFjNi1kYTczMmMyNjA2ZGEifQ.KDAQDCUrpUeFI9gkDwcZr5_vP9dPoh-adlv25JK-je8',
};

let appId = ''; // TODO this would be injected via ENV

const lowLevelClient = new Client(clientConfig); // just needed to provide some testing ids
const gc = new CloudService(clientConfig); // the high level one

try {
  const apps = await lowLevelClient.getApplications();
  const firstDockerApp = apps.find((a) => a.type === 'docker');
  appId = firstDockerApp?.id!;
  const newDockerVersion = await gc.createDockerApplicationVersion(appId, 'v1.0.0', 'hello: world');
  console.log(`* created new version ${newDockerVersion.name} (id: ${newDockerVersion.id}) for docker app ${appId}`);
} catch (e) {
  console.error(e);
}

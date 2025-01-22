import {CloudService} from '../client/service';
import {clientConfig, getSomeDockerAppId, getSomeKubernetesAppId} from './helper';

const gc = new CloudService(clientConfig);

// replace with your application ID and your compose file
const appId = await getSomeDockerAppId(clientConfig);
// TODO
const newDockerVersion = await gc.createDockerApplicationVersion(appId, 'v7.4.2', `
services:
  web:
    build: .
    ports:
      - "8000:5000"
  redis:
    image: "redis:7.4.2-alpine"
`);
console.log(`* created new version ${newDockerVersion.name} (id: ${newDockerVersion.id}) for docker app ${appId}`);

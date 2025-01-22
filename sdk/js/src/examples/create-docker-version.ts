import {CloudService} from '../client/service';
import {clientConfig, getSomeDockerAppId} from './helper';

const gc = new CloudService(clientConfig);
const appId = await getSomeDockerAppId(clientConfig);
const newDockerVersion = await gc.createDockerApplicationVersion(
  appId,
  'v7.4.2',
  `
services:
  web:
    build: .
    ports:
      - "8000:5000"
  redis:
    image: "redis:7.4.2-alpine"
`
);
console.log(`* created new version ${newDockerVersion.name} (id: ${newDockerVersion.id}) for docker app ${appId}`);

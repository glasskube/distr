import {DistrService} from '../client/service';
import {clientConfig} from './config';

const gc = new DistrService(clientConfig);
const appId = '<docker-application-id>';
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

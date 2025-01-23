import {CloudService} from '../client/service';
import {clientConfig, getSomeDockerAppId} from './helper';

const gc = new CloudService(clientConfig);

const appId = await getSomeDockerAppId(clientConfig);
const result = await gc.createDeployment({
  target: {
    name: 'test-docker-deployment',
    type: 'docker',
  },
  application: {
    id: appId,
  },
});
console.log(result);

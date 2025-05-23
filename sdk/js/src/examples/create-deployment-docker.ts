import {DistrService} from '../client/service';
import {clientConfig} from './config';

const gc = new DistrService(clientConfig);

const appId = '<docker-application-id>';
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

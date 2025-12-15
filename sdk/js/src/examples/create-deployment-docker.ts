import {DistrService} from '../client';
import {clientConfig} from './config';

const distr = new DistrService(clientConfig);

const appId = '<docker-application-id>';
const result = await distr.createDeployment({
  target: {
    name: 'test-docker-deployment',
    type: 'docker',
  },
  application: {
    id: appId,
  },
});
console.log(result);

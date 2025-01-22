import {CloudService} from '../client/service';
import {clientConfig, getSomeDockerAppId} from './helper';

const gc = new CloudService(clientConfig);

const appId = await getSomeDockerAppId(clientConfig);
const result = await gc.createDeployment({deploymentName: 'test-docker-deployment', type: 'docker'}, {id: appId});
console.log(result);

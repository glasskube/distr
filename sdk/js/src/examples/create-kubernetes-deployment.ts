import {CloudService} from '../client/service';
import {clientConfig, getSomeKubernetesAppId} from './helper';

const gc = new CloudService(clientConfig);

const appId = await getSomeKubernetesAppId(clientConfig);
const result = await gc.createDeployment(
  {
    deploymentName: 'test-kubernetes-deployment',
    type: 'kubernetes',
    namespace: 'my-namespace',
    scope: 'namespace',
  },
  {id: appId}
);
console.log(result);

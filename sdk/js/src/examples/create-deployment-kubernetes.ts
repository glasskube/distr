import {CloudService} from '../client/service';
import {clientConfig, getSomeKubernetesAppId} from './helper';

const gc = new CloudService(clientConfig);

const appId = await getSomeKubernetesAppId(clientConfig);
const result = await gc.createDeployment({
  target: {
    name: 'test-kubernetes-deployment',
    type: 'kubernetes',
    kubernetes: {
      namespace: 'my-namespace',
      scope: 'namespace',
    },
  },
  application: {
    id: appId,
  },
  kubernetesDeployment: {
    releaseName: 'my-release',
    valuesYaml: 'my-values: true',
  },
});
console.log(result);

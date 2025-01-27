import {DistrService} from '../client/service';
import {clientConfig} from './config';

const gc = new DistrService(clientConfig);

const appId = '<kubernetes-application-id>';
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

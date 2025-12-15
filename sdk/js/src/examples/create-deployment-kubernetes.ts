import {DistrService} from '../client';
import {clientConfig} from './config';

const distr = new DistrService(clientConfig);

const appId = '<kubernetes-application-id>';
await distr.createDeployment({
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

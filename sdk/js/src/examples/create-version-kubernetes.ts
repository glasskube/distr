import {DistrService} from '../client';
import {clientConfig} from './config';

const distr = new DistrService(clientConfig);
const kubernetesAppId = '<kubernetes-application-id>';
const newKubernetesVersion = await distr.createKubernetesApplicationVersion(kubernetesAppId, 'v1.0.1', {
  chartName: 'my-chart',
  chartVersion: '1.0.1',
  chartType: 'repository',
  chartUrl: 'https://example.com/my-chart-1.0.1.tgz',
  baseValuesFile: 'base: values',
  templateFile: 'template: true',
});
console.log(
  `* created new version ${newKubernetesVersion.name} (id: ${newKubernetesVersion.id}) for kubernetes app ${kubernetesAppId}`
);

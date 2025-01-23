import {CloudService} from '../client/service';
import {clientConfig, getSomeKubernetesAppId} from './helper';

const gc = new CloudService(clientConfig);
const kubernetesAppId = await getSomeKubernetesAppId(clientConfig);
const newKubernetesVersion = await gc.createKubernetesApplicationVersion(kubernetesAppId, 'v1.0.1', {
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

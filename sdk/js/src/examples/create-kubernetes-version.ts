import {CloudService} from '../client/service';
import {clientConfig, getSomeDockerAppId, getSomeKubernetesAppId} from './helper';

const gc = new CloudService(clientConfig); // client config should be injected via ENV

// replace with your application ID and your chart
const kubernetesAppId = await getSomeKubernetesAppId(clientConfig); // this would be replaced by something injected via ENV
const newKubernetesVersion = await gc.createKubernetesApplicationVersion(
  kubernetesAppId,
  'v1.0.0',
  'base: values',
  'template: true'
);
console.log(
  `* created new version ${newKubernetesVersion.name} (id: ${newKubernetesVersion.id}) for kubernetes app ${kubernetesAppId}`
);

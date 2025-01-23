import {CloudService} from '../client/service';
import {clientConfig, getSomeKubernetesDeploymentTargetId} from './helper';

const gc = new CloudService(clientConfig);

const deploymentTargetId = '755e9329-cc32-407a-aa59-d220202cf899'; // await getSomeKubernetesDeploymentTargetId(clientConfig);
await gc.updateDeployment({deploymentTargetId, kubernetesDeployment: {valuesYaml: 'new: values'}}); // update to latest version (according to the given strategy) of application that is already deployed

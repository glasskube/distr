import {CloudService} from '../client/service';
import {clientConfig, getSomeDockerDeploymentTargetId} from './helper';

const gc = new CloudService(clientConfig);

const deploymentTargetId = await getSomeDockerDeploymentTargetId(clientConfig);
await gc.updateDeployment({deploymentTargetId}); // update to latest version (according to the given strategy) of application that is already deployed

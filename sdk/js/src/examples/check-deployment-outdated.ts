import {CloudService} from '../client/service';
import {clientConfig, getSomeDockerDeploymentTargetId} from './helper';

const gc = new CloudService(clientConfig);

const deploymentTargetId = '79138254-31c1-4e09-b511-4862d257c80d'; // await getSomeDeploymentTargetId();
const outdatedRes = await gc.isOutdated(deploymentTargetId);
console.log(outdatedRes);

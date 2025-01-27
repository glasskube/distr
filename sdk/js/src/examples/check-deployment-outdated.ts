import {CloudService} from '../client/service';
import {clientConfig} from './config';

const gc = new CloudService(clientConfig);

const deploymentTargetId = '<your-deployment-target-id>';
const outdatedRes = await gc.isOutdated(deploymentTargetId);
console.log(outdatedRes);

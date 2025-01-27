import {DistrService} from '../client/service';
import {clientConfig} from './config';

const gc = new DistrService(clientConfig);

const deploymentTargetId = '<your-deployment-target-id>';
const outdatedRes = await gc.isOutdated(deploymentTargetId);
console.log(outdatedRes);

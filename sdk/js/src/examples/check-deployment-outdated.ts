import {DistrService} from '../client';
import {clientConfig} from './config';

const distr = new DistrService(clientConfig);

const deploymentTargetId = '<your-deployment-target-id>';
const outdatedRes = await distr.isOutdated(deploymentTargetId);
console.log(outdatedRes);

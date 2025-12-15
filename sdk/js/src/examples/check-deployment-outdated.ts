import {DistrService} from '../client';
import {clientConfig} from './config';

const distr = new DistrService(clientConfig);

const deploymentTargetId = '<your-deployment-target-id>';
await distr.isOutdated(deploymentTargetId);

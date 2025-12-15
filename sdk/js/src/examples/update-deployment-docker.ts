import {DistrService} from '../client';
import {clientConfig} from './config';

const distr = new DistrService(clientConfig);

const deploymentTargetId = '<docker-deployment-target-id>';
const applicationId = '<docker-application-id>';
const applicationVersionId = '<docker-application-version-id>';
await distr.updateDeployment({deploymentTargetId, applicationId, applicationVersionId}); // update to latest version (according to the given strategy) of application that is already deployed

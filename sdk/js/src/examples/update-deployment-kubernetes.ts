import {DistrService} from '../client';
import {clientConfig} from './config';

const distr = new DistrService(clientConfig);

const deploymentTargetId = '<kubernetes-deployment-target-id>';
const applicationId = '<kubernetes-application-id>';
const applicationVersionId = '<kubernetes-application-version-id>';
await distr.updateDeployment({
  deploymentTargetId,
  applicationId,
  applicationVersionId,
  kubernetesDeployment: {valuesYaml: 'new: values'},
}); // update to latest version (according to the given strategy) of application that is already deployed

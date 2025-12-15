import {DistrService} from '../client';
import {clientConfig} from './config';

const distr = new DistrService(clientConfig);

const applicationId = '<application-id>';
const applicationVersionId = '<application-version-id>';

await distr.updateAllDeployments(applicationId, applicationVersionId);

import {CloudService} from '../client/service';
import {clientConfig} from './helper';

const gc = new CloudService(clientConfig);

const deploymentTargetId = '79138254-31c1-4e09-b511-4862d257c80d'; // await getSomeDeploymentTargetId();
await gc.updateDeployment(deploymentTargetId); // update to latest version (according to the given strategy) of application that is already deployed

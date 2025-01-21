import {Client} from '../index';

const client = new Client({
  apiBase: 'http://localhost:8080/api/v1',
  apiKey:
    'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImhhaGFAaGFoYS5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZXhwIjoxNzM3NTMwMzcwLCJpYXQiOjE3Mzc0NDM5NzAsIm5hbWUiOiIiLCJuYmYiOjE3Mzc0NDM5NzAsIm9yZyI6IjkxYmZmMDcxLTRjZjMtNGQ2Ny1hMDMyLWU3YTkzZDRjNGYzMSIsInJvbGUiOiJ2ZW5kb3IiLCJzdWIiOiI1NjI4NTJmZi0xNWFiLTQwMjctOTFjNi1kYTczMmMyNjA2ZGEifQ.KDAQDCUrpUeFI9gkDwcZr5_vP9dPoh-adlv25JK-je8',
});

try {
  let newDockerApp = await client.createApplication({
    type: 'docker',
    name: 'My New Docker App via SDK',
  });
  log(newDockerApp, 'create docker application');

  newDockerApp.name = 'My Updated Docker App';
  newDockerApp = await client.updateApplication(newDockerApp);
  log(newDockerApp, 'update docker application');

  const newDockerVersion = await client.createApplicationVersion(
    newDockerApp,
    {
      name: 'v1',
    },
    {composeFile: 'hello: world'}
  );
  log(newDockerVersion, 'create docker application version');

  let newKubernetesApp = await client.createApplication({
    type: 'kubernetes',
    name: 'My New Kubernetes App via SDK',
  });
  log(newKubernetesApp, 'create kubernetes application');

  const newKubernetesVersion = await client.createApplicationVersion(
    newKubernetesApp,
    {
      name: 'v1',
      chartName: 'my-chart',
      chartVersion: '1.0.0',
      chartType: 'repository',
      chartUrl: 'https://my.chart.repo',
    },
    {templateFile: 'hello', baseValuesFile: 'base: values'}
  );
  log(newKubernetesVersion, 'create kubernetes application version');

  const applications = await client.getApplications();
  log(applications, 'get applications');
  for (let a of applications) {
    const app = await client.getApplication(a.id!);
    log(app, 'get application by id');
  }

  const newDockerDeploymentTarget = await client.createDeploymentTarget({
    name: 'My New Docker Deployment Target via SDK',
    type: 'docker',
  });
  log(newDockerDeploymentTarget, 'create docker deployment target');

  const newKubernetesDeploymentTarget = await client.createDeploymentTarget({
    name: 'My New Kubernetes Deployment Target via SDK',
    type: 'kubernetes',
    namespace: 'glasskube',
    scope: 'namespace',
  });
  log(newKubernetesDeploymentTarget, 'create kubernetes deployment target');

  await client.createOrUpdateDeployment({
    applicationVersionId: newDockerVersion.id!,
    deploymentTargetId: newDockerDeploymentTarget.id!,
  });
  const recentlyDeployedTo = await client.getDeploymentTarget(newDockerDeploymentTarget.id!);
  log(recentlyDeployedTo, 'get recently deployed to');

  const deploymentTargets = await client.getDeploymentTargets();
  log(deploymentTargets, 'get deployment targets');
  for (let dt of deploymentTargets) {
    const deploymentTarget = await client.getDeploymentTarget(dt.id!);
    log(deploymentTarget, 'get deployment target by id');
  }
} catch (error) {
  console.error(error);
}

function log(obj: any, title?: string) {
  if (title) {
    console.log(title);
  }
  console.log(JSON.stringify(obj, null, 2));
  console.log('-------------------');
}

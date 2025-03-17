[**@glasskube/distr-sdk**](../README.md)

---

[@glasskube/distr-sdk](../README.md) / DistrService

# Class: DistrService

The DistrService provides a higher-level API for the Distr API. It allows to create and update deployments, check
if a deployment is outdated, and get the latest version of an application according to a specified strategy.
Under the hood it uses the low-level [Client](Client.md).

## Constructors

### new DistrService()

> **new DistrService**(`config`, `latestVersionStrategy`): `DistrService`

Creates a new DistrService instance. A client config containing an API key must be provided, optionally the API
base URL can be set. Optionally, a strategy for determining the latest version of an application can be specified –
the default is semantic versioning.

#### Parameters

##### config

`ConditionalPartial`\<[`ClientConfig`](../type-aliases/ClientConfig.md), `"apiBase"`\>

ClientConfig containing at least an API key and optionally an API base URL

##### latestVersionStrategy

[`LatestVersionStrategy`](../type-aliases/LatestVersionStrategy.md) = `'semver'`

Strategy for determining the latest version of an application (default: 'semver')

#### Returns

`DistrService`

## Methods

### createDeployment()

> **createDeployment**(`params`): `Promise`\<[`CreateDeploymentResult`](../type-aliases/CreateDeploymentResult.md)\>

Creates a new deployment target and deploys the given application version to it.

- If deployment type is 'kubernetes', the namespace and scope must be provided.
- If deployment type is 'kubernetes', the helm release name and values YAML can be provided.
- If no application version ID is given, the latest version of the application will be deployed.

#### Parameters

##### params

[`CreateDeploymentParams`](../type-aliases/CreateDeploymentParams.md)

#### Returns

`Promise`\<[`CreateDeploymentResult`](../type-aliases/CreateDeploymentResult.md)\>

---

### createDockerApplicationVersion()

> **createDockerApplicationVersion**(`applicationId`, `name`, `data`): `Promise`\<[`ApplicationVersion`](../interfaces/ApplicationVersion.md)\>

Creates a new application version for the given docker application using a Docker Compose file and an
optional template file.

#### Parameters

##### applicationId

`string`

##### name

`string`

Name of the new version

##### data

###### composeFile

`string`

###### templateFile?

`string`

#### Returns

`Promise`\<[`ApplicationVersion`](../interfaces/ApplicationVersion.md)\>

---

### createKubernetesApplicationVersion()

> **createKubernetesApplicationVersion**(`applicationId`, `versionName`, `data`): `Promise`\<[`ApplicationVersion`](../interfaces/ApplicationVersion.md)\>

Creates a new application version for the given Kubernetes application using a Helm chart.

#### Parameters

##### applicationId

`string`

##### versionName

`string`

##### data

###### baseValuesFile?

`string`

###### chartName?

`string`

###### chartType

[`HelmChartType`](../type-aliases/HelmChartType.md)

###### chartUrl

`string`

###### chartVersion

`string`

###### templateFile?

`string`

#### Returns

`Promise`\<[`ApplicationVersion`](../interfaces/ApplicationVersion.md)\>

---

### getLatestVersion()

> **getLatestVersion**(`appId`): `Promise`\<`undefined` \| [`ApplicationVersion`](../interfaces/ApplicationVersion.md)\>

Returns the latest version of the given application according to the specified strategy.

#### Parameters

##### appId

`string`

#### Returns

`Promise`\<`undefined` \| [`ApplicationVersion`](../interfaces/ApplicationVersion.md)\>

---

### getNewerVersions()

> **getNewerVersions**(`appId`, `currentVersionId`?): `Promise`\<\{ `app`: [`Application`](../interfaces/Application.md); `newerVersions`: [`ApplicationVersion`](../interfaces/ApplicationVersion.md)[]; \}\>

Returns the application and all versions that are newer than the given version ID. If no version ID is given,
all versions are considered. The versions are ordered ascending according to the given strategy.

#### Parameters

##### appId

`string`

##### currentVersionId?

`string`

#### Returns

`Promise`\<\{ `app`: [`Application`](../interfaces/Application.md); `newerVersions`: [`ApplicationVersion`](../interfaces/ApplicationVersion.md)[]; \}\>

---

### isOutdated()

> **isOutdated**(`deploymentTargetId`): `Promise`\<[`IsOutdatedResult`](../type-aliases/IsOutdatedResult.md)\>

Checks if the given deployment target is outdated, i.e. if there is a newer version of the application available.
The result additionally contains versions that are newer than the currently deployed one, ordered ascending.

#### Parameters

##### deploymentTargetId

`string`

#### Returns

`Promise`\<[`IsOutdatedResult`](../type-aliases/IsOutdatedResult.md)\>

---

### updateDeployment()

> **updateDeployment**(`params`): `Promise`\<`void`\>

Updates the deployment of an existing deployment target. If no application version ID is given, the latest version
of the already deployed application will be deployed.

#### Parameters

##### params

[`UpdateDeploymentParams`](../type-aliases/UpdateDeploymentParams.md)

#### Returns

`Promise`\<`void`\>

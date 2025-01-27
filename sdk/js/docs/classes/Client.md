[**@glasskube/distr-sdk**](../README.md)

***

[@glasskube/distr-sdk](../README.md) / Client

# Class: Client

Defined in: [client/client.ts:21](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/client/client.ts#L21)

## Constructors

### new Client()

> **new Client**(`config`): [`Client`](Client.md)

Defined in: [client/client.ts:24](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/client/client.ts#L24)

#### Parameters

##### config

`ConditionalPartial`\<[`ClientConfig`](../type-aliases/ClientConfig.md), `"apiBase"`\>

#### Returns

[`Client`](Client.md)

## Methods

### createAccessForDeploymentTarget()

> **createAccessForDeploymentTarget**(`deploymentTargetId`): `Promise`\<[`DeploymentTargetAccessResponse`](../interfaces/DeploymentTargetAccessResponse.md)\>

Defined in: [client/client.ts:91](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/client/client.ts#L91)

#### Parameters

##### deploymentTargetId

`string`

#### Returns

`Promise`\<[`DeploymentTargetAccessResponse`](../interfaces/DeploymentTargetAccessResponse.md)\>

***

### createApplication()

> **createApplication**(`application`): `Promise`\<[`Application`](../interfaces/Application.md)\>

Defined in: [client/client.ts:39](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/client/client.ts#L39)

#### Parameters

##### application

[`Application`](../interfaces/Application.md)

#### Returns

`Promise`\<[`Application`](../interfaces/Application.md)\>

***

### createApplicationVersion()

> **createApplicationVersion**(`applicationId`, `version`, `files`?): `Promise`\<[`ApplicationVersion`](../interfaces/ApplicationVersion.md)\>

Defined in: [client/client.ts:47](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/client/client.ts#L47)

#### Parameters

##### applicationId

`string`

##### version

[`ApplicationVersion`](../interfaces/ApplicationVersion.md)

##### files?

[`ApplicationVersionFiles`](../type-aliases/ApplicationVersionFiles.md)

#### Returns

`Promise`\<[`ApplicationVersion`](../interfaces/ApplicationVersion.md)\>

***

### createDeploymentTarget()

> **createDeploymentTarget**(`deploymentTarget`): `Promise`\<[`DeploymentTarget`](../interfaces/DeploymentTarget.md)\>

Defined in: [client/client.ts:83](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/client/client.ts#L83)

#### Parameters

##### deploymentTarget

[`DeploymentTarget`](../interfaces/DeploymentTarget.md)

#### Returns

`Promise`\<[`DeploymentTarget`](../interfaces/DeploymentTarget.md)\>

***

### createOrUpdateDeployment()

> **createOrUpdateDeployment**(`deploymentRequest`): `Promise`\<[`DeploymentRequest`](../interfaces/DeploymentRequest.md)\>

Defined in: [client/client.ts:87](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/client/client.ts#L87)

#### Parameters

##### deploymentRequest

[`DeploymentRequest`](../interfaces/DeploymentRequest.md)

#### Returns

`Promise`\<[`DeploymentRequest`](../interfaces/DeploymentRequest.md)\>

***

### getApplication()

> **getApplication**(`applicationId`): `Promise`\<[`Application`](../interfaces/Application.md)\>

Defined in: [client/client.ts:35](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/client/client.ts#L35)

#### Parameters

##### applicationId

`string`

#### Returns

`Promise`\<[`Application`](../interfaces/Application.md)\>

***

### getApplications()

> **getApplications**(): `Promise`\<[`Application`](../interfaces/Application.md)[]\>

Defined in: [client/client.ts:31](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/client/client.ts#L31)

#### Returns

`Promise`\<[`Application`](../interfaces/Application.md)[]\>

***

### getDeploymentTarget()

> **getDeploymentTarget**(`deploymentTargetId`): `Promise`\<[`DeploymentTarget`](../interfaces/DeploymentTarget.md)\>

Defined in: [client/client.ts:79](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/client/client.ts#L79)

#### Parameters

##### deploymentTargetId

`string`

#### Returns

`Promise`\<[`DeploymentTarget`](../interfaces/DeploymentTarget.md)\>

***

### getDeploymentTargets()

> **getDeploymentTargets**(): `Promise`\<[`DeploymentTarget`](../interfaces/DeploymentTarget.md)[]\>

Defined in: [client/client.ts:75](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/client/client.ts#L75)

#### Returns

`Promise`\<[`DeploymentTarget`](../interfaces/DeploymentTarget.md)[]\>

***

### updateApplication()

> **updateApplication**(`application`): `Promise`\<[`Application`](../interfaces/Application.md)\>

Defined in: [client/client.ts:43](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/client/client.ts#L43)

#### Parameters

##### application

[`Application`](../interfaces/Application.md)

#### Returns

`Promise`\<[`Application`](../interfaces/Application.md)\>

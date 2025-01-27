[**@glasskube/distr-sdk**](../README.md)

***

[@glasskube/distr-sdk](../README.md) / CreateDeploymentParams

# Type Alias: CreateDeploymentParams

> **CreateDeploymentParams**: `object`

Defined in: [client/service.ts:16](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/client/service.ts#L16)

## Type declaration

### application

> **application**: `object`

#### application.id?

> `optional` **id**: `string`

#### application.versionId?

> `optional` **versionId**: `string`

### kubernetesDeployment?

> `optional` **kubernetesDeployment**: `object`

#### kubernetesDeployment.releaseName

> **releaseName**: `string`

#### kubernetesDeployment.valuesYaml?

> `optional` **valuesYaml**: `string`

### target

> **target**: `object`

#### target.kubernetes?

> `optional` **kubernetes**: `object`

#### target.kubernetes.namespace

> **namespace**: `string`

#### target.kubernetes.scope

> **scope**: [`DeploymentTargetScope`](DeploymentTargetScope.md)

#### target.name

> **name**: `string`

#### target.type

> **type**: [`DeploymentType`](DeploymentType.md)

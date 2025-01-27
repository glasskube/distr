[**@glasskube/cloud-sdk**](../README.md)

***

[@glasskube/cloud-sdk](../README.md) / CreateDeploymentParams

# Type Alias: CreateDeploymentParams

> **CreateDeploymentParams**: `object`

Defined in: [client/service.ts:16](https://github.com/glasskube/distr/blob/80de58e6e72221ca696881996e5ae90ce94cd9cf/sdk/js/src/client/service.ts#L16)

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

[**@glasskube/distr-sdk**](../README.md)

***

[@glasskube/distr-sdk](../README.md) / DeploymentTarget

# Interface: DeploymentTarget

Defined in: [types/deployment-target.ts:7](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/types/deployment-target.ts#L7)

## Extends

- [`BaseModel`](BaseModel.md).[`Named`](Named.md)

## Properties

### agentVersion?

> `optional` **agentVersion**: [`AgentVersion`](AgentVersion.md)

Defined in: [types/deployment-target.ts:16](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/types/deployment-target.ts#L16)

***

### createdAt?

> `optional` **createdAt**: `string`

Defined in: [types/base.ts:3](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/types/base.ts#L3)

#### Inherited from

[`BaseModel`](BaseModel.md).[`createdAt`](BaseModel.md#createdat)

***

### createdBy?

> `optional` **createdBy**: [`UserAccountWithRole`](UserAccountWithRole.md)

Defined in: [types/deployment-target.ts:13](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/types/deployment-target.ts#L13)

***

### currentStatus?

> `optional` **currentStatus**: [`DeploymentTargetStatus`](DeploymentTargetStatus.md)

Defined in: [types/deployment-target.ts:14](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/types/deployment-target.ts#L14)

***

### deployment?

> `optional` **deployment**: [`DeploymentWithLatestRevision`](DeploymentWithLatestRevision.md)

Defined in: [types/deployment-target.ts:15](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/types/deployment-target.ts#L15)

***

### geolocation?

> `optional` **geolocation**: [`Geolocation`](Geolocation.md)

Defined in: [types/deployment-target.ts:12](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/types/deployment-target.ts#L12)

***

### id?

> `optional` **id**: `string`

Defined in: [types/base.ts:2](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/types/base.ts#L2)

#### Inherited from

[`BaseModel`](BaseModel.md).[`id`](BaseModel.md#id)

***

### name

> **name**: `string`

Defined in: [types/deployment-target.ts:8](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/types/deployment-target.ts#L8)

#### Overrides

[`Named`](Named.md).[`name`](Named.md#name)

***

### namespace?

> `optional` **namespace**: `string`

Defined in: [types/deployment-target.ts:10](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/types/deployment-target.ts#L10)

***

### reportedAgentVersionId?

> `optional` **reportedAgentVersionId**: `string`

Defined in: [types/deployment-target.ts:17](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/types/deployment-target.ts#L17)

***

### scope?

> `optional` **scope**: [`DeploymentTargetScope`](../type-aliases/DeploymentTargetScope.md)

Defined in: [types/deployment-target.ts:11](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/types/deployment-target.ts#L11)

***

### type

> **type**: [`DeploymentType`](../type-aliases/DeploymentType.md)

Defined in: [types/deployment-target.ts:9](https://github.com/glasskube/distr/blob/1c5d885406264f4301a9de61610438b702cea814/sdk/js/src/types/deployment-target.ts#L9)

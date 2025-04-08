<h1 align="center">
  <a href="https://distr.sh/" target="_blank">
    <img alt="" src="frontend/ui/public/distr-logo.svg" style="height: 5em;">
  </a>
  <br>
  Distr
</h1>

<div align="center">

**Software Distribution Platform**

</div>

[![GitHub Repo stars](https://img.shields.io/github/stars/glasskube/distr?style=flat)](https://github.com/glasskube/distr)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Docs](https://img.shields.io/badge/docs-distr.sh-blue)](https://distr.sh/docs/?utm_source=github)

Distr is the easiest way to distribute enterprise software to customer-controlled or shared-responsibility environments.

## Main features

- View & manage all deployments, artifacts, connected agents, self-managed & BYOC customers via the **intuitive web UI**
- Let your customers control their deployments or download your artifacts via the **white-label customer portal**
- Distribute specific versions of your application to specific customers with our licensing feature
- Distribute OCI compatible artifacts (Docker images, Helm charts, Terraform modules) to your customers with our **integrated OCI registry**
- Access the API using our [**rich SDK**](#distr-sdk)
- Fully open-source and [self-hostable](#self-hosting)

Check out the hosted version at https://signup.distr.sh/

## About

Distr is an Open Source commercial software distribution platform designed for Independent Software Vendors (ISVs).
It supports ISVs in delivering their software solutions to end customers operating in complex, customer-controlled environments.

Use cases include:

- On-premises software deployments
- Bring Your Own Cloud (BYOC) automation
- Edge & Fleet management

Read more about Distr and our use cases at https://distr.sh/docs/getting-started/about/

## Architecture overview

```mermaid
architecture-beta
    group ctrl(cloud)[Distr Saas or Your Cloud]
    service hub(server)[Distr Hub] in ctrl
    service db(database)[PostgreSQL] in ctrl
    service oci(database)[Distr OCI Registry] in ctrl
    service s3(disk)[Object Storage] in ctrl
    oci:R -- L:hub
    db:T -- B:hub
    oci:B -- T:s3

    junction customerjunction

    hub:R <-- L:customerjunction
    customerjunction:T -- B:agent
    customerjunction:B -- T:client


    group agentcustomer(cloud)[Customer Cloud]
    service agent(internet)[Distr Agent] in agentcustomer
    service app(server)[Your Application] in agentcustomer
    agent:L --> R:app

    group ocicustomer(cloud)[Fully self managed customer]
    service client(internet)[OCI client] in ocicustomer
```

## Self-hosting

### Docker

The Distr Hub is distributed as a Docker image distributed via ghcr.io.
Check out [`deploy/docker`](deploy/docker) for our example deployment using Docker Compose.
To get started quickly, do the following:

```shell
mkdir distr && cd distr && curl -fsSL https://github.com/glasskube/distr/releases/latest/download/deploy-docker.tar.bz2 | tar -jx
# make necessary changes to the .env file
docker-compose up -d
```

### Kubernetes

Distr is also available as a Helm chart distributed via ghcr.io.
Check out [`deploy/charts/distr`](deploy/charts/distr) for our Helm Chart source code.
To install Distr in Kubernetes, simply run:

```shell
helm upgrade --install --wait --namespace distr --create-namespace \
  distr oci://ghcr.io/glasskube/charts/distr \
  --set postgresql.enabled=true --set minio.enabled=true
```

For a quick testing setup, you don't have to modify the values.
However, if you intend to use distr in production, please revisit all available configuration values and adapt them accordingly.
You can find them in the reference [values.yaml](https://artifacthub.io/packages/helm/distr/distr?modal=values) file.

<hr>

Register your first account at http://localhost:8080/register

The full self-hosting documentation is at https://distr.sh/docs/self-hosting/self-hosting/

Using Distr agents on macOS? Follow the [guide](https://distr.sh/docs/guides/distr-on-macos/) to get started.

### Building from source

To build Distr Hub from source, first ensure that the following build dependencies are installed:

- NodeJS (Version 22)
- Go (Version 1.24)
- Docker (when building the Docker images)

We recommend that you use [mise](https://mise.jdx.dev/) to install these tools, but you do don't have to.

All build tasks can be found in the [`Makefile`](Makefile), for example:

```shell
# Build the control plane
make build
# Build all docker images
make build-docker
```

### Local development

To run the Distr Hub locally, you need to clone the repository and run the following commands:

```shell
# Start the database and a mock SMTP server
docker-compose up -d
# Start Distr Hub
make run
```

Open your browser and navigate to [`http://localhost:8080/register`](http://localhost:8080/register) to register a user
and receive the E-Mail verification link via Mailpit on [`http://localhost:8025`](http://localhost:8025).

## Distr SDK

Interact with Distr directly from your application code using our first-party SDK.
The Distr SDK is currently available for JavaScript only, but more languages and frameworks are on the roadmap.
Let us know what you would like to see!

You can install the Distr SDK for JavaScript from [npmjs.org](https://npmjs.org/):

```shell
npm install --save @glasskube/distr-sdk
```

The full SDK documentation is at https://distr.sh/docs/integrations/sdk/

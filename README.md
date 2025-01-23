<h1 align="center">
  <a href="https://glasskube.cloud/" target="_blank">
    <img alt="Glasskube" src="frontend/cloud-ui/public/glasskube-logo.svg">
  </a>
  <br>
  Glasskube Cloud
</h1>

<div align="center">

**Software Distribution Platform**

</div>

![](https://img.shields.io/badge/build-passing-brightgreen)
![](https://img.shields.io/badge/build-passing-brightgreen)
![](https://img.shields.io/badge/build-passing-brightgreen)
![](https://img.shields.io/badge/build-passing-brightgreen)
![](https://img.shields.io/badge/build-passing-brightgreen)
![](https://img.shields.io/badge/build-passing-brightgreen)
![](https://img.shields.io/badge/build-passing-brightgreen)

Glasskube Cloud is the easiest way to distribute enterprise software.
Lorem Ipsum is simply dummy text of the printing and typesetting industry.
Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book.
It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged.

- Praesent pharetra sed lorem ac.
- Proin gravida lobortis enim, non.
- Cras quis tellus id sem.

Check out the hosted version at https://app.glasskube.cloud/

## Self-hosting

### Docker

The Glasskube Cloud backend and UI server is distributed as a Docker image.
Check out [`deploy/docker`](deploy/docker) for our example deployment using Docker Compose.
To get started quickly, do the following:

<!-- x-release-please-start-version -->

```shell
mkdir cloud && cd cloud && curl -fsSL https://github.com/glasskube/cloud/releases/download/0.12.0/deploy-docker.tar.bz2 | tar -jx
# make necessary changes to the .env file
docker-compose up -d
```

<!-- x-release-please-end -->

The full self-hosting documentation is at https://glasskube.dev/products/cloud/docs/self-hosting/

### Building from source

To build the Glasskube Cloud server from source, first ensure that the following build dependencies are installed:

- NodeJS (Version 22)
- Go (Version 1.23)
- Docker (when building the Docker images)

We recommend that you use [mise](https://mise.jdx.dev/) to install these tools, but you do don't have to.

All build tasks can be found in the [`Makefile`](Makefile), for example:

```shell
# Build the server
make build
# Build all docker images
make build-docker
```

### Local development

To run Glasskube Cloud locally you need to clone the repository and run the following commands:

```shell
# Start the database and a mock SMTP server
docker-compose up -d
# Start Glasskube Cloud
make run
```

Open your browser and navigate to [`http://localhost:8080`](http://localhost:8080) to register a user
and receive the E-Mail verification link via Mailpit on [`http://localhost:8025`](http://localhost:8025).

## Glasskube Cloud SDK

Interact with Glasskube Cloud directly from your application code using our first-party SDK.
The Glasskube Cloud SDK is currently available for JavaScript only, but more languages and frameworks are on the roadmap.
Let us know what you would like to see!

You can install the Glasskube Cloud SDK for JavaScript from [npmjs.org](https://npmjs.org/):

```shell
npm install --save @glasskube/cloud-sdk
```

The full SDK documentation is at https://glasskube.dev/products/cloud/docs/sdk/

# Glasskube Cloud - Software Distribution Platform

> The easiest way to distribute enterprise software

## Getting started

To run Glasskube Cloud locally you need to clone the repository and run the following commands:

```shell
docker-compose up -d # starts the database and a local mailserver
make run # starts Glasskube Cloud
```

Open your browser and navigate to [`http://localhost:8080`](http://localhost:8080) to register a user
and receive the E-Mail verification link via Mailpit on [`http://localhost:8025`](http://localhost:8025).

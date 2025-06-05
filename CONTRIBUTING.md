# Contributing

Thank you for your interest in contributing to Distr!

Distr is open-source software licensed under the [Apache 2.0 license](https://github.com/glasskube/distr/blob/main/LICENSE) and accepts contributions via GitHub pull requests.

## Communications

To avoid unnecessary and redundant work, please reach out before you start working on your contribution.

You can either create an issue on GitHub or join our [Discord](https://discord.gg/6qqBSAWZfW) server to get in touch with the community.

## How to run distr for development

To run the Distr Hub locally, you need to clone the repository and run the following commands:

```shell
# Start the database and a mock SMTP server
docker compose up -d
# Start Distr Hub
make run
```

Open your browser and navigate to [`http://localhost:8080/register`](http://localhost:8080/register) to register a user
and receive the E-Mail verification link via Mailpit on [`http://localhost:8025`](http://localhost:8025).

## Backporting bugfixes

If the `main` branch already contains changes that would warrant a major or minor version bump but there is need to create a patch release only, it is possible to backport commits by pushing to the relevant `v*.*.x` branch. For example, if a commit should be added to version 1.2.3, it must be pushed to the `v1.2.x` branch.

**Important:** Please keep in mind the following rules for backporting:

1. Do not backport changes that would require an inappropriate version bump. For example, do not add new features to the `v1.2.x` branch, only bugfixes.
2. Only backport changes that are already in `main`. Ideally, use `git cherry-pick`.

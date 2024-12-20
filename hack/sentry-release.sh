#! /bin/env sh

if [ -z "$SENTRY_AUTH_TOKEN" ]; then
  echo "SENTRY_AUTH_TOKEN is not set"
  exit 1
fi

if [ -z "$VERSION" ]; then
  echo "VERSION is not set"
  exit 1
fi

export SENTRY_ORG="glasskube"
export SENTRY_PROJECT="cloud-ui"

npx sentry-cli releases new "$VERSION"
npx sentry-cli releases set-commits "$VERSION" --auto
npx sentry-cli sourcemaps upload --release="$VERSION" internal/frontend/dist/cloud-ui/browser
npx sentry-cli releases finalize "$VERSION"

#! /bin/env sh

if [ -z "$SENTRY_AUTH_TOKEN" ]; then
  echo "SENTRY_AUTH_TOKEN is not set"
  exit 1
fi

if [ -z "$SENTRY_VERSION" ]; then
  echo "SENTRY_VERSION is not set"
  exit 1
fi

export SENTRY_ORG="glasskube"
export SENTRY_PROJECT="distr-frontend"

npx sentry-cli releases new "$SENTRY_VERSION"
npx sentry-cli releases set-commits "$SENTRY_VERSION" --auto
npx sentry-cli sourcemaps upload --release="$SENTRY_VERSION" internal/frontend/dist/ui/browser
npx sentry-cli releases finalize "$SENTRY_VERSION"

FROM node:22.12.0-alpine AS frontend
WORKDIR /workspace

COPY package-lock.json .
COPY package.json .
RUN npm ci

COPY frontend/ frontend/
COPY angular.json .
COPY tailwind.config.js .
COPY tsconfig.json .
RUN npm run build:prod

FROM golang:1.23 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY cmd/ cmd/
COPY api/ api/
# doesn't exist (yet?)
# COPY pkg/ pkg/
COPY internal/ internal/
COPY --from=frontend /workspace/internal/frontend/dist internal/frontend/dist
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o cloud ./cmd/

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM cgr.dev/chainguard/static:latest@sha256:5ff428f8a48241b93a4174dbbc135a4ffb2381a9e10bdbbc5b9db145645886d5
WORKDIR /
COPY --from=builder /workspace/cloud .
USER 65532:65532

ENTRYPOINT ["/cloud"]

GOCMD ?= go
VERSION ?= snapshot
COMMIT = $(shell git rev-parse --short HEAD)
LDFLAGS ?= -s -w -X github.com/glasskube/distr/internal/buildconfig.version=$(VERSION) -X github.com/glasskube/distr/internal/buildconfig.commit=$(COMMIT)

.PHONY: tidy
tidy:
	$(GOCMD) mod tidy

.PHONY: validate-migrations
validate-migrations:
	hack/validate-migrations.sh

.PHONY: lint-frontend
lint-frontend:
	pnpm run lint

.PHONY: lint-frontend-fix
lint-frontend-fix:
	pnpm run format

.PHONY: lint-go
lint-go: tidy
	golangci-lint run

.PHONY: lint-go-fix
lint-go-fix:
	golangci-lint run --fix

.PHONY: lint
lint: lint-go lint-frontend validate-migrations

.PHONY: lint-fix
lint-fix: lint-go-fix lint-frontend-fix validate-migrations

node_modules: package.json pnpm-lock.yaml
	pnpm install
	@touch node_modules

.PHONY: frontend-dev
frontend-dev: node_modules
	pnpm run build:dev

.PHONY: frontend-prod
frontend-prod: node_modules
	npm run build:prod

.PHONY: run
run: frontend-dev tidy
	DISTR_ENV=.env.development.local CGO_ENABLED=0 $(GOCMD) run -ldflags="$(LDFLAGS)" ./cmd/hub/ serve

.PHONY: run-kubernetes-agent
run-kubernetes-agent: tidy
	CGO_ENABLED=0 $(GOCMD) run -ldflags="$(LDFLAGS)" ./cmd/agent/kubernetes

.PHONY: run-docker-agent
run-docker-agent: tidy
	CGO_ENABLED=0 $(GOCMD) run -ldflags="$(LDFLAGS)" ./cmd/agent/docker

.PHONY: build
build: frontend-prod tidy
	CGO_ENABLED=0 $(GOCMD) build -ldflags="$(LDFLAGS)" -o dist/distr ./cmd/hub/

.PHONY: build-docker-agent
build-docker-agent: tidy
	CGO_ENABLED=0 $(GOCMD) build -ldflags="$(LDFLAGS)" -o dist/docker-agent ./cmd/agent/docker

.PHONY: build-kubernetes-agent
build-kubernetes-agent: tidy
	CGO_ENABLED=0 $(GOCMD) build -ldflags="$(LDFLAGS)" -o dist/kubernetes-agent ./cmd/agent/kubernetes

.PHONY: build-mcp-server
build-mcp-server: tidy
	CGO_ENABLED=0 $(GOCMD) build -ldflags="$(LDFLAGS)" -o dist/mcp-server ./cmd/mcp

.PHONY: docker-build-hub
docker-build-hub: build
	docker build -f Dockerfile.hub --tag ghcr.io/glasskube/distr:$(VERSION) .

.PHONY: docker-build-docker-agent
docker-build-docker-agent:
	docker build -f Dockerfile.docker-agent --tag ghcr.io/glasskube/distr/docker-agent:$(VERSION) --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --network host .

.PHONY: docker-build-kubernetes-agent
docker-build-kubernetes-agent:
	docker build -f Dockerfile.kubernetes-agent --tag ghcr.io/glasskube/distr/kubernetes-agent:$(VERSION) --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --network host .

.PHONY: docker-build
docker-build: docker-build-hub docker-build-docker-agent docker-build-kubernetes-agent

.PHONY: init-db
init-db:
	DISTR_ENV=.env.development.local CGO_ENABLED=0 $(GOCMD) run -ldflags="$(LDFLAGS)" ./cmd/hub/ migrate

.PHONY: purge-db
purge-db:
	DISTR_ENV=.env.development.local CGO_ENABLED=0 $(GOCMD) run -ldflags="$(LDFLAGS)" ./cmd/hub/ migrate --down

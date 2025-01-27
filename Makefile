GOCMD ?= go
VERSION ?= snapshot
COMMIT = $(shell git rev-parse --short HEAD)
LDFLAGS ?= -s -w -X github.com/glasskube/distr/internal/buildconfig.version=$(VERSION) -X github.com/glasskube/distr/internal/buildconfig.commit=$(COMMIT)

.PHONY: tidy
tidy:
	$(GOCMD) mod tidy

.PHONY: lint-frontend
lint-frontend:
	npm run lint

.PHONY: lint-frontend-fix
lint-frontend-fix:
	npm run lint:fix

.PHONY: lint-go tidy
lint-go:
	golangci-lint run

.PHONY: lint-go-fix
lint-go-fix:
	golangci-lint run --fix

.PHONY: lint
lint: lint-go lint-frontend

.PHONY: lint-fix
lint-fix: lint-go-fix lint-frontend-fix

node_modules: package-lock.json
	npm install --no-save
	@touch node_modules

.PHONY: frontend-dev
frontend-dev: node_modules
	npm run build:dev

.PHONY: frontend-prod
frontend-prod: node_modules
	npm run build:prod

.PHONY: run
run: frontend-dev tidy
	DISTR_ENV=.env.development.local CGO_ENABLED=0 $(GOCMD) run -ldflags="$(LDFLAGS)" ./cmd/hub/

.PHONY: run-kubernetes-agent
run-kubernetes-agent: tidy
	CGO_ENABLED=0 $(GOCMD) run -ldflags="$(LDFLAGS)" ./cmd/agent/kubernetes

.PHONY: build
build: frontend-prod tidy
	CGO_ENABLED=0 $(GOCMD) build -ldflags="$(LDFLAGS)" -o dist/distr ./cmd/hub/

.PHONY: run-kubernetes-agent
build-kubernetes-agent: tidy
	CGO_ENABLED=0 $(GOCMD) build -ldflags="$(LDFLAGS)" -o dist/kubernetes-agent ./cmd/agent/kubernetes

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
docker-build: docker-build-server docker-build-docker-agent docker-build-kubernetes-agent

.PHONY: init-db
init-db:
	 go run ./cmd/hub/migrate/

.PHONY: purge-db
purge-db:
	 go run ./cmd/hub/migrate/ --down

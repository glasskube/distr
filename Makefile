GOCMD ?= go
COMMIT = $(shell git rev-parse --short HEAD)
LDFLAGS ?= -X github.com/glasskube/cloud/internal/buildconfig.version=$(VERSION) -X github.com/glasskube/cloud/internal/buildconfig.commit=$(COMMIT)

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
	CGO_ENABLED=0 $(GOCMD) run -ldflags="$(LDFLAGS)" ./cmd/cloud/

.PHONY: build
build: frontend-prod tidy
	CGO_ENABLED=0 $(GOCMD) build -ldflags="$(LDFLAGS)" -o dist/cloud ./cmd/cloud/

.PHONY: docker-build
docker-build: build
	docker build -f Dockerfile.server --tag ghcr.io/glasskube/cloud .

.PHONY: docker-build-agent
docker-build-agent:
	docker build -f Dockerfile.agent --tag ghcr.io/glasskube/cloud-agent --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --network host .

.PHONY: init-db
init-db:
	 go run ./cmd/cloud/migrate/up/

.PHONY: purge-db
purge-db:
	 go run ./cmd/cloud/migrate/down/

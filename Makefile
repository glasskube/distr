GOCMD ?= go

.PHONY: lint-frontend
lint-frontend:
	npm run lint

.PHONY: lint-go
lint-go:
	golangci-lint run

.PHONY: lint
lint: lint-go lint-frontend

.PHONY: frontend-dev
frontend-dev:
	npm run build:dev

.PHONY: frontend-prod
frontend-prod:
	npm run build:prod

.PHONY: run
run: frontend-dev
	$(GOCMD) run ./cmd/

.PHONY: build
build: frontend-prod
	$(GOCMD) build -o dist/cloud ./cmd/

.PHONY: docker-build
docker-build:
	docker build . --tag cloud  --network host

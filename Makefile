GOCMD ?= go

.PHONY: lint-frontend
lint-frontend:
	npm run lint

.PHONY: lint-go
lint-go:
	golangci-lint run

.PHONY: lint
lint: lint-go lint-frontend

.PHONY: frontend
frontend:
	npm run build

.PHONY: run
run: frontend
	$(GOCMD) run ./cmd/

.PHONY: build
build: frontend
	$(GOCMD) build -o dist/cloud ./cmd/

.PHONY: docker-build
docker-build:
	docker build . --tag cloud  --network host

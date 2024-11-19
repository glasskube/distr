GOCMD ?= go

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

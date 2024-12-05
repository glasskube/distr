GOCMD ?= go

.PHONY: lint-frontend
lint-frontend:
	npm run lint

.PHONY: lint-frontend-fix
lint-frontend-fix:
	npm run lint:fix

.PHONY: lint-go
lint-go:
	golangci-lint run

.PHONY: lint-go-fix
lint-go-fix:
	golangci-lint run --fix

.PHONY: lint
lint: lint-go lint-frontend

.PHONY: lint-fix
lint-fix: lint-go-fix lint-frontend-fix

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

.PHONY: docker-build-agent
docker-build-agent:
	docker build -f Dockerfile.agent . --tag glasskube-agent --network host

.PHONY: init-db
init-db:
	 cat sql/init_db.sql sql/test_data.sql | docker compose exec -T postgres psql --dbname glasskube --user local

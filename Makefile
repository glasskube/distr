GOCMD ?= go

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
	$(GOCMD) run ./cmd/cloud/

.PHONY: build
build: frontend-prod tidy
	$(GOCMD) build -o dist/cloud ./cmd/cloud/

.PHONY: docker-build
docker-build:
	docker build -f Dockerfile.server --tag ghcr.io/glasskube/cloud  --network host .

.PHONY: docker-build-agent
docker-build-agent:
	docker build -f Dockerfile.agent --tag ghcr.io/glasskube/cloud-agent --network host .

.PHONY: init-db
init-db:
	 cat sql/init_db.sql | docker compose exec -T postgres psql --dbname glasskube --user local

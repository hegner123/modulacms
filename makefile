GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
AMD_BINARY_NAME=modulacms-amd
X86_BINARY_NAME=modulacms-x86
VERSION?=0.0.0
SERVICE_PORT?=3000
DOCKER_REGISTRY?= #if set it should finished by /
EXPORT_RESULT?=false # for CI please set EXPORT_RESULT to true

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: all test build vendor test-development check docker-up docker-dev docker-infra docker-down docker-reset docker-destroy docker-logs docker-build docker-release

all: help

## Test:
test: ## Run the tests of the project
	touch testdb/create_tests.db	
	touch ./testdb/testing2348263.db
	rm ./testdb/*.db
	
	touch ./backups/tmp.zip
	rm ./backups/*.zip
	$(GOTEST) -v ./... 
	rm ./testdb/*.db

template-test: ## Run the template test
	$(GOTEST) -run TestServeTemplate  -outputdir tests

test-development: ## Run tests for the development package
	$(GOTEST) -v ./internal/development 

coverage: ## Run the tests of the project and export the coverage
	$(GOTEST) -cover -covermode=count -coverprofile=profile.cov ./...
	$(GOCMD) tool cover -func profile.cov
ifeq ($(EXPORT_RESULT), true)
	GO111MODULE=off go get -u github.com/AlekSi/gocov-xml
	GO111MODULE=off go get -u github.com/axw/gocov/gocov
	gocov convert profile.cov | gocov-xml > coverage.xml
endif

## Dev
check: ## Check compilation of cmd and internal packages (no artifacts)
	@GO111MODULE=on $(GOCMD) build -mod vendor -o /dev/null ./cmd && \
		$(GOCMD) build -mod vendor ./internal/... && \
		echo "$(GREEN)Build check passed$(RESET)"

dev: ## Prepare binaries and templates in src dir for faster iteration
	echo "" > debug.log
	$(eval VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev"))
	$(eval COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown"))
	$(eval BUILD_DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S'))
	GO111MODULE=on $(GOCMD) build -mod vendor \
		-ldflags="-X 'github.com/hegner123/modulacms/internal/utility.Version=$(VERSION)' \
		-X 'github.com/hegner123/modulacms/internal/utility.GitCommit=$(COMMIT)' \
		-X 'github.com/hegner123/modulacms/internal/utility.BuildDate=$(BUILD_DATE)'" \
		-o $(X86_BINARY_NAME) ./cmd

run: dev ## Build and run the application
	./$(X86_BINARY_NAME)

## Build:
build: ## Build your project and put the output binary in out/bin/
	$(eval VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev"))
	$(eval COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown"))
	$(eval BUILD_DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S'))
	GO111MODULE=on $(GOCMD) build -mod vendor \
		-ldflags="-X 'github.com/hegner123/modulacms/internal/utility.Version=$(VERSION)' \
		-X 'github.com/hegner123/modulacms/internal/utility.GitCommit=$(COMMIT)' \
		-X 'github.com/hegner123/modulacms/internal/utility.BuildDate=$(BUILD_DATE)'" \
		-o out/bin/$(X86_BINARY_NAME) ./cmd

## Dump:
dump: ## Dump sqlite db to sql
	sqlite3 modula.db .dump > modula_db.sql


## Deploy:
deploy:
	rsync -av --delete out/ modula:/root/app/modula

## Release:
release:
	echo "Release update placeholder"
	
clean: ## Remove build related file
	rm -fr ./bin
	rm -fr ./out
	rm -f ./junit-report.xml checkstyle-report.xml ./coverage.xml ./profile.cov yamllint-checkstyle.xml

vendor: ## Copy of all packages needed to support builds and tests in the vendor directory
	$(GOCMD) mod vendor

watch: ## Run the code with cosmtrek/air to have automatic reload on changes
	$(eval PACKAGE_NAME=$(shell head -n 1 go.mod | cut -d ' ' -f2))
	docker run -it --rm -w /go/src/$(PACKAGE_NAME) -v $(shell pwd):/go/src/$(PACKAGE_NAME) -p $(SERVICE_PORT):$(SERVICE_PORT) cosmtrek/air

## SQL
sqlc: ## Run sqlc generate in sql directory
	cd ./sql && sqlc generate && echo "generated coded successfully"

	
## Lint:
lint: lint-go lint-dockerfile lint-yaml ## Run all available linters

lint-dockerfile: ## Lint your Dockerfile
# If dockerfile is present we lint it.
ifeq ($(shell test -e ./Dockerfile && echo -n yes),yes)
	$(eval CONFIG_OPTION = $(shell [ -e $(shell pwd)/.hadolint.yaml ] && echo "-v $(shell pwd)/.hadolint.yaml:/root/.config/hadolint.yaml" || echo "" ))
	$(eval OUTPUT_OPTIONS = $(shell [ "${EXPORT_RESULT}" == "true" ] && echo "--format checkstyle" || echo "" ))
	$(eval OUTPUT_FILE = $(shell [ "${EXPORT_RESULT}" == "true" ] && echo "| tee /dev/tty > checkstyle-report.xml" || echo "" ))
	docker run --rm -i $(CONFIG_OPTION) hadolint/hadolint hadolint $(OUTPUT_OPTIONS) - < ./Dockerfile $(OUTPUT_FILE)
endif

lint-go: ## Use golintci-lint on your project
	$(eval OUTPUT_OPTIONS = $(shell [ "${EXPORT_RESULT}" == "true" ] && echo "--out-format checkstyle ./... | tee /dev/tty > checkstyle-report.xml" || echo "" ))
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:latest-alpine golangci-lint run --deadline=65s $(OUTPUT_OPTIONS)

lint-yaml: ## Use yamllint on the yaml file of your projects
ifeq ($(EXPORT_RESULT), true)
	GO111MODULE=off go get -u github.com/thomaspoignant/yamllint-checkstyle
	$(eval OUTPUT_OPTIONS = | tee /dev/tty | yamllint-checkstyle > yamllint-checkstyle.xml)
endif
	docker run --rm -it -v $(shell pwd):/data cytopia/yamllint -f parsable $(shell git ls-files '*.yml' '*.yaml') $(OUTPUT_OPTIONS)

## Docker:
COMPOSE_FILE=deploy/docker/docker-compose.yml

docker-up: ## Start full stack (CMS + databases + MinIO)
	DOCKER_BUILDKIT=1 docker compose -f $(COMPOSE_FILE) up -d --build

docker-dev: ## Rebuild and restart CMS only (incremental via BuildKit cache)
	DOCKER_BUILDKIT=1 docker compose -f $(COMPOSE_FILE) up -d --build modulacms

docker-infra: ## Start infrastructure only (postgres, mysql, minio)
	docker compose -f $(COMPOSE_FILE) up -d postgres mysql minio

docker-down: ## Stop all containers, keep volumes
	docker compose -f $(COMPOSE_FILE) down

docker-reset: ## Stop all containers and delete volumes
	docker compose -f $(COMPOSE_FILE) down -v

docker-destroy: ## Remove all project containers, volumes, and images
	docker compose -f $(COMPOSE_FILE) down -v --rmi all

docker-logs: ## Tail CMS container logs
	docker compose -f $(COMPOSE_FILE) logs -f modulacms

docker-build: ## Build standalone CMS image (for CI)
	DOCKER_BUILDKIT=1 docker build --rm --tag modulacms .

docker-release: ## Release the container with tag latest and version
	docker tag modulacms $(DOCKER_REGISTRY)modulacms:latest
	docker tag modulacms $(DOCKER_REGISTRY)modulacms:$(VERSION)
	docker push $(DOCKER_REGISTRY)modulacms:latest
	docker push $(DOCKER_REGISTRY)modulacms:$(VERSION)

## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

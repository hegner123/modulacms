# ModulaCMS Justfile

gocmd := "go"
gotest := gocmd + " test"
govet := gocmd + " vet"
amd_binary_name := "modulacms-amd"
x86_binary_name := "modulacms-x86"
version := env_var_or_default("VERSION", "0.0.0")
service_port := env_var_or_default("SERVICE_PORT", "3000")
docker_registry := env_var_or_default("DOCKER_REGISTRY", "")
export_result := env_var_or_default("EXPORT_RESULT", "false")
compose_file := "deploy/docker/docker-compose.yml"
dealer_compose := "docker compose -p modulacms-dealer"

# Show available recipes
default:
    @just --list --unsorted

# [Test] Run the tests of the project
test:
    touch testdb/create_tests.db
    touch ./testdb/testing2348263.db
    rm ./testdb/*.db
    touch ./backups/tmp.zip
    rm ./backups/*.zip
    {{gotest}} -v ./...
    rm ./testdb/*.db

# [Test] Run the template test
template-test:
    {{gotest}} -run TestServeTemplate -outputdir tests

# [Test] Run tests for the development package
test-development:
    {{gotest}} -v ./internal/development

# [Test] Run the tests and export coverage
coverage:
    {{gotest}} -cover -covermode=count -coverprofile=profile.cov ./...
    {{gocmd}} tool cover -func profile.cov

# [Dev] Check compilation of cmd and internal packages (no artifacts)
check:
    GO111MODULE=on {{gocmd}} build -mod vendor -o /dev/null ./cmd
    {{gocmd}} build -mod vendor ./internal/...
    @echo "Build check passed"

# [Dev] Build local x86 binary for development
dev:
    #!/usr/bin/env bash
    echo "" > debug.log
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    BUILD_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S')
    GO111MODULE=on {{gocmd}} build -mod vendor \
        -ldflags="-X 'github.com/hegner123/modulacms/internal/utility.Version=${VERSION}' \
        -X 'github.com/hegner123/modulacms/internal/utility.GitCommit=${COMMIT}' \
        -X 'github.com/hegner123/modulacms/internal/utility.BuildDate=${BUILD_DATE}'" \
        -o {{x86_binary_name}} ./cmd

# [Dev] Build and run the application
run: dev
    ./{{x86_binary_name}}

# [Build] Build production binary to out/bin/
build:
    #!/usr/bin/env bash
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    BUILD_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S')
    GO111MODULE=on {{gocmd}} build -mod vendor \
        -ldflags="-X 'github.com/hegner123/modulacms/internal/utility.Version=${VERSION}' \
        -X 'github.com/hegner123/modulacms/internal/utility.GitCommit=${COMMIT}' \
        -X 'github.com/hegner123/modulacms/internal/utility.BuildDate=${BUILD_DATE}'" \
        -o out/bin/{{x86_binary_name}} ./cmd

# [Build] Remove build artifacts
clean:
    rm -fr ./bin
    rm -fr ./out
    rm -f ./junit-report.xml checkstyle-report.xml ./coverage.xml ./profile.cov yamllint-checkstyle.xml

# [Build] Update vendor directory
vendor:
    {{gocmd}} mod vendor

# [Build] Run the code with cosmtrek/air for automatic reload
watch:
    #!/usr/bin/env bash
    PACKAGE_NAME=$(head -n 1 go.mod | cut -d ' ' -f2)
    docker run -it --rm -w /go/src/${PACKAGE_NAME} -v $(pwd):/go/src/${PACKAGE_NAME} -p {{service_port}}:{{service_port}} cosmtrek/air

# [Dump] Dump sqlite db to sql
dump:
    sqlite3 modula.db .dump > modula_db.sql

# [Deploy] Deploy to remote server
deploy:
    rsync -av --delete out/ modula:/root/app/modula

# [Deploy] Release placeholder
release:
    echo "Release update placeholder"

# [SQL] Run sqlc generate in sql directory
sqlc:
    cd ./sql && sqlc generate && echo "generated code successfully"

# [Lint] Run all available linters
lint: lint-go lint-dockerfile lint-yaml

# [Lint] Lint Dockerfile with hadolint
lint-dockerfile:
    #!/usr/bin/env bash
    if [ -e ./Dockerfile ]; then
        CONFIG_OPTION=""
        if [ -e "$(pwd)/.hadolint.yaml" ]; then
            CONFIG_OPTION="-v $(pwd)/.hadolint.yaml:/root/.config/hadolint.yaml"
        fi
        OUTPUT_OPTIONS=""
        OUTPUT_FILE=""
        if [ "{{export_result}}" = "true" ]; then
            OUTPUT_OPTIONS="--format checkstyle"
            OUTPUT_FILE="| tee /dev/tty > checkstyle-report.xml"
        fi
        eval "docker run --rm -i ${CONFIG_OPTION} hadolint/hadolint hadolint ${OUTPUT_OPTIONS} - < ./Dockerfile ${OUTPUT_FILE}"
    fi

# [Lint] Lint Go code with golangci-lint
lint-go:
    #!/usr/bin/env bash
    OUTPUT_OPTIONS=""
    if [ "{{export_result}}" = "true" ]; then
        OUTPUT_OPTIONS="--out-format checkstyle ./... | tee /dev/tty > checkstyle-report.xml"
    fi
    eval "docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:latest-alpine golangci-lint run --deadline=65s ${OUTPUT_OPTIONS}"

# [Lint] Lint YAML files
lint-yaml:
    #!/usr/bin/env bash
    OUTPUT_OPTIONS=""
    if [ "{{export_result}}" = "true" ]; then
        GO111MODULE=off go get -u github.com/thomaspoignant/yamllint-checkstyle
        OUTPUT_OPTIONS="| tee /dev/tty | yamllint-checkstyle > yamllint-checkstyle.xml"
    fi
    eval "docker run --rm -it -v $(pwd):/data cytopia/yamllint -f parsable $(git ls-files '*.yml' '*.yaml') ${OUTPUT_OPTIONS}"

# [Docker] Start full stack (CMS + databases + MinIO)
docker-up:
    DOCKER_BUILDKIT=1 docker compose -f {{compose_file}} up -d --build

# [Docker] Rebuild and restart CMS only (incremental via BuildKit cache)
docker-dev:
    DOCKER_BUILDKIT=1 docker compose -f {{compose_file}} up -d --build modulacms

# [Docker] Start infrastructure only (postgres, mysql, minio)
docker-infra:
    docker compose -f {{compose_file}} up -d postgres mysql minio

# [Docker] Stop all containers, keep volumes
docker-down:
    docker compose -f {{compose_file}} down

# [Docker] Stop all containers and delete volumes
docker-reset:
    docker compose -f {{compose_file}} down -v

# [Docker] Remove all project containers, volumes, and images
docker-destroy:
    docker compose -f {{compose_file}} down -v --rmi all

# [Docker] Tail CMS container logs
docker-logs:
    docker compose -f {{compose_file}} logs -f modulacms

# [Docker] Build standalone CMS image (for CI)
docker-build:
    DOCKER_BUILDKIT=1 docker build --rm --tag modulacms .

# [Docker] Release container with tag latest and version
docker-release:
    docker tag modulacms {{docker_registry}}modulacms:latest
    docker tag modulacms {{docker_registry}}modulacms:{{version}}
    docker push {{docker_registry}}modulacms:latest
    docker push {{docker_registry}}modulacms:{{version}}

# [Dealer] Start dealer CMS container
dealer-up:
    DOCKER_BUILDKIT=1 {{dealer_compose}} up -d --build

# [Dealer] Stop dealer container, keep volumes
dealer-down:
    {{dealer_compose}} down

# [Dealer] Stop dealer container and delete volumes
dealer-reset:
    {{dealer_compose}} down -v

# [Dealer] Remove dealer container, volumes, and images
dealer-destroy:
    {{dealer_compose}} down -v --rmi all

# [Dealer] Tail dealer container logs
dealer-logs:
    {{dealer_compose}} logs -f modulacms

# [Dealer] Force rebuild dealer image and restart
dealer-rebuild:
    DOCKER_BUILDKIT=1 {{dealer_compose}} up -d --build --force-recreate

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
compose_file := "deploy/docker/docker-compose.full.yml"
compose_sqlite := "deploy/docker/docker-compose.sqlite.yml"
compose_mysql := "deploy/docker/docker-compose.mysql.yml"
compose_postgres := "deploy/docker/docker-compose.postgres.yml"
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

# [Test] Start MinIO container for integration tests
test-minio:
    docker compose -f {{compose_sqlite}} up -d minio

# [Test] Run S3 integration tests (requires MinIO running)
test-integration:
    {{gotest}} -tags integration -v -count=1 ./internal/media/ -run TestIntegration

# [Test] Stop MinIO after integration tests
test-minio-down:
    docker compose -f {{compose_sqlite}} down minio

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

# [SDK] Install TypeScript SDK dependencies
sdk-install:
    cd sdks/typescript && pnpm install

# [SDK] Build all TypeScript SDK packages
sdk-build:
    cd sdks/typescript && pnpm build

# [SDK] Run TypeScript SDK tests
sdk-test:
    cd sdks/typescript && pnpm test

# [SDK] Typecheck all TypeScript SDK packages
sdk-typecheck:
    cd sdks/typescript && pnpm typecheck

# [SDK] Clean TypeScript SDK build artifacts
sdk-clean:
    cd sdks/typescript && pnpm clean

# [SDK] Run Go SDK tests
sdk-go-test:
    cd sdks/go && go test -v ./...

# [SDK] Vet Go SDK
sdk-go-vet:
    cd sdks/go && go vet ./...

# [SDK] Build Swift SDK
sdk-swift-build:
    cd sdks/swift && swift build

# [SDK] Run Swift SDK tests
sdk-swift-test:
    cd sdks/swift && swift test

# [SDK] Clean Swift SDK build artifacts
sdk-swift-clean:
    cd sdks/swift && swift package clean

# [MCP] Build MCP server binary
mcp-build:
    cd mcp && go build -o modulacms-mcp .

# [MCP] Install MCP server binary to /usr/local/bin
mcp-install: mcp-build
    cp mcp/modulacms-mcp /usr/local/bin/modulacms-mcp

# [Plugin] List installed plugins
plugin-list:
    ./{{x86_binary_name}} plugin list

# [Plugin] Create a new plugin scaffold
plugin-init name:
    ./{{x86_binary_name}} plugin init {{name}}

# [Plugin] Validate a plugin
plugin-validate path:
    ./{{x86_binary_name}} plugin validate {{path}}

# [Plugin] Show plugin details (requires running server)
plugin-info name:
    ./{{x86_binary_name}} plugin info {{name}}

# [Plugin] Reload a plugin (requires running server)
plugin-reload name:
    ./{{x86_binary_name}} plugin reload {{name}}

# [Plugin] Enable a plugin (requires running server)
plugin-enable name:
    ./{{x86_binary_name}} plugin enable {{name}}

# [Plugin] Disable a plugin (requires running server)
plugin-disable name:
    ./{{x86_binary_name}} plugin disable {{name}}

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

# [Docker:SQLite] Start SQLite stack (CMS + MinIO)
docker-sqlite-up:
    DOCKER_BUILDKIT=1 docker compose -f {{compose_sqlite}} up -d --build

# [Docker:SQLite] Stop SQLite stack, keep volumes
docker-sqlite-down:
    docker compose -f {{compose_sqlite}} down

# [Docker:SQLite] Stop SQLite stack and delete volumes
docker-sqlite-reset:
    docker compose -f {{compose_sqlite}} down -v

# [Docker:SQLite] Rebuild and restart CMS only (keeps database intact)
docker-sqlite-dev:
    DOCKER_BUILDKIT=1 docker compose -f {{compose_sqlite}} up -d --build modulacms

# [Docker:SQLite] Wipe volumes and rebuild SQLite stack from scratch
docker-sqlite-fresh: docker-sqlite-reset docker-sqlite-up

# [Docker:SQLite] Tail SQLite stack CMS logs
docker-sqlite-logs:
    docker compose -f {{compose_sqlite}} logs -f modulacms

# [Docker:MySQL] Start MySQL stack (CMS + MySQL + MinIO)
docker-mysql-up:
    DOCKER_BUILDKIT=1 docker compose -f {{compose_mysql}} up -d --build

# [Docker:MySQL] Stop MySQL stack, keep volumes
docker-mysql-down:
    docker compose -f {{compose_mysql}} down

# [Docker:MySQL] Stop MySQL stack and delete volumes
docker-mysql-reset:
    docker compose -f {{compose_mysql}} down -v

# [Docker:MySQL] Rebuild and restart CMS only (keeps database intact)
docker-mysql-dev:
    DOCKER_BUILDKIT=1 docker compose -f {{compose_mysql}} up -d --build modulacms

# [Docker:MySQL] Wipe volumes and rebuild MySQL stack from scratch
docker-mysql-fresh: docker-mysql-reset docker-mysql-up

# [Docker:MySQL] Tail MySQL stack CMS logs
docker-mysql-logs:
    docker compose -f {{compose_mysql}} logs -f modulacms

# [Docker:Postgres] Start PostgreSQL stack (CMS + PostgreSQL + MinIO)
docker-postgres-up:
    DOCKER_BUILDKIT=1 docker compose -f {{compose_postgres}} up -d --build

# [Docker:Postgres] Stop PostgreSQL stack, keep volumes
docker-postgres-down:
    docker compose -f {{compose_postgres}} down

# [Docker:Postgres] Stop PostgreSQL stack and delete volumes
docker-postgres-reset:
    docker compose -f {{compose_postgres}} down -v

# [Docker:Postgres] Rebuild and restart CMS only (keeps database intact)
docker-postgres-dev:
    DOCKER_BUILDKIT=1 docker compose -f {{compose_postgres}} up -d --build modulacms

# [Docker:Postgres] Wipe volumes and rebuild PostgreSQL stack from scratch
docker-postgres-fresh: docker-postgres-reset docker-postgres-up

# [Docker:Postgres] Reset MinIO container and volumes only
docker-postgres-minio-reset:
    docker compose -f {{compose_postgres}} down minio
    docker compose -f {{compose_postgres}} up -d minio

# [Docker:Postgres] Tail PostgreSQL stack CMS logs
docker-postgres-logs:
    docker compose -f {{compose_postgres}} logs -f modulacms

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

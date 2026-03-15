# Modula Justfile

gocmd := "go"
gotest := gocmd + " test"
govet := gocmd + " vet"
amd_binary_name := "modula-amd"
x86_binary_name := "modula-x86"
version := env_var_or_default("VERSION", "0.0.0")
service_port := env_var_or_default("SERVICE_PORT", "3000")
docker_registry := env_var_or_default("DOCKER_REGISTRY", "")
export_result := env_var_or_default("EXPORT_RESULT", "false")
compose_file := "deploy/docker/docker-compose.full.yml"
compose_sqlite := "deploy/docker/docker-compose.sqlite.yml"
compose_mysql := "deploy/docker/docker-compose.mysql.yml"
compose_postgres := "deploy/docker/docker-compose.postgres.yml"
dealer_compose := "docker compose -p modula-dealer"
prod_host := "deploy@api.modulacms.com"
prod_compose := "deploy/docker/docker-compose.prod.yml"
prod_image := "modula-modula"

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

# [Test] Run cross-backend DB integration tests (requires MySQL + PostgreSQL via docker-infra)
test-integration-db:
    {{gotest}} -tags integration -v -count=1 -timeout 120s ./internal/db/ -run TestCrossBackend

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
    just admin bundle
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

# [Dev] Build and run with live CSS/JS from disk (no embed, no cache)
# Uses air for hot reload: rebuilds on .go and .templ changes, runs templ generate as pre_cmd
run-admin:
    air

# [Dev] Build with live static assets (CSS/JS served from disk)
dev-admin:
    #!/usr/bin/env bash
    mkdir -p out
    if [ ! -f out/modula.config.json ]; then
        cp modula.config.json out/modula.config.json
        echo "Copied modula.config.json to out/ — edit out/modula.config.json for dev settings"
    fi
    echo "" > debug.log
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    BUILD_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S')
    GO111MODULE=on {{gocmd}} build -mod vendor -tags dev \
        -ldflags="-X 'github.com/hegner123/modulacms/internal/utility.Version=${VERSION}' \
        -X 'github.com/hegner123/modulacms/internal/utility.GitCommit=${COMMIT}' \
        -X 'github.com/hegner123/modulacms/internal/utility.BuildDate=${BUILD_DATE}'" \
        -o out/modulacms-dev ./cmd
    codesign -s - out/modulacms-dev

# [Build] Build production binary to out/bin/
build:
    #!/usr/bin/env bash
    just admin bundle
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

# [Docs] Copy documentation to the website repo
doc-copy:
    cp -R documentation/ /Users/home/Documents/Code/Go_dev/cms/modulacms.com/documentation/ && echo "success"

# [Docs] Check documentation staleness against upstream (or explicit range)
doc-check *RANGE:
    @.githooks/doc-check {{RANGE}}

# [Docs] Install git hooks (.githooks/ directory)
install-hooks:
    git config core.hooksPath .githooks
    chmod +x .githooks/pre-push .githooks/doc-check
    @echo "Git hooks installed (core.hooksPath -> .githooks/)"

# [Dump] Dump sqlite db to sql
dump:
    sqlite3 modula.db .dump > modula_db.sql

# [Deploy] Deploy CMS to production (pull, build, health check, rollback on failure)
deploy:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Deploying to {{prod_host}}..."
    ssh {{prod_host}} 'bash -s' <<'DEPLOY'
    set -euo pipefail
    cd ~/modulacms

    echo "Pulling latest changes..."
    git pull

    echo "Updating vendor dependencies..."
    go mod vendor

    echo "Tagging current image for rollback..."
    docker tag {{prod_image}}:latest {{prod_image}}:previous 2>/dev/null || echo "No existing image to tag"

    echo "Building and starting new CMS container..."
    docker compose -f {{prod_compose}} up -d --build modula

    echo "Waiting for health check (up to 30s)..."
    for i in $(seq 1 15); do
        sleep 2
        HTTP_CODE=$(curl -Lkso /dev/null -w '%{http_code}' https://localhost/api/v1/health 2>/dev/null || echo "000")
        if [ "$HTTP_CODE" != "000" ] && [ "$HTTP_CODE" != "502" ] && [ "$HTTP_CODE" != "503" ]; then
            echo "Health check passed (HTTP $HTTP_CODE)"
            docker rmi {{prod_image}}:previous 2>/dev/null || true
            echo "Deploy successful!"
            exit 0
        fi
        echo "  Attempt $i/15: HTTP $HTTP_CODE"
    done

    echo "Health check failed after 30s! Rolling back..."
    docker compose -f {{prod_compose}} stop modula
    docker tag {{prod_image}}:previous {{prod_image}}:latest 2>/dev/null || true
    docker compose -f {{prod_compose}} up -d modula
    echo "Rolled back to previous version"
    exit 1
    DEPLOY

# [Deploy] Show production container status
status:
    ssh {{prod_host}} "cd ~/modulacms && docker compose -f {{prod_compose}} ps"

# [Deploy] Tail production CMS logs
logs:
    ssh -t {{prod_host}} "cd ~/modulacms && docker compose -f {{prod_compose}} logs -f --tail 100 modula"

# [Deploy] Rollback CMS to previous image
rollback:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Rolling back CMS on {{prod_host}}..."
    ssh {{prod_host}} 'bash -s' <<'ROLLBACK'
    set -euo pipefail
    cd ~/modulacms

    if ! docker image inspect {{prod_image}}:previous >/dev/null 2>&1; then
        echo "No previous image found. Nothing to roll back to."
        exit 1
    fi

    echo "Stopping current CMS container..."
    docker compose -f {{prod_compose}} stop modula

    echo "Restoring previous image..."
    docker tag {{prod_image}}:previous {{prod_image}}:latest

    echo "Starting CMS with previous image..."
    docker compose -f {{prod_compose}} up -d modula

    echo "Waiting for health check (up to 30s)..."
    for i in $(seq 1 15); do
        sleep 2
        HTTP_CODE=$(curl -Lkso /dev/null -w '%{http_code}' https://localhost/api/v1/health 2>/dev/null || echo "000")
        if [ "$HTTP_CODE" != "000" ] && [ "$HTTP_CODE" != "502" ] && [ "$HTTP_CODE" != "503" ]; then
            echo "Health check passed (HTTP $HTTP_CODE)"
            docker rmi {{prod_image}}:previous 2>/dev/null || true
            echo "Rollback successful!"
            exit 0
        fi
        echo "  Attempt $i/15: HTTP $HTTP_CODE"
    done

    echo "Rollback health check failed! Container may need manual intervention."
    exit 1
    ROLLBACK

# [Admin] Manage admin panel codegen: just admin <action>
# Actions: generate, watch, verify, bundle, bundle-watch, bundle-verify
admin action:
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{action}}" in
        generate)      templ generate && tailwindcss -i internal/admin/static/css/input.css -o internal/admin/static/css/tailwind.css --minify ;;
        watch)         templ generate --watch & tailwindcss -i internal/admin/static/css/input.css -o internal/admin/static/css/tailwind.css --watch ;;
        verify)        templ generate && tailwindcss -i internal/admin/static/css/input.css -o internal/admin/static/css/tailwind.css --minify && git diff --exit-code internal/admin/ ;;
        bundle)        esbuild internal/admin/static/js/block-editor-src/index.js --bundle --format=esm --banner:js="// AUTO-GENERATED — DO NOT EDIT. Source: block-editor-src/. Regenerate: just admin bundle" --outfile=internal/admin/static/js/block-editor.js ;;
        bundle-watch)  esbuild internal/admin/static/js/block-editor-src/index.js --bundle --format=esm --banner:js="// AUTO-GENERATED — DO NOT EDIT. Source: block-editor-src/. Regenerate: just admin bundle" --outfile=internal/admin/static/js/block-editor.js --watch ;;
        bundle-verify) just admin bundle && git diff --exit-code internal/admin/static/js/block-editor.js ;;
        test)          cd internal/admin/static/js/block-editor-src && npx vitest run ;;
        *)             echo "Unknown action: {{action}}"; echo "Actions: generate, watch, verify, bundle, bundle-watch, bundle-verify, test"; exit 1 ;;
    esac

# [Codegen] Generate sqlc.yml from shared definitions
sqlc-config:
    {{gocmd}} run ./tools/sqlcgen/...

# [Codegen] Verify sqlc.yml is up-to-date (for CI)
sqlc-config-verify:
    {{gocmd}} run ./tools/sqlcgen/... -verify

# [SQL] Generate sqlc.yml then run sqlc generate
sqlc: sqlc-config
    cd ./sql && sqlc generate && echo "generated code successfully"

# [Codegen] Generate db wrapper code from entity definitions
dbgen:
    {{gocmd}} run ./tools/dbgen/...

# [Codegen] Generate a single entity (e.g., just dbgen-entity Users)
dbgen-entity name:
    {{gocmd}} run ./tools/dbgen/... -entity {{name}}

# [Codegen] Verify generated files are up-to-date (for CI)
dbgen-verify:
    {{gocmd}} run ./tools/dbgen/... -verify

# [Codegen] Generate MySQL/PSQL sections in _custom.go files from SQLite source
drivergen:
    {{gocmd}} run ./tools/drivergen/...

# [Codegen] Generate a single custom file
drivergen-file file:
    {{gocmd}} run ./tools/drivergen/... {{file}}

# [Codegen] Verify generated custom sections are up-to-date (for CI)
drivergen-verify:
    {{gocmd}} run ./tools/drivergen/... -verify

# [Codegen] Generate non-admin files from admin source (admin is source of truth)
drivergen-admin:
    {{gocmd}} run ./tools/drivergen/... --mode admin

# [Codegen] Verify admin-generated files are up-to-date (for CI)
drivergen-admin-verify:
    {{gocmd}} run ./tools/drivergen/... --mode admin -verify

# [SDK] Run SDK command: just sdk <lang> <action>
# Langs: ts (install, build, test, typecheck, clean), go (test, vet), swift (build, test, clean)
sdk lang action:
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{lang}}" in
        ts)
            case "{{action}}" in
                install)   cd sdks/typescript && pnpm install ;;
                build)     cd sdks/typescript && pnpm build ;;
                test)      cd sdks/typescript && pnpm test ;;
                typecheck) cd sdks/typescript && pnpm typecheck ;;
                clean)     cd sdks/typescript && pnpm clean ;;
                *)         echo "Unknown ts action: {{action}}"; echo "Actions: install, build, test, typecheck, clean"; exit 1 ;;
            esac ;;
        go)
            case "{{action}}" in
                test) cd sdks/go && go test -v ./... ;;
                vet)  cd sdks/go && go vet ./... ;;
                *)    echo "Unknown go action: {{action}}"; echo "Actions: test, vet"; exit 1 ;;
            esac ;;
        swift)
            case "{{action}}" in
                build) cd sdks/swift && swift build ;;
                test)  cd sdks/swift && swift test ;;
                clean) cd sdks/swift && swift package clean ;;
                *)     echo "Unknown swift action: {{action}}"; echo "Actions: build, test, clean"; exit 1 ;;
            esac ;;
        *)
            echo "Unknown lang: {{lang}}"; echo "Langs: ts, go, swift"; exit 1 ;;
    esac

# [MCP] The MCP server is built into the main binary. Use: modula mcp
# Legacy standalone build (mcp/ directory) is deprecated.
mcp-build:
    @echo "MCP server is now built into the main binary. Use: modula mcp"

mcp-install:
    @echo "MCP server is now built into the main binary. Use: modula mcp"

# [Plugin] Manage plugins: just plugin <action> [name/path]
# Actions: list, init, validate, info, reload, enable, disable
plugin action arg='':
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{action}}" in
        list)     ./{{x86_binary_name}} plugin list ;;
        init)     [ -z "{{arg}}" ] && echo "Usage: just plugin init <name>" && exit 1; ./{{x86_binary_name}} plugin init "{{arg}}" ;;
        validate) [ -z "{{arg}}" ] && echo "Usage: just plugin validate <path>" && exit 1; ./{{x86_binary_name}} plugin validate "{{arg}}" ;;
        info)     [ -z "{{arg}}" ] && echo "Usage: just plugin info <name>" && exit 1; ./{{x86_binary_name}} plugin info "{{arg}}" ;;
        reload)   [ -z "{{arg}}" ] && echo "Usage: just plugin reload <name>" && exit 1; ./{{x86_binary_name}} plugin reload "{{arg}}" ;;
        enable)   [ -z "{{arg}}" ] && echo "Usage: just plugin enable <name>" && exit 1; ./{{x86_binary_name}} plugin enable "{{arg}}" ;;
        disable)  [ -z "{{arg}}" ] && echo "Usage: just plugin disable <name>" && exit 1; ./{{x86_binary_name}} plugin disable "{{arg}}" ;;
        *)        echo "Unknown action: {{action}}"; echo "Actions: list, init, validate, info, reload, enable, disable"; exit 1 ;;
    esac

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

# [Docker] Manage a database stack: just dc <backend> <action>
# Backends: full, sqlite, mysql, postgres, prod
# Actions: up, down, reset, dev, fresh, logs, destroy (full only), minio-reset (postgres only)
dc backend action:
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{backend}}" in
        full)     FILE="{{compose_file}}" ;;
        sqlite)   FILE="{{compose_sqlite}}" ;;
        mysql)    FILE="{{compose_mysql}}" ;;
        postgres) FILE="{{compose_postgres}}" ;;
        prod)     FILE="{{prod_compose}}" ;;
        *)        echo "Unknown backend: {{backend}}"; echo "Backends: full, sqlite, mysql, postgres, prod"; exit 1 ;;
    esac
    case "{{action}}" in
        up)           DOCKER_BUILDKIT=1 docker compose -f "$FILE" up -d --build ;;
        down)         docker compose -f "$FILE" down ;;
        reset)        docker compose -f "$FILE" down -v ;;
        dev)          DOCKER_BUILDKIT=1 docker compose -f "$FILE" up -d --build modula ;;
        fresh)        docker compose -f "$FILE" down -v && DOCKER_BUILDKIT=1 docker compose -f "$FILE" up -d --build ;;
        logs)         docker compose -f "$FILE" logs -f modula ;;
        destroy)
            if [ "{{backend}}" != "full" ]; then echo "destroy is only supported for the full backend"; exit 1; fi
            docker compose -f "$FILE" down -v --rmi all ;;
        minio-reset)
            if [ "{{backend}}" != "postgres" ]; then echo "minio-reset is only supported for postgres backend"; exit 1; fi
            docker compose -f "$FILE" down minio && docker compose -f "$FILE" up -d minio ;;
        *)            echo "Unknown action: {{action}}"; echo "Actions: up, down, reset, dev, fresh, logs, destroy (full), minio-reset (postgres)"; exit 1 ;;
    esac

# [Docker] Start infrastructure only (postgres, mysql, minio)
docker-infra:
    docker compose -f {{compose_file}} up -d postgres mysql minio

# [Docker] Build standalone CMS image (for CI)
docker-build:
    DOCKER_BUILDKIT=1 docker build --rm --tag modula .

# [Docker] Release container with tag latest and version
docker-release:
    docker tag modula {{docker_registry}}modula:latest
    docker tag modula {{docker_registry}}modula:{{version}}
    docker push {{docker_registry}}modula:latest
    docker push {{docker_registry}}modula:{{version}}

# [Dealer] Manage dealer container: just dealer <action>
# Actions: up, down, reset, destroy, rebuild, logs
dealer action:
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{action}}" in
        up)      DOCKER_BUILDKIT=1 {{dealer_compose}} up -d --build ;;
        down)    {{dealer_compose}} down ;;
        reset)   {{dealer_compose}} down -v ;;
        destroy) {{dealer_compose}} down -v --rmi all ;;
        rebuild) DOCKER_BUILDKIT=1 {{dealer_compose}} up -d --build --force-recreate ;;
        logs)    {{dealer_compose}} logs -f modula ;;
        *)       echo "Unknown action: {{action}}"; echo "Actions: up, down, reset, destroy, rebuild, logs"; exit 1 ;;
    esac

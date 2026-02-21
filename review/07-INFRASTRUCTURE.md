# Infrastructure Review (Docker, CI/CD, Build)

## Docker: Strong

### Dockerfile
Multi-stage build (golang:1.24-bookworm builder, debian:bookworm-slim runtime). Non-root user (uid 1000). Stripped binary. BuildKit cache mounts for Go build cache. 5 persistent volumes (data, certs, ssh, backups, plugins). Runtime dependencies properly separated from build dependencies.

This is a well-constructed Dockerfile. The non-root user, stripped binary, and minimal runtime image demonstrate security awareness.

### Compose Variants (6 files)
- `docker-compose.yml`: Dev/dealer setup (CMS only)
- `docker-compose.full.yml`: All databases + MinIO
- `docker-compose.sqlite.yml`: Lightest footprint (CMS + MinIO)
- `docker-compose.mysql.yml`: CMS + MySQL + MinIO
- `docker-compose.postgres.yml`: CMS + PostgreSQL + MinIO
- `docker-compose.prod.yml`: Caddy reverse proxy + PostgreSQL + MinIO

All stateful services have healthchecks. Service dependencies are declared. Named volumes with clear prefixes.

**What's extra:** Six compose variants may be more than needed. The sqlite, mysql, and postgres variants could be merged into one file with profiles. The full stack compose (all databases) is mainly useful for testing.

**What's good about prod:** Caddy for automatic HTTPS, proper service dependencies (service_healthy), read-only config mount, env_file support. This is deployment-ready.

## CI/CD: Solid with Gaps

### Go Workflow
- Tests run on ubuntu-latest with Go 1.26
- 4-platform build matrix: macOS amd64/arm64 + Linux amd64/arm64
- Cross-compilation with proper CGO setup (aarch64-linux-gnu-gcc)
- Release job creates GitHub release with artifacts on tag push
- Dev deploy via SSH with health check (5 attempts, 5s delay)

### SDK Workflow
- TypeScript: pnpm 9, Node 22, frozen lockfile
- Go SDK: vet + test
- Swift SDK: Xcode 15.4 on macOS-14

### What's Missing from CI
- **No container image build/push** to a registry (Docker Hub, ghcr.io, ECR)
- **No release signing** or checksum files for binary artifacts
- **No SAST scanning** (no golangci-lint in CI, no trivy for images)
- **No staging deployment** - dev branch goes straight to a server
- **No integration tests** in CI (S3 tests require MinIO, always skipped)

## Justfile: Comprehensive

80+ recipes covering test, dev, build, SQL, SDK (TypeScript/Go/Swift), plugins, lint, Docker, deploy, and MCP. Smart defaults with `env_var_or_default`. Multi-database Docker stacks. Safe deploy with rollback capability.

This is a well-designed developer experience. `just dev` builds and runs. `just test` handles test database lifecycle. `just sdk-build` builds all three SDK packages. `just docker-up` starts the full stack.

## Dependencies: Well-Chosen

18 major direct Go dependencies. Charmbracelet ecosystem for TUI (Bubbletea, Wish, Lipgloss). gopher-lua for plugins. AWS SDK for S3. Cobra for CLI. Three database drivers (sqlite3, mysql, pq). All are established, maintained packages.

Two CGO dependencies: mattn/go-sqlite3 and kolesa-team/go-webp. CGO complicates cross-compilation but is necessary for SQLite performance and WebP encoding.

Zero external dependencies in all three SDKs (stdlib/Foundation only).

## Backup System

Multi-database backup/restore. ZIP packaging with manifest metadata. Local filesystem and S3 storage options. Proper CLI commands.

**What's missing:** No incremental/differential backups. No automatic scheduling (must be CLI-driven or cron). No compression beyond ZIP.

## Media System

Upload pipeline: validate size, detect MIME, check duplicate, write temp, upload to S3, create DB record, run optimization, upload variants. Image optimization with WebP conversion, focal point support, responsive image presets.

**What's good:** Rollback on failure, S3-compatible (works with MinIO, AWS, etc.), audit trail integration.
**What's missing:** No batch upload support. WebP via C binding adds CGO complexity.

## Documentation

70+ markdown files in `ai/` directory covering architecture, domain knowledge, detailed package docs, workflows, sqlc reference, refactoring plans, and learning notes. Regularly updated (February 2026 timestamps).

This is some of the best project documentation I've seen. It's clearly designed for AI-assisted development (the `ai/` prefix is a hint), but it serves human developers equally well.

## Recommendations

1. **Add container image CI** - build and push to ghcr.io on tag
2. **Add release checksums** - sha256sum for all binary artifacts
3. **Add SAST** - golangci-lint in CI, trivy for container images
4. **Consolidate compose files** - use Docker Compose profiles instead of separate files
5. **Add incremental backups** for large deployments
6. **Add cron-based backup scheduling** as a server feature

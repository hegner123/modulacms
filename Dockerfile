# =============================================================================
# ModulaCMS Production Dockerfile
# =============================================================================
# Multi-stage build: compile with CGO, run on minimal Debian slim
#
# Build:
#   docker build -t modulacms .
#   docker build --build-arg VERSION=1.0.0 -t modulacms:1.0.0 .
#
# Run:
#   docker run -d -p 8080:8080 -p 4000:4000 -p 2233:2233 \
#     -v modulacms-data:/app/data \
#     -v modulacms-certs:/app/certs \
#     -v modulacms-ssh:/app/.ssh \
#     -v modulacms-backups:/app/backups \
#     modulacms
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Builder
# -----------------------------------------------------------------------------
FROM golang:1.24-bookworm AS builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

WORKDIR /build

ENV TERM=xterm-256color

# Build dependencies: C libraries required by CGO packages (sqlite3, webp)
RUN apt-get update && apt-get install -y --no-install-recommends \
    libwebp-dev \
    && rm -rf /var/lib/apt/lists/*

# Layer 1: vendor (large, rarely changes — cached)
COPY vendor/ vendor/
COPY go.mod go.sum ./

# Layer 2: source (small, changes frequently)
COPY cmd/ cmd/
COPY internal/ internal/
COPY sql/ sql/

# Build with CGO enabled (required for sqlite3)
# BuildKit cache mount preserves Go build cache across Docker builds
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -mod vendor \
    -ldflags="-s -w \
        -X 'github.com/hegner123/modulacms/internal/utility.Version=${VERSION}' \
        -X 'github.com/hegner123/modulacms/internal/utility.GitCommit=${COMMIT}' \
        -X 'github.com/hegner123/modulacms/internal/utility.BuildDate=${BUILD_DATE}'" \
    -o modulacms ./cmd

# -----------------------------------------------------------------------------
# Stage 2: Runtime
# -----------------------------------------------------------------------------
FROM debian:bookworm-slim

ENV TERM=xterm-256color

LABEL org.opencontainers.image.title="ModulaCMS"
LABEL org.opencontainers.image.description="Headless CMS with HTTP, HTTPS, and SSH servers"
LABEL org.opencontainers.image.source="https://github.com/hegner123/modulacms"
LABEL org.opencontainers.image.licenses="AGPL-3.0"

# Runtime dependencies: TLS certs, timezone data, shared libraries for CGO,
# and database client tools for backup/restore (pg_dump, mysqldump)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    libwebp7 \
    postgresql-client \
    default-mysql-client \
    && rm -rf /var/lib/apt/lists/* \
    && update-ca-certificates

# Non-root user
RUN groupadd --gid 1000 modulacms \
    && useradd --uid 1000 --gid modulacms --shell /bin/false --create-home modulacms

WORKDIR /app

# Persistent data directories
RUN mkdir -p /app/data /app/certs /app/.ssh /app/backups \
    && chown -R modulacms:modulacms /app

# Copy binary from builder
COPY --from=builder --chown=modulacms:modulacms /build/modulacms /app/modulacms

USER modulacms

# HTTP, HTTPS, SSH
EXPOSE 8080 4000 2233

VOLUME ["/app/data", "/app/certs", "/app/.ssh", "/app/backups"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/app/modulacms", "--version"]

ENTRYPOINT ["/bin/sh", "-c", "/app/modulacms cert generate && /app/modulacms serve"]

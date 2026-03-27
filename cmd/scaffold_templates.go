package main

import "fmt"

// scaffoldData holds all derived flags for template rendering.
type scaffoldData struct {
	DBDriver string
	Storage  string
	Proxy    string

	IsSQLite   bool
	IsMySQL    bool
	IsPostgres bool
	IsMinio    bool
	IsExternal bool
	IsCaddy    bool
	IsNginx    bool
	IsRaw      bool
}

func newScaffoldData(a scaffoldArgs) scaffoldData {
	return scaffoldData{
		DBDriver:   a.DBDriver,
		Storage:    a.Storage,
		Proxy:      a.Proxy,
		IsSQLite:   a.DBDriver == "sqlite",
		IsMySQL:    a.DBDriver == "mysql",
		IsPostgres: a.DBDriver == "postgres",
		IsMinio:    a.Storage == "minio",
		IsExternal: a.Storage == "external",
		IsCaddy:    a.Proxy == "caddy",
		IsNginx:    a.Proxy == "nginx",
		IsRaw:      a.Proxy == "raw",
	}
}

// ---------------------------------------------------------------------------
// Dockerfile
// ---------------------------------------------------------------------------

func renderDockerfile(_ scaffoldData) string {
	return `# ModulaCMS — multi-stage production build
# Build:  docker compose build
# Source: https://github.com/hegner123/modulacms

# --- Builder ---
FROM golang:1.25-bookworm AS builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

WORKDIR /build
ENV TERM=xterm-256color

RUN apt-get update && apt-get install -y --no-install-recommends \
    libwebp-dev \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
COPY sdks/go/ sdks/go/
COPY vendor/ vendor/
COPY cmd/ cmd/
COPY internal/ internal/
COPY sql/ sql/

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -mod vendor \
    -ldflags="-s -w \
        -X 'github.com/hegner123/modulacms/internal/utility.Version=${VERSION}' \
        -X 'github.com/hegner123/modulacms/internal/utility.GitCommit=${COMMIT}' \
        -X 'github.com/hegner123/modulacms/internal/utility.BuildDate=${BUILD_DATE}'" \
    -o modula ./cmd

# --- Runtime ---
FROM debian:bookworm-slim

ENV TERM=xterm-256color

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    libwebp7 \
    postgresql-client \
    default-mysql-client \
    neovim \
    && rm -rf /var/lib/apt/lists/* \
    && update-ca-certificates

ENV EDITOR=nvim

RUN groupadd --gid 1000 modula \
    && useradd --uid 1000 --gid modula --shell /bin/false --create-home modula

WORKDIR /app

RUN mkdir -p /app/data /app/certs /app/.ssh /app/backups /app/plugins \
    && chown -R modula:modula /app

COPY --from=builder --chown=modula:modula /build/modula /app/modula

USER modula

EXPOSE 8080 4000 2233

VOLUME ["/app/data", "/app/certs", "/app/.ssh", "/app/backups", "/app/plugins"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/app/modula", "version"]

ENTRYPOINT ["/bin/sh", "-c", "/app/modula cert generate && /app/modula serve"]
`
}

// ---------------------------------------------------------------------------
// docker-compose.yml
// ---------------------------------------------------------------------------

func renderCompose(d scaffoldData) string {
	s := "name: modula\n\nservices:\n"

	// ---- Proxy ----
	if d.IsCaddy {
		s += `  caddy:
    image: caddy:2
    container_name: modula_caddy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
    depends_on:
      - modula
`
	}
	if d.IsNginx {
		s += `  nginx:
    image: nginx:alpine
    container_name: modula_nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro
      # Uncomment for TLS:
      # - ./certs:/etc/nginx/certs:ro
    depends_on:
      - modula
`
	}

	// ---- CMS ----
	s += "\n  modula:\n"
	s += `    build:
      context: .
      dockerfile: Dockerfile
      args:
        VERSION: ${VERSION:-dev}
        COMMIT: ${COMMIT:-unknown}
        BUILD_DATE: ${BUILD_DATE:-unknown}
    container_name: modula_cms
    restart: unless-stopped
    environment:
      - TERM=xterm-256color
    ports:
      - "2233:2233"
`
	if d.IsRaw {
		s += "      - \"8080:8080\"\n"
		s += "      - \"4000:4000\"\n"
	}

	s += `    volumes:
      - cms_data:/app/data
      - cms_ssh:/app/.ssh
      - cms_backups:/app/backups
      - cms_plugins:/app/plugins
      - ./modula.config.json:/app/modula.config.json:ro
    env_file:
      - .env
    entrypoint: ["/bin/sh", "-c", "/app/modula cert generate && /app/modula serve"]
`

	// depends_on
	deps := []string{}
	if d.IsPostgres {
		deps = append(deps, "postgres")
	}
	if d.IsMySQL {
		deps = append(deps, "mysql")
	}
	if d.IsMinio {
		deps = append(deps, "minio")
	}
	if len(deps) > 0 {
		s += "    depends_on:\n"
		for _, dep := range deps {
			s += fmt.Sprintf("      %s:\n        condition: service_healthy\n", dep)
		}
	}

	// ---- Database ----
	if d.IsPostgres {
		s += `
  postgres:
    image: postgres:17
    container_name: modula_postgres
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${POSTGRES_USER:-modula}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-modula}
      POSTGRES_DB: ${POSTGRES_DB:-modula_db}
    volumes:
      - pg_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-modula}"]
      interval: 10s
      timeout: 5s
      retries: 5
`
	}
	if d.IsMySQL {
		s += `
  mysql:
    image: mysql:8.0
    container_name: modula_mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD:-root_password}
      MYSQL_DATABASE: ${MYSQL_DATABASE:-modula_db}
      MYSQL_USER: ${MYSQL_USER:-modula}
      MYSQL_PASSWORD: ${MYSQL_PASSWORD:-modula}
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5
`
	}

	// ---- MinIO ----
	if d.IsMinio {
		s += `
  minio:
    image: minio/minio:latest
    container_name: modula_minio
    restart: unless-stopped
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_ROOT_USER:-modula}
      MINIO_ROOT_PASSWORD: ${MINIO_ROOT_PASSWORD:-CHANGE_ME_minio}
    ports:
      - "9001:9001"
    volumes:
      - minio_data:/data
    healthcheck:
      test: ["CMD", "mc", "ready", "local"]
      interval: 10s
      timeout: 5s
      retries: 5
`
	}

	// ---- Volumes ----
	s += "\nvolumes:\n"
	s += "  cms_data:\n  cms_ssh:\n  cms_backups:\n  cms_plugins:\n"

	if d.IsCaddy {
		s += "  caddy_data:\n  caddy_config:\n"
	}
	if d.IsPostgres {
		s += "  pg_data:\n"
	}
	if d.IsMySQL {
		s += "  mysql_data:\n"
	}
	if d.IsMinio {
		s += "  minio_data:\n"
	}

	return s
}

// ---------------------------------------------------------------------------
// modula.config.json
// ---------------------------------------------------------------------------

func renderConfig(d scaffoldData) string {
	dbURL := "/app/data/modula.db"
	dbName := ""
	dbUser := ""
	dbPass := ""

	switch {
	case d.IsPostgres:
		dbURL = "postgres"
		dbName = "modula_db"
		dbUser = "modula"
		dbPass = "modula"
	case d.IsMySQL:
		dbURL = "mysql:3306"
		dbName = "modula_db"
		dbUser = "modula"
		dbPass = "modula"
	}

	bucketEndpoint := ""
	bucketPathStyle := "false"
	bucketKey := ""
	bucketSecret := ""
	bucketName := "modula-media"
	adminBucketName := "modula-admin-media"

	if d.IsMinio {
		bucketEndpoint = "minio:9000"
		bucketPathStyle = "true"
		bucketKey = "modula"
		bucketSecret = "CHANGE_ME_minio"
	} else {
		bucketEndpoint = "CHANGE_ME_s3_endpoint"
		bucketKey = "CHANGE_ME_access_key"
		bucketSecret = "CHANGE_ME_secret_key"
	}

	return fmt.Sprintf(`{
  "environment": "local-docker",
  "db_driver": "%s",
  "db_url": "%s",
  "db_name": "%s",
  "db_username": "%s",
  "db_password": "%s",
  "port": "8080",
  "ssl_port": "4000",
  "ssh_port": "2233",
  "ssh_host": "0.0.0.0",
  "client_site": "0.0.0.0",
  "admin_site": "0.0.0.0",
  "cookie_secure": true,
  "bucket_name": "%s",
  "bucket_admin_name": "%s",
  "bucket_endpoint": "%s",
  "bucket_access_key": "%s",
  "bucket_secret_key": "%s",
  "bucket_region": "us-east-1",
  "bucket_use_ssl": false,
  "bucket_force_path_style": %s,
  "plugin_enabled": true,
  "plugin_directory": "/app/plugins/"
}
`, d.DBDriver, dbURL, dbName, dbUser, dbPass,
		bucketName, adminBucketName,
		bucketEndpoint, bucketKey, bucketSecret, bucketPathStyle)
}

// ---------------------------------------------------------------------------
// .env
// ---------------------------------------------------------------------------

func renderEnv(d scaffoldData) string {
	s := `# ModulaCMS deployment environment
VERSION=dev
COMMIT=unknown
BUILD_DATE=unknown
`

	if d.IsPostgres {
		s += `
# PostgreSQL
POSTGRES_USER=modula
POSTGRES_PASSWORD=CHANGE_ME_postgres
POSTGRES_DB=modula_db
`
	}

	if d.IsMySQL {
		s += `
# MySQL
MYSQL_ROOT_PASSWORD=CHANGE_ME_root
MYSQL_DATABASE=modula_db
MYSQL_USER=modula
MYSQL_PASSWORD=CHANGE_ME_mysql
`
	}

	if d.IsMinio {
		s += `
# MinIO
MINIO_ROOT_USER=modula
MINIO_ROOT_PASSWORD=CHANGE_ME_minio
`
	}

	if d.IsExternal {
		s += `
# External S3
BUCKET_ENDPOINT=CHANGE_ME_s3_endpoint
BUCKET_ACCESS_KEY=CHANGE_ME_access_key
BUCKET_SECRET_KEY=CHANGE_ME_secret_key
`
	}

	return s
}

// ---------------------------------------------------------------------------
// Caddyfile
// ---------------------------------------------------------------------------

func renderCaddyfile(_ scaffoldData) string {
	return `# ModulaCMS reverse proxy
# Replace :80 with your domain for automatic HTTPS:
#   your-domain.com {
#       reverse_proxy modula:8080
#   }

:80 {
    reverse_proxy modula:8080
}
`
}

// ---------------------------------------------------------------------------
// nginx.conf
// ---------------------------------------------------------------------------

func renderNginxConf(_ scaffoldData) string {
	return `# ModulaCMS reverse proxy
# For TLS, add ssl directives and certificate paths.

upstream modula_backend {
    server modula:8080;
}

server {
    listen 80;
    server_name _;

    client_max_body_size 100M;

    location / {
        proxy_pass http://modula_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
`
}

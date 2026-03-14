# Docker Deployment

ModulaCMS provides Docker Compose configurations for running the CMS with different database backends. Each configuration includes the CMS binary, the chosen database, and MinIO for S3-compatible media storage.

All Docker commands use `just` as the task runner. The underlying compose files live in `deploy/docker/`.

## Available Stacks

### Full Stack

Starts the CMS with all three database engines (SQLite, MySQL, PostgreSQL) and MinIO:

```bash
just dc full up
```

### Single Database Stacks

Run the CMS with a specific database backend:

```bash
# SQLite
just dc sqlite up

# MySQL
just dc mysql up

# PostgreSQL
just dc postgres up
```

### Infrastructure Only

Start only the supporting services (PostgreSQL, MySQL, MinIO) without the CMS binary. Useful when you want to run the CMS locally against containerized databases:

```bash
just docker-infra
```

## Managing Stacks

The `just dc` command accepts a backend and an action:

```bash
just dc <backend> <action>
```

| Backend | Compose File |
|---------|-------------|
| `full` | `deploy/docker/docker-compose.full.yml` |
| `sqlite` | `deploy/docker/docker-compose.sqlite.yml` |
| `mysql` | `deploy/docker/docker-compose.mysql.yml` |
| `postgres` | `deploy/docker/docker-compose.postgres.yml` |

| Action | Description |
|--------|-------------|
| `up` | Build images and start all containers |
| `down` | Stop containers, keep volumes |
| `reset` | Stop containers and delete volumes (destroys data) |
| `dev` | Rebuild and restart only the CMS container |
| `fresh` | Delete volumes, rebuild, and start everything from scratch |
| `logs` | Follow CMS container logs |
| `destroy` | Stop containers, delete volumes, and remove all images (`full` backend only) |
| `minio` | Restart just the MinIO container |

### Examples

Rebuild the CMS container after code changes without touching the database:

```bash
just dc sqlite dev
```

Wipe all data and start fresh:

```bash
just dc mysql fresh
```

View CMS logs:

```bash
just dc postgres logs
```

Destroy everything including images (full backend only):

```bash
just dc full destroy
```

## Volume Management

Docker volumes persist data between container restarts. Understanding when data is preserved and when it is destroyed matters:

| Command | Containers | Volumes | Images |
|---------|-----------|---------|--------|
| `just dc <backend> down` | Stopped | Preserved | Preserved |
| `just dc <backend> reset` | Stopped | Deleted | Preserved |
| `just dc <backend> fresh` | Restarted | Deleted | Rebuilt |
| `just dc full destroy` | Stopped | Deleted | Deleted |

Use `down` during normal development. Use `reset` or `fresh` when you want to start with an empty database. Use `destroy` (full backend only) when you need to clear cached Docker layers.

## Building the CMS Image

Build a standalone CMS image without starting any services:

```bash
just docker-build
```

This produces a `modula` image tagged as `latest`, suitable for CI pipelines or custom deployments.

## Releasing Container Images

Tag and push the CMS image to a registry:

```bash
just docker-release
```

This tags the image with both `latest` and the current version from `justfile`, then pushes both tags to the configured registry.

## MinIO for Local S3

All Docker stacks include MinIO for S3-compatible media storage. When running in Docker, the CMS connects to MinIO using the container hostname (`minio:9000`), but browsers need the externally reachable address.

Set `bucket_public_url` in your `modula.config.json` to the address accessible from outside Docker:

```json
{
  "bucket_endpoint": "minio:9000",
  "bucket_public_url": "http://localhost:9000"
}
```

To start MinIO independently for running integration tests:

```bash
just test-minio       # Start MinIO
just test-integration # Run S3 integration tests
just test-minio-down  # Stop MinIO
```

## Notes

- All Docker stacks require `DOCKER_BUILDKIT=1`, which `just` enables automatically.
- The CMS binary is built inside the Docker image using a multi-stage build. CGO is enabled for SQLite support.
- Configuration inside Docker containers uses the same `modula.config.json` format as bare-metal deployments.
- The `docker-infra` command is useful for running cross-backend database integration tests locally.

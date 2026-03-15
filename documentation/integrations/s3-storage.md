# S3 Storage

Connect ModulaCMS to S3-compatible object storage for media assets and backups.

## Supported Providers

Any S3-compatible storage provider works with ModulaCMS:

- AWS S3
- MinIO (self-hosted)
- DigitalOcean Spaces
- Backblaze B2
- Cloudflare R2
- Linode Object Storage

## Configuration

Set these fields in `modula.config.json`:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `bucket_region` | string | `"us-east-1"` | S3 region |
| `bucket_media` | string | `""` | Bucket name for media assets |
| `bucket_backup` | string | `""` | Bucket name for backups |
| `bucket_endpoint` | string | `""` | S3 API endpoint hostname (without scheme) |
| `bucket_access_key` | string | `""` | S3 access key ID |
| `bucket_secret_key` | string | `""` | S3 secret access key |
| `bucket_public_url` | string | (falls back to endpoint) | Public-facing base URL for media links |
| `bucket_default_acl` | string | `""` | ACL applied to uploaded objects |
| `bucket_force_path_style` | bool | `true` | Use path-style URLs instead of virtual-hosted |
| `max_upload_size` | integer | `10485760` | Maximum upload size in bytes (10 MB) |

All S3 storage fields are hot-reloadable. Changes take effect without restarting the server.

> **Good to know**: Do not include the URL scheme (`http://` or `https://`) in `bucket_endpoint`. The scheme is determined by the `environment` config field. `http-only` and `docker` environments use HTTP; all others use HTTPS.

## Set Up AWS S3

1. Create an S3 bucket in the AWS Console.
2. Create an IAM user with `s3:PutObject`, `s3:GetObject`, `s3:DeleteObject`, and `s3:HeadBucket` permissions on the bucket.
3. Generate access keys for the IAM user.

```json
{
  "bucket_region": "us-east-1",
  "bucket_media": "my-cms-media",
  "bucket_backup": "my-cms-backups",
  "bucket_endpoint": "s3.us-east-1.amazonaws.com",
  "bucket_access_key": "AKIAIOSFODNN7EXAMPLE",
  "bucket_secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
  "bucket_public_url": "https://my-cms-media.s3.us-east-1.amazonaws.com",
  "bucket_force_path_style": false
}
```

> **Good to know**: AWS S3 uses virtual-hosted style URLs by default. Set `bucket_force_path_style` to `false` for AWS. Most other providers require it set to `true`.

## Set Up MinIO

MinIO is a self-hosted S3-compatible server, commonly used for local development and Docker-based deployments.

```json
{
  "bucket_region": "us-east-1",
  "bucket_media": "media",
  "bucket_backup": "backups",
  "bucket_endpoint": "localhost:9000",
  "bucket_access_key": "minioadmin",
  "bucket_secret_key": "minioadmin",
  "bucket_public_url": "http://localhost:9000",
  "bucket_force_path_style": true
}
```

### MinIO in Docker

When running both ModulaCMS and MinIO in Docker containers, the two services communicate over the Docker network using container hostnames. Browsers cannot resolve container hostnames, so you need different values for `bucket_endpoint` (internal) and `bucket_public_url` (external).

```json
{
  "bucket_region": "us-east-1",
  "bucket_media": "media",
  "bucket_backup": "backups",
  "bucket_endpoint": "minio:9000",
  "bucket_access_key": "minioadmin",
  "bucket_secret_key": "minioadmin",
  "bucket_public_url": "http://localhost:9000",
  "bucket_force_path_style": true
}
```

- `bucket_endpoint` points to the MinIO container hostname (`minio:9000`) for API calls.
- `bucket_public_url` points to the externally reachable address (`http://localhost:9000`) so media URLs in API responses work in the browser.

## Set Up DigitalOcean Spaces

1. Create a Space in the DigitalOcean Console.
2. Generate a Spaces access key under **API > Spaces Keys**.

```json
{
  "bucket_region": "nyc3",
  "bucket_media": "my-cms-media",
  "bucket_endpoint": "nyc3.digitaloceanspaces.com",
  "bucket_access_key": "DO00EXAMPLE...",
  "bucket_secret_key": "your-spaces-secret-key",
  "bucket_public_url": "https://my-cms-media.nyc3.digitaloceanspaces.com",
  "bucket_force_path_style": false
}
```

## Set Up Backblaze B2

1. Create a B2 bucket in the Backblaze Console.
2. Create an application key with read/write access to the bucket.

```json
{
  "bucket_region": "us-west-004",
  "bucket_media": "my-cms-media",
  "bucket_endpoint": "s3.us-west-004.backblazeb2.com",
  "bucket_access_key": "your-b2-key-id",
  "bucket_secret_key": "your-b2-application-key",
  "bucket_public_url": "https://my-cms-media.s3.us-west-004.backblazeb2.com",
  "bucket_force_path_style": false
}
```

## Set Up Cloudflare R2

1. Create an R2 bucket in the Cloudflare Dashboard.
2. Create an API token under **R2 > Manage R2 API Tokens** with Object Read & Write permissions.

```json
{
  "bucket_region": "auto",
  "bucket_media": "my-cms-media",
  "bucket_endpoint": "your-account-id.r2.cloudflarestorage.com",
  "bucket_access_key": "your-r2-access-key",
  "bucket_secret_key": "your-r2-secret-key",
  "bucket_public_url": "https://media.example.com",
  "bucket_force_path_style": true
}
```

> **Good to know**: Cloudflare R2 does not charge egress fees. Set `bucket_public_url` to your R2 custom domain or the public bucket URL for serving media.

## Understand bucket_public_url vs bucket_endpoint

These two fields serve different purposes:

- **`bucket_endpoint`** is the S3 API hostname that ModulaCMS uses to upload, download, and delete objects. This is where API calls go.
- **`bucket_public_url`** is the base URL that appears in media URLs returned by the API. This is what browsers and frontends use to load images and files.

In most production setups, these are the same. They diverge in Docker environments where the S3 service has an internal hostname that browsers cannot reach.

## Verify the Connection

After configuring S3, check the health endpoint to verify connectivity:

```bash
curl http://localhost:8080/api/v1/health
```

The `storage` check performs a HeadBucket call against your configured bucket. A `true` value confirms ModulaCMS can reach your storage provider.

```json
{
  "status": "ok",
  "checks": {
    "database": true,
    "storage": true,
    "plugins": true
  }
}
```

## Media Uploads

With S3 configured, upload files through the media API:

```bash
curl -X POST http://localhost:8080/api/v1/media \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -F "file=@/path/to/photo.jpg"
```

ModulaCMS uploads the original file and generates optimized image variants (WebP at configured dimension presets) automatically.

## Next Steps

- [Media management guide](../building-content/media.md) -- uploading, optimization, dimension presets
- [Observability](observability.md) -- storage health checks and monitoring
- [Configuration reference](../getting-started/configuration.md) -- all config fields

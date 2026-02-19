# Media Management

ModulaCMS stores media assets (images, documents, and other files) in S3-compatible object storage. When you upload an image, the system automatically generates optimized variants at each configured dimension preset, uploads them alongside the original, and records the full set of URLs. Non-image files are stored as-is. All media operations go through the REST API, which handles validation, storage, database records, and cleanup as a single transaction with automatic rollback on failure.

## Concepts

**Media asset** -- A file stored in S3 with a database record tracking its metadata (name, display name, alt text, caption, MIME type, URL, dimensions, and author). Each asset has a unique 26-character ULID as its identifier.

**Dimension presets** -- Named width/height pairs (e.g., "thumbnail" at 150x150, "hero" at 1920x1080) that define the sizes the optimization pipeline produces. When you upload an image, the system generates a cropped and scaled variant for each preset where the image is large enough. Presets that exceed the source image dimensions are skipped to avoid upscaling.

**Srcset** -- A JSON array of URLs for the optimized variants. After a successful upload, the media record's `srcset` field contains the public URLs for all generated sizes. Use these to serve responsive images.

**Focal point** -- Normalized coordinates (`focal_x`, `focal_y`) ranging from 0.0 to 1.0 that define the center of interest in an image. When set, the optimization pipeline crops around this point instead of the image center. Set the focal point on the media record before or after upload; re-uploading is not required for the focal point to take effect on future optimization runs.

**S3-compatible storage** -- ModulaCMS uses the AWS S3 API for object storage. Any S3-compatible provider works: AWS S3, MinIO, DigitalOcean Spaces, Backblaze B2, Cloudflare R2, and others. You configure the endpoint, credentials, and bucket name in `config.json`.

## Configuration

Set these fields in your `config.json` to connect media storage:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `bucket_region` | string | `"us-east-1"` | S3 region for the storage provider |
| `bucket_media` | string | -- | Bucket name for media assets |
| `bucket_endpoint` | string | -- | S3 API endpoint hostname (without scheme) |
| `bucket_access_key` | string | -- | S3 access key ID |
| `bucket_secret_key` | string | -- | S3 secret access key |
| `bucket_public_url` | string | (falls back to `bucket_endpoint`) | Public-facing base URL for media links |
| `bucket_default_acl` | string | `"public-read"` | ACL applied to uploaded objects |
| `bucket_force_path_style` | bool | `true` | Use path-style URLs (`endpoint/bucket/key`) instead of virtual-hosted |
| `max_upload_size` | integer | `10485760` (10 MB) | Maximum upload size in bytes |

Example configuration for MinIO running locally:

```json
{
  "bucket_region": "us-east-1",
  "bucket_media": "media",
  "bucket_endpoint": "localhost:9000",
  "bucket_access_key": "minioadmin",
  "bucket_secret_key": "minioadmin",
  "bucket_public_url": "http://localhost:9000",
  "bucket_force_path_style": true,
  "max_upload_size": 10485760
}
```

When running in Docker, `bucket_endpoint` typically points to the container hostname (e.g., `minio:9000`), which browsers cannot resolve. Set `bucket_public_url` to the externally reachable address (e.g., `http://localhost:9000`) so that media URLs in API responses work in the browser.

The scheme (`http://` or `https://`) for the S3 API endpoint is determined by the `environment` config field. Environments set to `"http-only"` or `"docker"` use `http`; all others use `https`.

## Uploading Media

Upload a file with a multipart form POST to `/api/v1/media`. The form field name must be `file`.

```bash
curl -X POST http://localhost:8080/api/v1/media \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -F "file=@/path/to/photo.jpg"
```

The upload pipeline:

1. Validates the file size against `max_upload_size`.
2. Detects the MIME type from the file contents (not the extension).
3. Rejects the upload if a file with the same name already exists (HTTP 409).
4. Uploads the original file to S3.
5. Creates a database record with the S3 URL and detected MIME type.
6. If the file is a supported image type, generates optimized variants for each dimension preset and uploads them to S3. The media record's `srcset` is updated with the variant URLs.

On success, the response is the created media record (HTTP 201):

```json
{
  "media_id": "01JMKX5V6QNPZ3R8W4T2YH9B0D",
  "name": "photo.jpg",
  "display_name": null,
  "alt": null,
  "caption": null,
  "description": null,
  "class": null,
  "mimetype": "image/jpeg",
  "dimensions": null,
  "url": "http://localhost:9000/media/2026/2/photo.jpg",
  "srcset": "[\"http://localhost:9000/media/2026/2/photo-150x150.jpg\",\"http://localhost:9000/media/2026/2/photo-1920x1080.jpg\"]",
  "focal_x": null,
  "focal_y": null,
  "author_id": "01JMKW8N3QRYZ7T1B5K6F2P4HD",
  "date_created": "2026-02-18T14:30:00Z",
  "date_modified": "2026-02-18T14:30:00Z"
}
```

### Custom Storage Path

By default, uploaded files are organized by date: `YYYY/M/filename`. You can override this with the `path` form field:

```bash
curl -X POST http://localhost:8080/api/v1/media \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -F "file=@/path/to/logo.png" \
  -F "path=branding/logos"
```

This stores the file at `branding/logos/logo.png` in the bucket. The path must contain only letters, numbers, forward slashes, hyphens, underscores, and periods. Path traversal (`..`) is rejected.

### Supported Image Types

The optimization pipeline processes these MIME types:

- `image/png`
- `image/jpeg`
- `image/gif`
- `image/webp`

Files with other MIME types (PDFs, videos, documents) are stored as-is without optimization. WebP files can be decoded for optimization but may encounter encoding limitations during variant generation.

### Error Responses

| Status | Condition |
|--------|-----------|
| 400 | File too large, invalid multipart form, or invalid path |
| 409 | A media record with the same filename already exists |
| 500 | S3 unavailable, database error, or optimization pipeline failure |

If the pipeline fails after the original is uploaded to S3, the system rolls back by deleting the S3 object and removing the database record.

## Managing Media

### Listing Media

List all media items:

```bash
curl http://localhost:8080/api/v1/media \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

With pagination (add `limit` and/or `offset` query parameters):

```bash
curl "http://localhost:8080/api/v1/media?limit=20&offset=40" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Paginated responses include total count:

```json
{
  "data": [
    { "media_id": "...", "name": "photo.jpg", "url": "..." }
  ],
  "total": 142,
  "limit": 20,
  "offset": 40
}
```

Pagination defaults: `limit=50`, `offset=0`. Maximum `limit` is 1000.

### Getting a Single Media Item

```bash
curl "http://localhost:8080/api/v1/media/?q=01JMKX5V6QNPZ3R8W4T2YH9B0D" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

The `q` parameter is the media ID (26-character ULID).

### Updating Media Metadata

Update metadata fields on an existing media record with a PUT request. Include the `media_id` in the request body along with the fields you want to change:

```bash
curl -X PUT http://localhost:8080/api/v1/media/ \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "media_id": "01JMKX5V6QNPZ3R8W4T2YH9B0D",
    "name": "photo.jpg",
    "display_name": "Company Headquarters",
    "alt": "Aerial view of the company headquarters building",
    "caption": "Our main office in Portland, OR",
    "description": "High-resolution aerial photograph taken in 2026",
    "focal_x": 0.45,
    "focal_y": 0.3
  }'
```

Updatable fields:

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Internal filename |
| `display_name` | string | Human-readable display name |
| `alt` | string | Alt text for accessibility |
| `caption` | string | Caption for display |
| `description` | string | Longer description |
| `class` | string | CSS class or category tag |
| `focal_x` | float (0.0-1.0) | Horizontal focal point |
| `focal_y` | float (0.0-1.0) | Vertical focal point |

### Deleting Media

```bash
curl -X DELETE "http://localhost:8080/api/v1/media/?q=01JMKX5V6QNPZ3R8W4T2YH9B0D" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Deletion removes all S3 objects associated with the media record (the original file and all srcset variants) before deleting the database record. If any S3 deletions fail, they are logged but the database record is still removed.

## Dimension Presets

Dimension presets control the sizes generated during image optimization. Each preset defines a label, width, height, and optional aspect ratio. Presets apply to all future uploads -- existing media is not retroactively reprocessed when presets change.

### Listing Presets

```bash
curl http://localhost:8080/api/v1/mediadimensions \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Response:

```json
[
  {
    "md_id": "01JMKY2R8QWTZ5P3N7V1A4H6BK",
    "label": "thumbnail",
    "width": 150,
    "height": 150,
    "aspect_ratio": "1:1"
  },
  {
    "md_id": "01JMKY3T9RXUA6Q4M8W2B5J7CL",
    "label": "hero",
    "width": 1920,
    "height": 1080,
    "aspect_ratio": "16:9"
  }
]
```

### Creating a Preset

```bash
curl -X POST http://localhost:8080/api/v1/mediadimensions \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "label": "card",
    "width": 400,
    "height": 300,
    "aspect_ratio": "4:3"
  }'
```

Labels must be unique. Width and height are in pixels.

### Updating a Preset

```bash
curl -X PUT http://localhost:8080/api/v1/mediadimensions/ \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "md_id": "01JMKY2R8QWTZ5P3N7V1A4H6BK",
    "label": "thumbnail",
    "width": 200,
    "height": 200,
    "aspect_ratio": "1:1"
  }'
```

### Getting a Single Preset

```bash
curl "http://localhost:8080/api/v1/mediadimensions/?q=01JMKY2R8QWTZ5P3N7V1A4H6BK" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

### Deleting a Preset

```bash
curl -X DELETE "http://localhost:8080/api/v1/mediadimensions/?q=01JMKY2R8QWTZ5P3N7V1A4H6BK" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

## Storage Health

Two admin-only endpoints help you find and clean up orphaned files in the media bucket -- S3 objects that have no corresponding database record.

### Checking for Orphans

```bash
curl http://localhost:8080/api/v1/media/health \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Response:

```json
{
  "total_objects": 347,
  "tracked_keys": 340,
  "orphaned_keys": [
    "2025/12/deleted-image.jpg",
    "2025/12/deleted-image-150x150.jpg"
  ],
  "orphan_count": 7
}
```

This compares every object in the S3 media bucket against the URLs and srcset entries in the database. Objects not referenced by any media record are reported as orphans.

### Cleaning Up Orphans

```bash
curl -X DELETE http://localhost:8080/api/v1/media/cleanup \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Response:

```json
{
  "deleted": ["2025/12/deleted-image.jpg", "2025/12/deleted-image-150x150.jpg"],
  "deleted_count": 2,
  "failed": [],
  "failed_count": 0
}
```

Run the health check first to review orphaned keys before cleanup. Both endpoints require the `media:admin` permission.

## API Reference

All endpoints are prefixed with `/api/v1`.

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/media` | `media:read` | List all media (supports `limit` and `offset` for pagination) |
| POST | `/api/v1/media` | `media:create` | Upload a new media file (multipart form, field name: `file`) |
| GET | `/api/v1/media/` | `media:read` | Get a single media item (`?q=MEDIA_ID`) |
| PUT | `/api/v1/media/` | `media:update` | Update media metadata (JSON body with `media_id`) |
| DELETE | `/api/v1/media/` | `media:delete` | Delete a media item and its S3 objects (`?q=MEDIA_ID`) |
| GET | `/api/v1/media/health` | `media:admin` | Check for orphaned S3 objects |
| DELETE | `/api/v1/media/cleanup` | `media:admin` | Delete orphaned S3 objects |
| GET | `/api/v1/mediadimensions` | `media:read` | List all dimension presets |
| POST | `/api/v1/mediadimensions` | `media:create` | Create a dimension preset |
| GET | `/api/v1/mediadimensions/` | `media:read` | Get a single dimension preset (`?q=MD_ID`) |
| PUT | `/api/v1/mediadimensions/` | `media:update` | Update a dimension preset |
| DELETE | `/api/v1/mediadimensions/` | `media:delete` | Delete a dimension preset (`?q=MD_ID`) |

Permission mapping for `/api/v1/media` and `/api/v1/mediadimensions` uses `RequireResourcePermission("media")`, which auto-maps HTTP methods: GET to `media:read`, POST to `media:create`, PUT to `media:update`, DELETE to `media:delete`. The health and cleanup endpoints use explicit `RequirePermission("media:admin")`.

## Notes

- **Image dimension limits.** The optimization pipeline rejects images exceeding 10,000 pixels in either width or height, or 50 megapixels total. This prevents memory exhaustion from decompression bombs.
- **No upscaling.** Dimension presets larger than the source image are silently skipped. An 800x600 image will not produce a 1920x1080 variant.
- **Duplicate filenames.** Uploading a file with the same name as an existing media record returns HTTP 409. Rename the file or delete the existing record first.
- **WebP support.** WebP files are decoded during optimization, but the encoder may produce errors in some configurations. The system logs a warning on WebP upload.
- **Rollback on failure.** If any step after the S3 upload fails (database insert, optimization pipeline), the original S3 object is deleted and the database record is removed. For optimization failures specifically, all partially uploaded variant files are also cleaned up.
- **Srcset format.** The `srcset` field is a JSON-encoded array of URL strings, not the HTML `srcset` attribute format. Parse it as JSON when consuming in your frontend.
- **Optimized variant filenames.** Variants follow the pattern `{basename}-{width}x{height}.{ext}`. For example, `photo.jpg` at 150x150 becomes `photo-150x150.jpg`.

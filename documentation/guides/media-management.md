# Media Management

ModulaCMS stores media assets in S3-compatible object storage. When you upload an image, the system automatically generates optimized variants at each configured dimension preset, uploads them alongside the original, and records the full set of URLs in the database. Non-image files are stored as-is. All uploads are transactional: if any step fails, the system rolls back S3 objects and database records automatically.

## Concepts

**Media asset** -- A file stored in S3 with a database record tracking its metadata: name, display name, alt text, caption, MIME type, URL, responsive image URLs, focal point, and author. Each asset has a unique 26-character ULID as its identifier.

**Dimension presets** -- Named width/height pairs (e.g., "thumbnail" at 150x150, "hero" at 1920x1080) that define what sizes the optimization pipeline produces. Presets that exceed the source image dimensions are skipped to avoid upscaling.

**Srcset** -- A JSON array of URLs for the optimized variants stored on each media record. Use these to serve responsive images in your frontend.

**Focal point** -- Normalized coordinates (`focal_x`, `focal_y`) ranging from 0.0 to 1.0 that define the center of interest in an image. When set, the optimization pipeline crops around this point instead of the image center. You can set the focal point before or after upload; re-uploading is not required.

## Media Folders

Media assets can be organized into a hierarchical folder structure. Folders support arbitrary nesting through parent-child relationships, letting you build a directory tree that mirrors your project's organizational needs (e.g., `branding/logos/`, `blog/2026/march/`).

### Folder Hierarchy

Each folder has an optional `parent_id` that references another folder. A folder with no parent sits at the root level. You can nest folders to any depth by setting the `parent_id` to an existing folder's ID.

### API Endpoints

All endpoints are prefixed with `/api/v1` and require `media:read`, `media:create`, `media:update`, or `media:delete` permissions as appropriate.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/media-folders` | List all media folders |
| POST | `/media-folders` | Create a new folder |
| GET | `/media-folders/` | Get a single folder (`?q=FOLDER_ID`) |
| PUT | `/media-folders/` | Update a folder |
| DELETE | `/media-folders/` | Delete a folder (`?q=FOLDER_ID`) |
| GET | `/media-folders/tree` | Get the full folder tree (nested structure) |
| GET | `/media-folders/{id}/media` | List media assets within a folder |

### Moving Media Between Folders

To move a media asset into a folder, update the media record's `folder_id` field:

```bash
curl -X PUT http://localhost:8080/api/v1/media/ \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "media_id": "01JMKX5V6QNPZ3R8W4T2YH9B0D",
    "folder_id": "01JNRWHSA1LQWZ3X5D8F2G9JKT"
  }'
```

Set `folder_id` to `null` or omit it to move a media asset back to the root level.

## Configuration

Set these fields in `modula.config.json` to connect to your S3-compatible storage provider:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `bucket_region` | string | `"us-east-1"` | S3 region |
| `bucket_media` | string | -- | Bucket name for media assets |
| `bucket_endpoint` | string | -- | S3 API endpoint hostname (without scheme) |
| `bucket_access_key` | string | -- | S3 access key ID |
| `bucket_secret_key` | string | -- | S3 secret access key |
| `bucket_public_url` | string | falls back to `bucket_endpoint` | Public-facing base URL for media links |
| `bucket_default_acl` | string | `"public-read"` | ACL applied to uploaded objects |
| `bucket_force_path_style` | bool | `true` | Use path-style URLs instead of virtual-hosted |
| `max_upload_size` | integer | `10485760` (10 MB) | Maximum upload size in bytes |

Example for MinIO running locally:

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

Example for DigitalOcean Spaces:

```json
{
  "bucket_region": "nyc3",
  "bucket_media": "media",
  "bucket_endpoint": "nyc3.digitaloceanspaces.com",
  "bucket_access_key": "DO00...",
  "bucket_secret_key": "...",
  "bucket_public_url": "https://nyc3.digitaloceanspaces.com"
}
```

The URL scheme (`http://` or `https://`) for the S3 API endpoint is determined by the `environment` config field. Environments set to `"http-only"` or `"docker"` use HTTP; all others use HTTPS.

When running in Docker, `bucket_endpoint` typically points to a container hostname (e.g., `minio:9000`) that browsers cannot resolve. Set `bucket_public_url` to the externally reachable address so that media URLs in API responses work in the browser.

**Compatible providers:** AWS S3, DigitalOcean Spaces, MinIO, Backblaze B2, Cloudflare R2, Linode Object Storage, and any other S3-compatible API.

## Uploading Media

Upload a file with a multipart form POST. The form field name must be `file`.

```bash
curl -X POST http://localhost:8080/api/v1/media \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -F "file=@/path/to/photo.jpg"
```

The upload pipeline:

1. Validates the file size against `max_upload_size`.
2. Detects the MIME type from file contents.
3. Rejects the upload if a file with the same name already exists (HTTP 409).
4. Uploads the original file to S3.
5. Creates a database record with the S3 URL and detected MIME type.
6. For supported image types, generates optimized variants for each dimension preset and uploads them to S3. Updates the media record's `srcset` with the variant URLs.

Response on success (HTTP 201):

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
  "srcset": "[\"http://localhost:9000/media/2026/2/photo-150x150.webp\",\"http://localhost:9000/media/2026/2/photo-1920x1080.webp\"]",
  "focal_x": null,
  "focal_y": null,
  "author_id": "01JMKW8N3QRYZ7T1B5K6F2P4HD",
  "date_created": "2026-02-18T14:30:00Z",
  "date_modified": "2026-02-18T14:30:00Z"
}
```

### Custom Storage Path

By default, files are organized by date: `YYYY/M/filename`. Override with the `path` form field:

```bash
curl -X POST http://localhost:8080/api/v1/media \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -F "file=@/path/to/logo.png" \
  -F "path=branding/logos"
```

This stores the file at `branding/logos/logo.png` in the bucket. The path must contain only letters, numbers, forward slashes, hyphens, underscores, and periods. Path traversal (`..`) is rejected.

### Image Optimization

The pipeline processes these image types:

- `image/png`
- `image/jpeg`
- `image/gif`
- `image/webp`

Non-image files (PDFs, videos, documents) are stored as-is without optimization.

For each configured dimension preset, the pipeline:

1. Center-crops the source image to the target aspect ratio (or crops around the focal point if set).
2. Scales the cropped region to the target size using bilinear interpolation.
3. Encodes the variant as WebP at quality 80.
4. Uploads the variant to S3 alongside the original.

Optimized variant filenames follow the pattern `{basename}-{width}x{height}.webp`. For example, `photo.jpg` with a 150x150 preset produces `photo-150x150.webp`.

### Error Responses

| Status | Condition |
|--------|-----------|
| 400 | File too large, invalid multipart form, or invalid path |
| 409 | A media record with the same filename already exists |
| 500 | S3 unavailable, database error, or optimization pipeline failure |

If the pipeline fails after the original is uploaded, the system rolls back by deleting the S3 object and removing the database record.

## Managing Media Records

### Listing Media

```bash
curl http://localhost:8080/api/v1/media \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

With pagination:

```bash
curl "http://localhost:8080/api/v1/media?limit=20&offset=40" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Paginated response:

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

### Getting a Single Media Item

```bash
curl "http://localhost:8080/api/v1/media/?q=01JMKX5V6QNPZ3R8W4T2YH9B0D" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

### Updating Media Metadata

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
    "focal_x": 0.45,
    "focal_y": 0.3
  }'
```

Updatable fields: `name`, `display_name`, `alt`, `caption`, `description`, `class`, `focal_x`, `focal_y`.

### Deleting Media

```bash
curl -X DELETE "http://localhost:8080/api/v1/media/?q=01JMKX5V6QNPZ3R8W4T2YH9B0D" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Deletion removes all S3 objects (original and all srcset variants) before removing the database record.

## Dimension Presets

Presets control the sizes generated during image optimization. They apply to all future uploads -- existing media is not retroactively reprocessed when presets change.

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

### Listing Presets

```bash
curl http://localhost:8080/api/v1/mediadimensions \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

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

### Updating and Deleting Presets

```bash
# Update
curl -X PUT http://localhost:8080/api/v1/mediadimensions/ \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"md_id": "01JMKY2R8QWTZ5P3N7V1A4H6BK", "label": "thumbnail", "width": 200, "height": 200, "aspect_ratio": "1:1"}'

# Delete
curl -X DELETE "http://localhost:8080/api/v1/mediadimensions/?q=01JMKY2R8QWTZ5P3N7V1A4H6BK" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

## Storage Health

Two admin-only endpoints help find and clean up orphaned files -- S3 objects with no corresponding database record.

```bash
# Check for orphans
curl http://localhost:8080/api/v1/media/health \
  -H "Cookie: session=YOUR_SESSION_COOKIE"

# Clean up orphans (review health check results first)
curl -X DELETE http://localhost:8080/api/v1/media/cleanup \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Both require the `media:admin` permission.

## API Reference

All endpoints are prefixed with `/api/v1`.

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/media` | `media:read` | List all media (supports `limit` and `offset`) |
| POST | `/media` | `media:create` | Upload a new media file (multipart, field: `file`) |
| GET | `/media/` | `media:read` | Get single media item (`?q=MEDIA_ID`) |
| PUT | `/media/` | `media:update` | Update media metadata |
| DELETE | `/media/` | `media:delete` | Delete media and S3 objects (`?q=MEDIA_ID`) |
| GET | `/media/health` | `media:admin` | Check for orphaned S3 objects |
| DELETE | `/media/cleanup` | `media:admin` | Delete orphaned S3 objects |
| GET | `/mediadimensions` | `media:read` | List dimension presets |
| POST | `/mediadimensions` | `media:create` | Create dimension preset |
| GET | `/mediadimensions/` | `media:read` | Get single preset (`?q=MD_ID`) |
| PUT | `/mediadimensions/` | `media:update` | Update dimension preset |
| DELETE | `/mediadimensions/` | `media:delete` | Delete dimension preset (`?q=MD_ID`) |

## Notes

- **Image dimension limits.** The pipeline rejects images exceeding 10,000 pixels in either dimension or 50 megapixels total, preventing memory exhaustion from decompression bombs.
- **No upscaling.** Presets larger than the source image are silently skipped. An 800x600 image will not produce a 1920x1080 variant.
- **Duplicate filenames.** Uploading a file with the same name as an existing record returns HTTP 409. Rename the file or delete the existing record first.
- **Variant encoding.** All optimized variants are encoded as WebP at quality 80, regardless of the original format.
- **Srcset format.** The `srcset` field is a JSON-encoded string array of URLs, not the HTML `srcset` attribute format. Parse it as JSON when consuming in your frontend.
- **Rollback on failure.** If any step after S3 upload fails, all uploaded objects and the database record are cleaned up automatically.

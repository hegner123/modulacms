# Media

Upload images and files, configure dimension presets, set focal points, and serve responsive images in your frontend.

## Upload a file

Send a multipart form POST to `/api/v1/media` with a `file` field:

```bash
curl -X POST http://localhost:8080/api/v1/media \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -F "file=@/path/to/hero.jpg"
```

Response (HTTP 201):

```json
{
  "media_id": "01HXK4N2F8RJZGP6VTQY3MCSW9",
  "name": "hero.jpg",
  "mimetype": "image/jpeg",
  "url": "https://cdn.example.com/media/hero.jpg",
  "srcset": "https://cdn.example.com/media/hero-320w.jpg 320w, https://cdn.example.com/media/hero-768w.jpg 768w, https://cdn.example.com/media/hero-1920w.jpg 1920w",
  "focal_x": null,
  "focal_y": null,
  "date_created": "2026-01-15T10:00:00Z",
  "date_modified": "2026-01-15T10:00:00Z"
}
```

When you upload an image and dimension presets exist, ModulaCMS generates a resized variant for each preset and includes all variant URLs in the `srcset` field. Non-image files (PDFs, videos, documents) are stored as-is without optimization.

### Upload with SDKs

**Go SDK:**

```go
f, err := os.Open("/path/to/hero.jpg")
if err != nil {
    // handle error
}
defer f.Close()

media, err := client.MediaUpload.Upload(ctx, f, "hero.jpg", nil)
if err != nil {
    // handle error
}
fmt.Printf("Uploaded: %s (URL: %s)\n", media.MediaID, media.URL)
```

Upload with progress tracking:

```go
stat, _ := f.Stat()
media, err := client.MediaUpload.UploadWithProgress(ctx, f, "hero.jpg", stat.Size(),
    func(sent, total int64) {
        pct := float64(sent) / float64(total) * 100
        fmt.Printf("\r%.0f%%", pct)
    },
    nil,
)
```

**TypeScript SDK (admin):**

```typescript
const fileInput = document.querySelector<HTMLInputElement>('#file')
const file = fileInput!.files![0]

const media = await admin.mediaUpload.upload(file)
console.log(`Uploaded: ${media.media_id} (URL: ${media.url})`)
```

### Custom storage path

By default, files are organized by date (`YYYY/M/filename`). Override with the `path` form field:

```bash
curl -X POST http://localhost:8080/api/v1/media \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -F "file=@/path/to/logo.png" \
  -F "path=branding/logos"
```

This stores the file at `branding/logos/logo.png` in the bucket.

```go
media, err := client.MediaUpload.Upload(ctx, f, "logo.png", &modula.MediaUploadOptions{
    Path: "branding/logos",
})
```

```typescript
const media = await admin.mediaUpload.upload(file, {
  path: 'branding/logos',
})
```

### Supported image types

| Format | Extensions |
|--------|------------|
| PNG | `.png` |
| JPEG | `.jpg`, `.jpeg` |
| GIF | `.gif` |
| WebP | `.webp` |

### Upload limits

| Limit | Value |
|-------|-------|
| Maximum file size | 10 MB |
| Maximum image width | 10,000 pixels |
| Maximum image height | 10,000 pixels |
| Maximum total pixels | 50 megapixels |

> **Good to know**: Uploading a file with the same name as an existing record returns HTTP 409. Rename the file or delete the existing record first.

## Dimension presets

**Dimension presets** define the target sizes for responsive image variants. When you upload an image, ModulaCMS generates a resized variant for each active preset.

Each preset specifies:

| Property | Purpose |
|----------|---------|
| `label` | Human-readable name (e.g., "Thumbnail", "Hero", "Card") |
| `width` | Target width in pixels. Leave empty to scale by height only. |
| `height` | Target height in pixels. Leave empty to scale by width only. |
| `aspect_ratio` | Aspect ratio constraint for cropping (e.g., `"16:9"`, `"1:1"`) |

When both width and height are set, the image is resized to fit within those bounds while maintaining its aspect ratio -- unless an explicit `aspect_ratio` forces cropping.

> **Good to know**: Define your dimension presets before uploading images. Existing images are not retroactively resized when you add new presets. Presets that exceed the source image dimensions are skipped to avoid upscaling.

### Create a preset

```bash
curl -X POST http://localhost:8080/api/v1/mediadimensions \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"label": "social-card", "width": 1200, "height": 630, "aspect_ratio": "1.91:1"}'
```

**Go SDK:**

```go
w := int64(1200)
h := int64(630)
label := "social-card"
ratio := "1.91:1"

dim, err := client.MediaDimensions.Create(ctx, modula.CreateMediaDimensionParams{
    Label:       &label,
    Width:       &w,
    Height:      &h,
    AspectRatio: &ratio,
})
```

**TypeScript SDK (admin):**

```typescript
const dim = await admin.mediaDimensions.create({
  label: 'social-card',
  width: 1200,
  height: 630,
  aspect_ratio: '1.91:1',
})
```

### List presets

```bash
curl http://localhost:8080/api/v1/mediadimensions \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

```json
[
  { "md_id": "01HXK5A1...", "label": "thumbnail", "width": 150, "height": 150, "aspect_ratio": "1:1" },
  { "md_id": "01HXK5B2...", "label": "small", "width": 320, "height": null, "aspect_ratio": null },
  { "md_id": "01HXK5C3...", "label": "medium", "width": 768, "height": null, "aspect_ratio": null },
  { "md_id": "01HXK5D4...", "label": "large", "width": 1280, "height": null, "aspect_ratio": null },
  { "md_id": "01HXK5E5...", "label": "hero", "width": 1920, "height": null, "aspect_ratio": "16:9" }
]
```

## Focal point cropping

Each media item can store a **focal point** -- a normalized position that marks the most important area of the image.

- `(0.0, 0.0)` = top-left corner
- `(0.5, 0.5)` = center (default)
- `(1.0, 1.0)` = bottom-right corner

When a dimension preset requires cropping (e.g., a landscape image resized to a square), ModulaCMS centers the crop on the focal point instead of the image center. This keeps the important part of the image visible across all variants.

Set the focal point when updating media metadata:

```bash
curl -X PUT http://localhost:8080/api/v1/media/ \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "media_id": "01JMKX5V6QNPZ3R8W4T2YH9B0D",
    "name": "photo.jpg",
    "alt": "Aerial view of the company headquarters",
    "focal_x": 0.3,
    "focal_y": 0.25
  }'
```

> **Good to know**: You can set the focal point before or after upload. Re-uploading is not required. Updatable metadata fields: `name`, `display_name`, `alt`, `caption`, `description`, `class`, `focal_x`, `focal_y`.

## Serve responsive images

After upload, each media record includes a `srcset` field with URLs for every dimension variant. Use this data to build responsive `<img>` elements.

### Use the prebuilt srcset

**TypeScript:**

```typescript
function responsiveImage(media: Media): string {
  const alt = media.alt ?? ''

  if (media.srcset) {
    return `<img src="${media.url}" srcset="${media.srcset}" sizes="(max-width: 768px) 100vw, 50vw" alt="${alt}">`
  }

  return `<img src="${media.url}" alt="${alt}">`
}
```

**Go (template helper):**

```go
func responsiveImage(m modula.Media) string {
    alt := ""
    if m.Alt != nil {
        alt = *m.Alt
    }

    if m.Srcset != nil && *m.Srcset != "" {
        return fmt.Sprintf(
            `<img src="%s" srcset="%s" sizes="(max-width: 768px) 100vw, 50vw" alt="%s">`,
            m.URL, *m.Srcset, alt,
        )
    }

    return fmt.Sprintf(`<img src="%s" alt="%s">`, m.URL, alt)
}
```

### Build srcset manually

If you need to construct srcset from dimension presets and a known URL pattern:

```typescript
function buildSrcset(baseUrl: string, dims: MediaDimension[]): string {
  return dims
    .filter(d => d.width !== null)
    .sort((a, b) => (a.width ?? 0) - (b.width ?? 0))
    .map(d => {
      const ext = baseUrl.substring(baseUrl.lastIndexOf('.'))
      const base = baseUrl.substring(0, baseUrl.lastIndexOf('.'))
      return `${base}-${d.width}w${ext} ${d.width}w`
    })
    .join(', ')
}
```

### Use the picture element

For art direction -- serving different crops at different breakpoints -- use the `<picture>` element with your dimension presets:

```html
<picture>
  <source media="(min-width: 1280px)" srcset="hero-1920w.webp">
  <source media="(min-width: 768px)" srcset="hero-768w.webp">
  <img src="hero-320w.webp" alt="Hero image">
</picture>
```

> **Good to know**: All optimized variants are encoded as WebP at quality 80, regardless of the original format. Variant filenames follow the pattern `{basename}-{width}x{height}.webp`.

## Organize with folders

Media assets can be organized into a hierarchical folder structure. Folders support arbitrary nesting, letting you build a directory tree like `branding/logos/` or `blog/2026/march/`.

### Create a folder

```bash
curl -X POST http://localhost:8080/api/v1/media-folders \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"name": "branding"}'
```

Create a nested folder by setting the `parent_id`:

```bash
curl -X POST http://localhost:8080/api/v1/media-folders \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"name": "logos", "parent_id": "01JNRWHSA1LQWZ3X5D8F2G9JKT"}'
```

### Move media between folders

Update the media record's `folder_id` field:

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

### Get the folder tree

```bash
curl http://localhost:8080/api/v1/media-folders/tree \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

## Check media references

Before deleting media, check which content fields reference it:

```bash
curl "http://localhost:8080/api/v1/media/references?q=01HXK4N2F8RJZGP6VTQY3MCSW9" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

**TypeScript SDK (admin):**

```typescript
const refs = await admin.media.getReferences('01HXK4N2F8RJZGP6VTQY3MCSW9' as MediaID)
```

## Storage configuration

Configure the storage backend in `modula.config.json`. Any S3-compatible service works: AWS S3, MinIO, DigitalOcean Spaces, Backblaze B2, Cloudflare R2.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `bucket_region` | string | `"us-east-1"` | S3 region |
| `bucket_media` | string | -- | Bucket name for media assets |
| `bucket_endpoint` | string | -- | S3 API endpoint hostname (without scheme) |
| `bucket_access_key` | string | -- | S3 access key ID |
| `bucket_secret_key` | string | -- | S3 secret access key |
| `bucket_public_url` | string | falls back to `bucket_endpoint` | Public-facing base URL for media links |
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

> **Good to know**: When running in Docker, `bucket_endpoint` typically points to a container hostname (e.g., `minio:9000`) that browsers cannot resolve. Set `bucket_public_url` to the externally reachable address so that media URLs in API responses work in the browser.

## API reference

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| POST | `/api/v1/media` | `media:create` | Upload a new media file (multipart, field: `file`) |
| GET | `/api/v1/media` | `media:read` | List all media (supports `limit` and `offset`) |
| GET | `/api/v1/media/` | `media:read` | Get single media item (`?q=MEDIA_ID`) |
| PUT | `/api/v1/media/` | `media:update` | Update media metadata |
| DELETE | `/api/v1/media/` | `media:delete` | Delete media and S3 objects (`?q=MEDIA_ID`) |
| GET | `/api/v1/media/references/` | `media:read` | Scan for content fields referencing a media item |
| GET | `/api/v1/mediadimensions` | `media:read` | List dimension presets |
| POST | `/api/v1/mediadimensions` | `media:create` | Create dimension preset |
| PUT | `/api/v1/mediadimensions/` | `media:update` | Update dimension preset |
| DELETE | `/api/v1/mediadimensions/` | `media:delete` | Delete dimension preset |
| GET | `/api/v1/media-folders` | `media:read` | List all media folders |
| POST | `/api/v1/media-folders` | `media:create` | Create a new folder |
| GET | `/api/v1/media-folders/tree` | `media:read` | Get the full folder tree |

## Next steps

- [Serving your frontend](serving-your-frontend.md) -- render content and images in a real app
- [Routing](routing.md) -- create routes that map URLs to content
- [Querying content](querying.md) -- filter and paginate content collections

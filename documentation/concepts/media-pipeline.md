# Media Pipeline

ModulaCMS handles media uploads, image optimization, and storage in an automated pipeline. Files are uploaded via multipart form POST, images are optimized at configured dimension presets, and all assets are stored in S3-compatible object storage. The system generates responsive srcset data for images and supports focal point cropping.

## Upload Flow

1. Client sends a `POST /api/v1/media` multipart form request with a `file` field and an optional `path` field.
2. Server validates the file size (max 10 MB) and content type.
3. Server saves the file to a temporary directory.
4. If the file is an image, the optimization pipeline runs (see below).
5. The original file and all optimized variants are uploaded to S3.
6. A `Media` record is created in the database with the S3 URL, srcset, dimensions, and metadata.
7. The temporary files are cleaned up.

### Upload via Go SDK

```go
file, err := os.Open("photo.jpg")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

media, err := client.MediaUpload.Upload(ctx, file, "photo.jpg", nil)
// media.MediaID, media.URL, media.Srcset are populated
```

With progress reporting:

```go
media, err := client.MediaUpload.UploadWithProgress(ctx, file, "photo.jpg", fileSize,
    func(sent, total int64) {
        fmt.Printf("%d / %d bytes\n", sent, total)
    }, nil)
```

### Path Organization

By default, uploaded files are organized by date (`YYYY/M/filename`). Override this with the `path` form field or `MediaUploadOptions.Path`:

```go
media, err := client.MediaUpload.Upload(ctx, file, "shoe.png", &modula.MediaUploadOptions{
    Path: "products/shoes",
})
// Stored at: products/shoes/shoe.png
```

## Supported Image Types

The optimization pipeline processes these image MIME types:

| MIME Type | Extension |
|-----------|-----------|
| `image/png` | `.png` |
| `image/jpeg` | `.jpg`, `.jpeg` |
| `image/gif` | `.gif` |
| `image/webp` | `.webp` |

Non-image files (PDFs, videos, documents) are uploaded to S3 without optimization.

## Upload Limits

| Limit | Value |
|-------|-------|
| Maximum file size | 10 MB |
| Maximum image width | 10,000 pixels |
| Maximum image height | 10,000 pixels |
| Maximum total pixels | 50 megapixels |

Files exceeding these limits are rejected with an appropriate error.

## Dimension Presets

Media dimensions are reusable image size presets. When an image is uploaded, the optimization pipeline generates a resized variant for each defined dimension preset.

```go
type MediaDimension struct {
    MdID        MediaDimensionID `json:"md_id"`
    Label       *string          `json:"label"`
    Width       *int64           `json:"width"`
    Height      *int64           `json:"height"`
    AspectRatio *string          `json:"aspect_ratio"`
}
```

| Field | Purpose |
|-------|---------|
| `Label` | Human-readable name (e.g., `"Thumbnail"`, `"Hero"`, `"Card"`) |
| `Width` | Target width in pixels. Null means unconstrained (scale by height). |
| `Height` | Target height in pixels. Null means unconstrained (scale by width). |
| `AspectRatio` | Aspect ratio constraint for cropping (e.g., `"16:9"`, `"1:1"`). |

At least one of `Width` or `Height` should be specified per preset. When both are set, the image is resized to fit within those bounds while maintaining aspect ratio (unless an explicit `AspectRatio` forces cropping).

### Managing Presets

```bash
# Create a dimension preset
curl -X POST http://localhost:8080/api/v1/mediadimensions \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"label": "Thumbnail", "width": 200, "height": 200, "aspect_ratio": "1:1"}'

# List presets
curl http://localhost:8080/api/v1/mediadimensions \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Define your dimension presets before uploading images. Images uploaded after a preset is created will generate a variant for that preset. Existing images are not retroactively resized when new presets are added.

## Image Optimization

When an image is uploaded and dimension presets exist, the `OptimizeUpload` function:

1. Decodes the source image.
2. For each dimension preset, resizes the image to fit the preset's width/height constraints.
3. If a focal point is set, crops around the focal point rather than center-cropping.
4. Encodes each resized variant in the original format.
5. Returns the list of optimized file paths.

All variants are uploaded to S3 alongside the original, and their URLs are stored as a JSON array in the media record's `Srcset` field.

## Focal Point Cropping

Each media record can store a focal point via `FocalX` and `FocalY` fields. Both are floating-point values between 0.0 and 1.0, representing a normalized position within the image:

- `(0.0, 0.0)` = top-left corner
- `(0.5, 0.5)` = center (default behavior)
- `(1.0, 1.0)` = bottom-right corner

When dimension presets require cropping (e.g., a landscape image resized to a square), the crop region is centered on the focal point rather than the image center. This ensures the important part of the image remains visible across all dimension variants.

Set the focal point when updating media metadata:

```bash
curl -X PUT http://localhost:8080/api/v1/media/ \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"media_id": "01ABC...", "focal_x": 0.3, "focal_y": 0.25}'
```

## S3 Storage

All media files are stored in S3-compatible object storage. Configure the storage backend in `config.json`:

| Config Field | Purpose |
|-------------|---------|
| `bucket_name` | S3 bucket name |
| `bucket_url` | S3 endpoint URL |
| `bucket_region` | AWS region (default: `us-southeast-1`) |
| `bucket_access_key` | S3 access key ID |
| `bucket_secret_key` | S3 secret access key |

Any S3-compatible service works: AWS S3, MinIO, DigitalOcean Spaces, Backblaze B2, etc.

## Srcset and Responsive Images

After image optimization, the media record's `Srcset` field contains a JSON array of S3 URLs -- one for each dimension variant:

```json
[
    "https://s3.example.com/bucket/2026/3/photo-200x200.jpg",
    "https://s3.example.com/bucket/2026/3/photo-800x600.jpg",
    "https://s3.example.com/bucket/2026/3/photo-1200x800.jpg"
]
```

Frontend clients use this data to build HTML `srcset` attributes for responsive image loading. The primary `URL` field contains the original uploaded file's S3 URL.

## Media References

The media reference scan endpoint tells you which content fields reference a given media asset:

```go
type MediaReferenceScanResponse struct {
    MediaID        string               `json:"media_id"`
    References     []MediaReferenceInfo  `json:"references"`
    ReferenceCount int                   `json:"reference_count"`
}

type MediaReferenceInfo struct {
    ContentFieldID string `json:"content_field_id"`
    ContentDataID  string `json:"content_data_id"`
    FieldID        string `json:"field_id"`
}
```

Use this before deleting media to check if any content fields still reference it, or to audit where a media item is used across the site.

## Media Entity

```go
type Media struct {
    MediaID      MediaID   `json:"media_id"`
    Name         *string   `json:"name"`
    DisplayName  *string   `json:"display_name"`
    Alt          *string   `json:"alt"`
    Caption      *string   `json:"caption"`
    Description  *string   `json:"description"`
    Class        *string   `json:"class"`
    Mimetype     *string   `json:"mimetype"`
    Dimensions   *string   `json:"dimensions"`
    URL          URL       `json:"url"`
    Srcset       *string   `json:"srcset"`
    FocalX       *float64  `json:"focal_x"`
    FocalY       *float64  `json:"focal_y"`
    AuthorID     *UserID   `json:"author_id"`
    DateCreated  Timestamp `json:"date_created"`
    DateModified Timestamp `json:"date_modified"`
}
```

Only metadata fields can be updated after upload (name, display name, alt text, caption, description, class, focal point). The underlying file cannot be replaced -- upload a new file instead.

## API Endpoints

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| POST | `/api/v1/media` | `media:create` | Upload a media file (multipart form) |
| GET | `/api/v1/media` | `media:read` | List all media |
| GET | `/api/v1/media/` | `media:read` | Get a single media record (`?q=ID`) |
| PUT | `/api/v1/media/` | `media:update` | Update media metadata |
| DELETE | `/api/v1/media/` | `media:delete` | Delete a media record and its S3 objects |
| GET | `/api/v1/mediadimensions` | `media:read` | List dimension presets |
| POST | `/api/v1/mediadimensions` | `media:create` | Create a dimension preset |
| PUT | `/api/v1/mediadimensions/` | `media:update` | Update a dimension preset |
| DELETE | `/api/v1/mediadimensions/` | `media:delete` | Delete a dimension preset |
| GET | `/api/v1/media/references/` | `media:read` | Scan for content fields referencing a media item |

# Responsive Images

Recipes for working with media assets and building responsive image markup. ModulaCMS generates resized variants of uploaded images based on configured dimension presets. The media API provides the data needed to build `srcset` attributes and `<picture>` elements.

For background on the media system, see [media management](../guides/media-management.md).

## Upload an Image

Upload via multipart form POST. The server validates the file, generates dimension variants, uploads all sizes to S3, and returns the media record.

**curl:**

```bash
curl -X POST http://localhost:8080/api/v1/media \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -F "file=@/path/to/hero.jpg"
```

Response (201):

```json
{
  "media_id": "01HXK4N2F8RJZGP6VTQY3MCSW9",
  "name": "hero.jpg",
  "display_name": "hero.jpg",
  "alt": null,
  "caption": null,
  "description": null,
  "class": null,
  "mimetype": "image/jpeg",
  "dimensions": "{\"width\":1920,\"height\":1080}",
  "url": "https://cdn.example.com/media/hero.jpg",
  "srcset": "https://cdn.example.com/media/hero-320w.jpg 320w, https://cdn.example.com/media/hero-768w.jpg 768w, https://cdn.example.com/media/hero-1920w.jpg 1920w",
  "author_id": "01JMKW8N3QRYZ7T1B5K6F2P4HD",
  "date_created": "2026-01-15T10:00:00Z",
  "date_modified": "2026-01-15T10:00:00Z"
}
```

**Go SDK:**

```go
import (
    "os"

    modula "github.com/hegner123/modulacms/sdks/go"
)

f, err := os.Open("/path/to/hero.jpg")
if err != nil {
    // handle error
}
defer f.Close()

media, err := client.MediaUpload.Upload(ctx, f, "hero.jpg", nil)
if err != nil {
    if modula.IsDuplicateMedia(err) {
        fmt.Println("file with this name already exists")
        return
    }
    // handle error
}

fmt.Printf("Uploaded: %s (URL: %s)\n", media.MediaID, media.URL)
```

Upload with a specific storage path:

```go
media, err := client.MediaUpload.Upload(ctx, f, "hero.jpg", &modula.MediaUploadOptions{
    Path: "images/heroes",
})
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
// Browser environment
const fileInput = document.querySelector<HTMLInputElement>('#file')
const file = fileInput!.files![0]

const media = await admin.mediaUpload.upload(file)
console.log(`Uploaded: ${media.media_id} (URL: ${media.url})`)

// With options
const media = await admin.mediaUpload.upload(file, {
  path: 'images/heroes',
})
```

## List Dimension Presets

Dimension presets define the target sizes for responsive image variants. Each uploaded image is resized to match every active preset.

**curl:**

```bash
curl http://localhost:8080/api/v1/mediadimensions \
  -H "Authorization: Bearer YOUR_API_KEY"
```

Response:

```json
[
  { "md_id": "01HXK5A1...", "label": "thumbnail", "width": 150, "height": 150, "aspect_ratio": "1:1" },
  { "md_id": "01HXK5B2...", "label": "small", "width": 320, "height": null, "aspect_ratio": null },
  { "md_id": "01HXK5C3...", "label": "medium", "width": 768, "height": null, "aspect_ratio": null },
  { "md_id": "01HXK5D4...", "label": "large", "width": 1280, "height": null, "aspect_ratio": null },
  { "md_id": "01HXK5E5...", "label": "hero", "width": 1920, "height": null, "aspect_ratio": "16:9" }
]
```

**Go SDK:**

```go
dims, err := client.MediaDimensions.List(ctx)
if err != nil {
    // handle error
}

for _, d := range dims {
    label := ""
    if d.Label != nil {
        label = *d.Label
    }
    w := int64(0)
    if d.Width != nil {
        w = *d.Width
    }
    fmt.Printf("%s: %dpx wide\n", label, w)
}
```

**TypeScript SDK (read-only):**

```typescript
const dims = await client.listMediaDimensions()

for (const d of dims) {
  console.log(`${d.label}: ${d.width}px wide`)
}
```

**TypeScript SDK (admin):**

```typescript
const dims = await admin.mediaDimensions.list()
```

## Create a Dimension Preset

**curl:**

```bash
curl -X POST http://localhost:8080/api/v1/mediadimensions \
  -H "Authorization: Bearer YOUR_API_KEY" \
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

## Build an HTML srcset from Media Data

The `srcset` field on a media record contains a prebuilt srcset string. Use it directly in your HTML.

**Go (template helper):**

```go
func responsiveImage(m modula.Media) string {
    alt := ""
    if m.Alt != nil {
        alt = *m.Alt
    }

    // Use prebuilt srcset if available
    if m.Srcset != nil && *m.Srcset != "" {
        return fmt.Sprintf(
            `<img src="%s" srcset="%s" sizes="(max-width: 768px) 100vw, 50vw" alt="%s">`,
            m.URL, *m.Srcset, alt,
        )
    }

    // Fallback: single image
    return fmt.Sprintf(`<img src="%s" alt="%s">`, m.URL, alt)
}
```

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

**Build srcset manually from dimension presets and a known URL pattern:**

```typescript
function buildSrcset(baseUrl: string, dims: MediaDimension[]): string {
  return dims
    .filter(d => d.width !== null)
    .sort((a, b) => (a.width ?? 0) - (b.width ?? 0))
    .map(d => {
      // Convention: dimension variants use -{width}w suffix before extension
      const ext = baseUrl.substring(baseUrl.lastIndexOf('.'))
      const base = baseUrl.substring(0, baseUrl.lastIndexOf('.'))
      return `${base}-${d.width}w${ext} ${d.width}w`
    })
    .join(', ')
}
```

## Get Media References

Find which content fields reference a specific media asset. Useful before deleting media to understand the impact.

**curl:**

```bash
curl "http://localhost:8080/api/v1/media/references?q=01HXK4N2F8RJZGP6VTQY3MCSW9" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**TypeScript SDK (admin):**

```typescript
const refs = await admin.media.getReferences('01HXK4N2F8RJZGP6VTQY3MCSW9' as MediaID)
console.log(`Found ${JSON.stringify(refs)} references`)
```

## Delete Media with Cleanup

Delete a media asset and automatically remove all content field references to it. This removes the S3 files for all dimension variants and nullifies any content fields that referenced the asset.

**curl:**

```bash
curl -X DELETE "http://localhost:8080/api/v1/media/?q=01HXK4N2F8RJZGP6VTQY3MCSW9&clean_refs=true" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
// Standard delete (fails if references exist)
err := client.Media.Delete(ctx, modula.MediaID("01HXK4N2F8RJZGP6VTQY3MCSW9"))

// For admin operations with reference cleanup, use the admin media resource
```

**TypeScript SDK (admin):**

```typescript
// Standard delete
await admin.media.remove('01HXK4N2F8RJZGP6VTQY3MCSW9' as MediaID)

// Delete with cleanup (removes S3 files and cleans content references)
await admin.media.deleteWithCleanup('01HXK4N2F8RJZGP6VTQY3MCSW9' as MediaID)
```

## Check Media Health

Scan for orphaned files in the media S3 bucket that have no corresponding database record.

**curl:**

```bash
curl http://localhost:8080/api/v1/media/health \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**TypeScript SDK (admin):**

```typescript
// Check for orphaned files
const health = await admin.media.health()

// Clean up orphaned files
const cleanup = await admin.media.cleanup()
```

## Next Steps

- [Content Modeling](../guides/content-modeling.md) -- define media fields on your datatypes
- [Media Management](../guides/media-management.md) -- full media system documentation

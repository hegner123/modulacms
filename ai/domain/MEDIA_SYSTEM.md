# MEDIA_SYSTEM.md

Domain guide for the media upload and optimization system.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/MEDIA_SYSTEM.md`
**Purpose:** Media upload workflow, image optimization, and S3 integration
**Last Updated:** 2026-01-12

---

## Overview

ModulaCMS handles media uploads with automatic image optimization and S3-compatible object storage. Images are center-cropped and scaled to multiple responsive sizes based on configurable dimensions.

**Supported Formats:** PNG, JPEG, GIF, WebP (decode only)

---

## Upload Workflow

**Flow:**
1. HTTP multipart upload (10 MB limit)
2. Validate filename uniqueness
3. Optimize image (generate responsive sizes)
4. Upload all variants to S3
5. Store srcset URLs in database

**Handler** (`internal/router/mediaUpload.go`):
```go
func MediaUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)  // 10 MB limit
	file, header, err := r.FormFile("file")

	// Check uniqueness
	_, err = d.GetMediaByName(header.Filename)
	if err == nil {
		// File exists
		return
	}

	// Create temp directory
	tmp, err := os.MkdirTemp("/tmp", "media")
	defer exec.Command("rm", "-r", tmp)

	// Process upload
	media.HandleMediaUpload(srcPath, dstPath, config)
}
```

**Processing** (`internal/media/media_upload.go`):
```go
func HandleMediaUpload(srcFile, dstPath string, c config.Config) error {
	// Organize by date: media/YYYY/MM/filename.jpg
	year := time.Now().Year()
	month := time.Now().Month()
	newPath := fmt.Sprintf("%s/%d/%d/%s", bucketDir, year, month, filename)

	// Generate responsive images
	images, err := OptimizeUpload(srcFile, newPath, c)

	// Upload all variants to S3
	for _, imgPath := range *images {
		prep, err := bucket.UploadPrep(imgPath, "media", file)
		bucket.ObjectUpload(s3Session, prep)
	}

	// Update database with srcset
	params.Srcset = db.StringToNullString(string(srcsetJSON))
	d.UpdateMedia(params)
}
```

---

## Image Optimization

**File:** `internal/media/media_optomize.go`

**Steps:**
1. Decode source image
2. Fetch dimensions from database
3. For each dimension:
   - Center-crop to target aspect ratio
   - Bilinear scale to target size
   - Encode with dimension suffix
4. Return array of file paths

**Center Cropping:**
```go
func OptimizeUpload(fSrc, dstPath string, c config.Config) (*[]string, error) {
	// Decode image based on format
	var dImg image.Image
	switch ext {
	case ".png":
		dImg, _ = png.Decode(file)
	case ".jpg", ".jpeg":
		dImg, _ = jpeg.Decode(file)
	case ".webp":
		dImg, _ = webp.Decode(file)
	case ".gif":
		dImg, _ = gif.Decode(file)
	}

	bounds := dImg.Bounds()
	centerX := (bounds.Min.X + bounds.Max.X) / 2
	centerY := (bounds.Min.Y + bounds.Max.Y) / 2

	// For each configured dimension
	dimensions, _ := d.ListMediaDimensions()
	for _, dx := range dimensions {
		cropWidth := int(dx.Width.Int64)
		cropHeight := int(dx.Height.Int64)

		// Crop from center
		x0 := centerX - cropWidth/2
		y0 := centerY - cropHeight/2
		cropRect := image.Rect(x0, y0, x0+cropWidth, y0+cropHeight)
		cropRect = cropRect.Intersect(bounds)  // Prevent out-of-bounds

		// Scale with bilinear interpolation
		dstRect := image.Rect(0, 0, cropWidth, cropHeight)
		img := image.NewRGBA(dstRect)
		draw.BiLinear.Scale(img, dstRect, dImg, cropRect, draw.Over, nil)

		// Save with dimension suffix
		size := fmt.Sprintf("%dx%d", dx.Width.Int64, dx.Height.Int64)
		filename := fmt.Sprintf("%s-%s%s", baseName, size, ext)
		// Encode and save...
	}
}
```

---

## S3-Compatible Storage

**File:** `internal/bucket/object_storage.go`

**Credentials:**
```go
type S3Credintials struct {
	AccessKey string
	SecretKey string
	URL       string
}
```

**Connection:**
```go
func (cs S3Credintials) GetBucket() (*s3.S3, error) {
	sess, _ := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(cs.AccessKey, cs.SecretKey, ""),
		Endpoint:         aws.String(cs.URL),
		Region:           aws.String("us-southeast-1"),
		S3ForcePathStyle: aws.Bool(true),  // Required for some providers
	})
	return s3.New(sess), nil
}
```

**Upload:**
```go
func ObjectUpload(s3 *s3.S3, payload *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	upload, err := s3.PutObject(payload)
	return upload, err
}

func UploadPrep(uploadPath, bucketName string, data *os.File) (*s3.PutObjectInput, error) {
	return &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(uploadPath),
		Body:   data,
		ACL:    aws.String("public-read"),
	}, nil
}
```

**Configuration (config.json):**
```json
{
	"bucket_media": "media",
	"bucket_endpoint": "nyc3.digitaloceanspaces.com",
	"bucket_access_key": "DO00...",
	"bucket_secret_key": "...",
	"bucket_region": "nyc3",
	"bucket_url": "https://nyc3.digitaloceanspaces.com"
}
```

**Compatible Providers:**
- AWS S3
- DigitalOcean Spaces
- Linode Object Storage
- MinIO
- Backblaze B2
- Any S3-compatible API

---

## Database Schema

**Media Table** (`sql/schema/14_media/schema.sql`):
```sql
CREATE TABLE IF NOT EXISTS media (
	media_id INTEGER PRIMARY KEY,
	name TEXT,                   -- Original filename
	display_name TEXT,           -- User-friendly name
	alt TEXT,                    -- Alt text
	caption TEXT,                -- Caption
	description TEXT,            -- Description
	class TEXT,                  -- CSS classes
	mimetype TEXT,               -- image/jpeg
	dimensions TEXT,             -- Dimension reference
	url TEXT UNIQUE,             -- Primary S3 URL
	srcset TEXT,                 -- JSON array of responsive URLs
	author_id INTEGER DEFAULT 1,
	date_created TEXT DEFAULT CURRENT_TIMESTAMP,
	date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

**Go Model:**
```go
type Media struct {
	MediaID      int64
	Name         sql.NullString  // "photo.jpg"
	DisplayName  sql.NullString  // "Hero Photo"
	Alt          sql.NullString  // "A sunset over mountains"
	Caption      sql.NullString  // "Taken in Colorado"
	Description  sql.NullString
	Class        sql.NullString  // "full-width centered"
	Mimetype     sql.NullString  // "image/jpeg"
	Dimensions   sql.NullString
	Url          sql.NullString  // Primary URL
	Srcset       sql.NullString  // JSON: ["url1", "url2", ...]
	AuthorID     int64
	DateCreated  sql.NullString
	DateModified sql.NullString
}
```

---

## Media Dimensions

**Table** (`sql/schema/3_media_dimension/schema.sql`):
```sql
CREATE TABLE IF NOT EXISTS media_dimensions (
	md_id INTEGER PRIMARY KEY,
	label TEXT UNIQUE,       -- "hero", "thumbnail"
	width INTEGER,           -- 1920
	height INTEGER,          -- 1080
	aspect_ratio TEXT        -- "16:9"
);
```

**Go Model:**
```go
type MediaDimensions struct {
	MdID        int64
	Label       sql.NullString  // "hero"
	Width       sql.NullInt64   // 1920
	Height      sql.NullInt64   // 1080
	AspectRatio sql.NullString  // "16:9"
}
```

**Typical Configuration:**
- **hero:** 1920x1080 (16:9) - Full-width images
- **thumbnail:** 300x300 (1:1) - Grid thumbnails
- **card:** 800x600 (4:3) - Card components
- **mobile:** 600x400 (3:2) - Mobile viewports

---

## Srcset Format

Images are stored as JSON arrays in the `srcset` field:

```json
[
	"https://cdn.example.com/media/2024/01/photo-1920x1080.jpg",
	"https://cdn.example.com/media/2024/01/photo-800x600.jpg",
	"https://cdn.example.com/media/2024/01/photo-300x300.jpg"
]
```

**Usage in HTML:**
```html
<img src="photo-1920x1080.jpg"
     srcset="photo-1920x1080.jpg 1920w,
             photo-800x600.jpg 800w,
             photo-300x300.jpg 300w"
     alt="..." />
```

---

## API Endpoints

**Media:**
```
POST   /api/media/upload              # Upload with optimization
GET    /api/media                     # List all media
GET    /api/media?q={media_id}       # Get single record
PUT    /api/media?q={media_id}       # Update metadata
DELETE /api/media?q={media_id}       # Delete record
```

**Dimensions:**
```
GET    /api/media-dimensions          # List all
POST   /api/media-dimensions          # Create
GET    /api/media-dimensions?q={id}  # Get one
PUT    /api/media-dimensions?q={id}  # Update
DELETE /api/media-dimensions?q={id}  # Delete
```

---

## Upload Flow Diagram

```
Client uploads file
    ↓
MediaUploadHandler
    ├─ Validate uniqueness
    ├─ Create temp directory
    └─ Call HandleMediaUpload()
        ↓
OptimizeUpload()
    ├─ Decode source image
    ├─ Fetch dimensions from DB
    └─ For each dimension:
        ├─ Center-crop
        ├─ Bilinear scale
        └─ Encode to disk
        ↓
Upload to S3
    ├─ For each optimized image
    └─ Path: media/YYYY/MM/filename-{w}x{h}.ext
        ↓
Update Database
    ├─ Set srcset (JSON URLs)
    ├─ Set mimetype
    └─ Set date_modified
        ↓
Return Response
    └─ Media record with URLs
```

---

## Practical Example

**Upload:**
```bash
curl -X POST http://localhost:8080/api/media/upload \
  -F "file=@hero.jpg"
```

**Generated Files:**
```
media/2024/01/hero-1920x1080.jpg
media/2024/01/hero-800x600.jpg
media/2024/01/hero-300x300.jpg
```

**Database Record:**
```json
{
	"media_id": 1,
	"name": "hero.jpg",
	"mimetype": "image/jpeg",
	"url": "https://cdn.example.com/media/2024/01/hero-1920x1080.jpg",
	"srcset": [
		"https://cdn.example.com/media/2024/01/hero-1920x1080.jpg",
		"https://cdn.example.com/media/2024/01/hero-800x600.jpg",
		"https://cdn.example.com/media/2024/01/hero-300x300.jpg"
	]
}
```

---

## Related Documentation

- **[CONTENT_MODEL.md](../architecture/CONTENT_MODEL.md)** - How content references media
- **[DATATYPES_AND_FIELDS.md](DATATYPES_AND_FIELDS.md)** - Image field types

---

## Quick Reference

**Key Files:**
- `internal/media/media_upload.go` - Upload handling
- `internal/media/media_optomize.go` - Image processing
- `internal/bucket/object_storage.go` - S3 integration
- `sql/schema/14_media/schema.sql` - Media table
- `sql/schema/3_media_dimension/schema.sql` - Dimensions table

**Key Operations:**
- Upload: Multipart form → optimize → S3 → database
- Optimize: Decode → crop → scale → encode
- S3: Configure credentials → upload with ACL public-read
- Dimensions: Configure in database → used for all uploads

**Supported Formats:**
- PNG (encode/decode)
- JPEG (encode/decode)
- GIF (encode/decode)
- WebP (decode only)

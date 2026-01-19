# MEDIA_PACKAGE.md

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/media/`

**Purpose:** This document explains the media package, which handles image processing, optimization, and S3-compatible object storage uploads for ModulaCMS. It covers the dimension preset system, center cropping algorithm, format support, and the complete upload workflow.

**Last Updated:** 2026-01-12

---

## Overview

The media package provides image processing and storage capabilities for ModulaCMS. It automatically generates multiple optimized versions of uploaded images based on configurable dimension presets, uploads them to S3-compatible object storage, and stores metadata in the database.

**Key Features:**
- Multi-format image support (PNG, JPEG, GIF, WebP)
- Center cropping and BiLinear scaling
- Database-driven dimension presets
- S3-compatible object storage integration (AWS, Linode, DigitalOcean, etc.)
- Automatic srcset generation for responsive images
- Organized year/month directory structure

**Package Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/media/`

---

## Package Structure

```
internal/media/
├── media_assert.go         # Placeholder for media validation (not implemented)
├── media_create.go         # Creates media database entries
├── media_optomize.go       # Image processing (cropping, scaling)
├── media_upload.go         # Complete upload workflow with S3
├── media_optomize_test.go  # Tests for image optimization
└── media_upload_test.go    # Tests for upload workflow
```

**Dependencies:**
- `internal/config` - Configuration management
- `internal/db` - Database abstraction layer
- `internal/bucket` - S3-compatible storage client
- `internal/utility` - Logging and utilities
- `golang.org/x/image/draw` - Image scaling and drawing
- `golang.org/x/image/webp` - WebP format support
- Standard library: `image`, `image/png`, `image/jpeg`, `image/gif`

---

## Database Schema

### Media Table

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/14_media/schema.sql`

```sql
CREATE TABLE IF NOT EXISTS media (
    media_id INTEGER PRIMARY KEY,
    name TEXT,
    display_name TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    mimetype TEXT,
    dimensions TEXT,
    url TEXT UNIQUE,
    srcset TEXT,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

**Key Fields:**
- `name` - Original filename
- `srcset` - JSON array of all optimized image URLs
- `url` - Primary/original image URL
- `dimensions` - Original image dimensions
- `mimetype` - Image MIME type

### Media Dimensions Table

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/3_media_dimension/schema.sql`

```sql
CREATE TABLE IF NOT EXISTS media_dimensions (
    md_id INTEGER PRIMARY KEY,
    label TEXT UNIQUE,
    width INTEGER,
    height INTEGER,
    aspect_ratio TEXT
);
```

**Purpose:** Stores dimension presets that define which image sizes to generate during optimization.

**Example Presets:**
- `thumbnail` - 150x150
- `medium` - 800x600
- `large` - 1920x1080
- `hero` - 2560x1440

---

## Core Functionality

### 1. Image Optimization

**File:** `internal/media/media_optomize.go`

**Primary Function:** `OptimizeUpload(fSrc string, dstPath string, c config.Config) (*[]string, error)`

#### What It Does

Takes a source image and generates multiple optimized versions based on dimension presets stored in the database. Uses center cropping and BiLinear scaling.

#### Algorithm

**Step 1: Load Source Image**
```go
file, err := os.Open(fSrc)
defer file.Close()

// Decode based on file extension
var dImg image.Image
switch ext {
case ".png":
    dImg, err = png.Decode(file)
case ".jpg", ".jpeg":
    dImg, err = jpeg.Decode(file)
case ".webp":
    dImg, err = webp.Decode(file)
case ".gif":
    dImg, err = gif.Decode(file)
}
```

**Step 2: Load Dimension Presets**
```go
// Query database for all dimension presets
dimensionsPTR, err := d.ListMediaDimensions()
dimensions := *dimensionsPTR
```

**Step 3: Center Cropping**

The algorithm calculates the center point of the source image, then crops a rectangle around that center:

```go
bounds := dImg.Bounds()
centerX := (bounds.Min.X + bounds.Max.X) / 2
centerY := (bounds.Min.Y + bounds.Max.Y) / 2

for _, dx := range dimensions {
    cropWidth := int(dx.Width.Int64)
    cropHeight := int(dx.Height.Int64)

    // Calculate crop rectangle centered on image center
    x0 := centerX - cropWidth/2
    y0 := centerY - cropHeight/2
    cropRect := image.Rect(x0, y0, x0+cropWidth, y0+cropHeight)

    // Ensure crop rectangle stays within bounds
    cropRect = cropRect.Intersect(bounds)
}
```

**Why Center Cropping?**
- Focuses on the most important part of the image (typically centered)
- Consistent behavior across all images
- No need for manual crop point configuration

**Step 4: BiLinear Scaling**

```go
var in draw.Scaler = draw.BiLinear
dstRect := image.Rect(0, 0, cropWidth, cropHeight)
img := image.NewRGBA(dstRect)
in.Scale(img, dstRect, dImg, cropRect, draw.Over, nil)
```

**BiLinear vs Other Scalers:**
- **BiLinear** - Good balance of quality and performance (used by ModulaCMS)
- **NearestNeighbor** - Fast but low quality
- **CatmullRom** - High quality but slower

**Step 5: Encode and Save**

```go
for i, im := range images {
    widthString := strconv.FormatInt(dimensions[i].Width.Int64, 10)
    heightString := strconv.FormatInt(dimensions[i].Height.Int64, 10)
    size := widthString + "x" + heightString
    filename := fmt.Sprintf("%s-%v%s", baseName, size, ext)

    f, err := os.Create(filename)
    defer f.Close()

    // Encode based on original format
    switch ext {
    case ".png":
        err = png.Encode(f, im)
    case ".jpg", ".jpeg":
        err = jpeg.Encode(f, im, nil)
    case ".gif":
        err = gif.Encode(f, im, nil)
    }
}
```

**Output Example:**
```
Input:  photo.jpg
Output: photo-150x150.jpg
        photo-800x600.jpg
        photo-1920x1080.jpg
        photo-2560x1440.jpg
```

#### Format Support

| Format | Extension | Decode | Encode | Notes |
|--------|-----------|--------|--------|-------|
| PNG | .png | ✅ | ✅ | Lossless, supports transparency |
| JPEG | .jpg, .jpeg | ✅ | ✅ | Lossy, no transparency |
| WebP | .webp | ✅ | ❌ | Decode only, re-encoded as original format |
| GIF | .gif | ✅ | ✅ | Supports animation (first frame only) |

**Note:** WebP images can be read but are re-encoded to the original format after processing.

---

### 2. S3 Upload Integration

**File:** `internal/media/media_upload.go`

**Primary Function:** `HandleMediaUpload(srcFile string, dstPath string, c config.Config) error`

#### Complete Upload Workflow

**Step 1: Optimize Image**
```go
optimized, err := OptimizeUpload(srcFile, dstPath, c)
// Returns slice of optimized filenames
```

**Step 2: Initialize S3 Session**

The media package uses the bucket package (`internal/bucket/`) for S3-compatible storage:

```go
s3Creds := bucket.S3Credintials{
    AccessKey: c.Bucket_Access_Key,
    SecretKey: c.Bucket_Secret_Key,
    URL:       c.Bucket_Endpoint,
}

s3Session, err := s3Creds.GetBucket()
```

**S3 Configuration (from config):**
```json
{
    "bucket_url": "us-iad-10.linodeobjects.com",
    "bucket_endpoint": "modulacms.us-iad-10.linodeobjects.com",
    "bucket_access_key": "YOUR_ACCESS_KEY",
    "bucket_secret_key": "YOUR_SECRET_KEY"
}
```

**Step 3: Upload All Optimized Versions**

```go
srcset := []string{}
now := time.Now()
year := now.Year()
month := now.Month()

for _, f := range *optimized {
    file, err := os.Open(f)

    // Organize by year/month
    newPath := fmt.Sprintf("%s/%d/%d/%s", bucketDir, year, month, f)
    uploadPath := fmt.Sprintf("https://%s%s", c.Bucket_Endpoint, newPath)

    // Prepare upload
    prep, err := bucket.UploadPrep(newPath, "media", file)

    // Upload to S3
    _, err = bucket.ObjectUpload(s3Session, prep)

    srcset = append(srcset, uploadPath)
}
```

**Upload Path Structure:**
```
/media/2026/01/photo-150x150.jpg
/media/2026/01/photo-800x600.jpg
/media/2026/01/photo-1920x1080.jpg
/media/2026/01/photo-2560x1440.jpg
```

**Why Year/Month Organization?**
- Prevents directory overload (too many files in one directory)
- Enables time-based queries and cleanup
- Mirrors common CDN organization patterns
- Easy to implement backup/archival strategies

**Step 4: Update Database with Srcset**

```go
// Convert srcset array to JSON
newSrcSet, err := json.Marshal(srcset)

// Load existing media entry
rowPtr, err := d.GetMediaByName(baseName)
row := *rowPtr

// Update with srcset
params := MapMediaParams(row)
params.Srcset = db.StringToNullString(string(newSrcSet))
params.DateModified = db.StringToNullString(utility.TimestampReadable())

_, err = d.UpdateMedia(params)
```

**Srcset JSON Example:**
```json
[
  "https://modulacms.us-iad-10.linodeobjects.com/media/2026/01/photo-150x150.jpg",
  "https://modulacms.us-iad-10.linodeobjects.com/media/2026/01/photo-800x600.jpg",
  "https://modulacms.us-iad-10.linodeobjects.com/media/2026/01/photo-1920x1080.jpg",
  "https://modulacms.us-iad-10.linodeobjects.com/media/2026/01/photo-2560x1440.jpg"
]
```

**Frontend Usage:**
The srcset can be used for responsive images:
```html
<img src="photo-800x600.jpg"
     srcset="photo-150x150.jpg 150w,
             photo-800x600.jpg 800w,
             photo-1920x1080.jpg 1920w,
             photo-2560x1440.jpg 2560w"
     sizes="(max-width: 600px) 150px,
            (max-width: 1200px) 800px,
            1920px"
     alt="Photo">
```

---

### 3. Media Creation

**File:** `internal/media/media_create.go`

**Function:** `CreateMedia(name string, c config.Config) string`

#### What It Does

Creates a new media database entry with the given filename. This is typically called before uploading the actual file.

```go
func CreateMedia(name string, c config.Config) string {
    d := db.ConfigDB(c)
    params := db.CreateMediaParams{
        Name: db.StringToNullString(name),
    }
    mediaRow := d.CreateMedia(params)
    return mediaRow.Name.String
}
```

**Workflow:**
1. Call `CreateMedia(filename, config)` to create database entry
2. Call `HandleMediaUpload(filepath, destpath, config)` to process and upload
3. Database entry is updated with srcset after upload completes

---

## Bucket Package Integration

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/bucket/`

The media package depends on the bucket package for S3-compatible storage operations.

### Bucket Package Structure

```
internal/bucket/
├── object_storage.go       # S3 session and upload functions
├── structs.go              # S3Credintials struct
└── object_storage_test.go  # Tests
```

### Key Functions

**GetBucket() - Create S3 Session**
```go
func (cs S3Credintials) GetBucket() (*s3.S3, error) {
    sess, err := session.NewSession(&aws.Config{
        Credentials:      credentials.NewStaticCredentials(cs.AccessKey, cs.SecretKey, ""),
        Endpoint:         aws.String(cs.URL),
        Region:           aws.String("us-southeast-1"),
        S3ForcePathStyle: aws.Bool(true), // Required for non-AWS S3
    })
    return s3.New(sess), nil
}
```

**UploadPrep() - Prepare Upload Payload**
```go
func UploadPrep(uploadPath string, bucketName string, data *os.File) (*s3.PutObjectInput, error) {
    upload := &s3.PutObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(uploadPath),
        Body:   data,
        ACL:    aws.String("public-read"), // Makes uploaded files publicly accessible
    }
    return upload, nil
}
```

**ObjectUpload() - Execute Upload**
```go
func ObjectUpload(s3 *s3.S3, payload *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
    upload, err := s3.PutObject(payload)
    return upload, nil
}
```

### S3-Compatible Providers

ModulaCMS uses the AWS SDK with `S3ForcePathStyle: true`, making it compatible with:
- **AWS S3** - Amazon's object storage
- **Linode Object Storage** - Used in development/testing
- **DigitalOcean Spaces** - Alternative S3-compatible storage
- **Backblaze B2** - Cost-effective storage
- **MinIO** - Self-hosted S3-compatible storage
- **Wasabi** - Hot cloud storage

**Configuration:**
Only the endpoint and credentials need to change. The code remains the same.

---

## Dimension Preset System

### How It Works

1. **Admin configures dimension presets** in the `media_dimensions` table
2. **Upload triggers optimization** which loads all dimension presets from database
3. **Each preset generates one optimized image** with center cropping and scaling
4. **All optimized versions are uploaded** to S3
5. **Srcset is stored** in the media table as JSON

### Managing Dimension Presets

**Query Dimension Presets:**
```sql
-- sqlc: name: ListMediaDimension :many
SELECT * FROM media_dimensions ORDER BY label;
```

**Add New Preset:**
```sql
-- sqlc: name: CreateMediaDimension :one
INSERT INTO media_dimensions (label, width, height, aspect_ratio)
VALUES ('thumbnail', 150, 150, '1:1')
RETURNING *;
```

**Update Preset:**
```sql
-- sqlc: name: UpdateMediaDimension :exec
UPDATE media_dimensions
SET width = 300, height = 300
WHERE label = 'thumbnail';
```

**Delete Preset:**
```sql
-- sqlc: name: DeleteMediaDimension :exec
DELETE FROM media_dimensions WHERE md_id = ?;
```

**Note:** Changing dimension presets does NOT re-process existing images. Only new uploads will use the updated presets.

### Common Dimension Presets

```sql
INSERT INTO media_dimensions (label, width, height, aspect_ratio) VALUES
('thumbnail', 150, 150, '1:1'),
('small', 300, 225, '4:3'),
('medium', 800, 600, '4:3'),
('large', 1280, 720, '16:9'),
('hero', 1920, 1080, '16:9'),
('fullhd', 1920, 1080, '16:9'),
('4k', 3840, 2160, '16:9');
```

### Aspect Ratio Field

The `aspect_ratio` field is **informational only** and not used in calculations. The actual aspect ratio is determined by `width / height`.

**Purpose:**
- Documentation for admins
- Display in TUI/admin panel
- Helps organize presets by aspect ratio

---

## Error Handling

### Common Errors

**1. Unsupported File Format**
```go
return nil, fmt.Errorf("unsupported file extension: %s", ext)
```
**Solution:** Only upload PNG, JPEG, GIF, or WebP files.

**2. Dimensions List is Nil**
```go
return nil, fmt.Errorf("dimensions list is nil")
```
**Solution:** Ensure media_dimensions table has at least one preset.

**3. Failed to Create File**
```go
return nil, fmt.Errorf("error creating file %s: %w", filename, err)
```
**Solution:** Check file permissions in the destination directory.

**4. S3 Upload Failed**
```go
utility.DefaultLogger.Error("failed to upload ", err)
```
**Solution:** Verify S3 credentials, bucket name, and network connectivity.

**5. Database Entry Not Found**
```go
rowPtr, err := d.GetMediaByName(baseName)
if err != nil {
    return err
}
```
**Solution:** Ensure `CreateMedia()` was called before `HandleMediaUpload()`.

### Error Recovery

The media package does NOT automatically clean up partial uploads. If an upload fails midway:

1. **Some optimized files may be uploaded** to S3
2. **Database srcset may be incomplete** or empty
3. **Manual cleanup required** for orphaned S3 objects

**Future Enhancement Opportunity:**
Implement transaction-like behavior with rollback on failure.

---

## Testing

### Test Files

**Test Image:** `internal/media/test.png`

**Optimization Test:** `internal/media/media_optomize_test.go`
```go
func TestOptimize(t *testing.T) {
    c := config.Config{
        Db_Driver: "sqlite",
        Db_Name:   "modula.db",
        Db_URL:    "./modula.db",
    }
    _, err := OptimizeUpload("./test.png", "test.png", c)
    if err != nil {
        t.Fatal(err)
    }
}
```

**Upload Test:** `internal/media/media_upload_test.go`
```go
func TestMediaUpload(t *testing.T) {
    p := config.NewFileProvider("")
    m := config.NewManager(p)
    c, err := m.Config()
    err = HandleMediaUpload("test.png", "test.png", *c)
    if err != nil {
        t.Fatal(err)
    }
}
```

### Running Tests

```bash
# Test optimization only
go test -v ./internal/media -run TestOptimize

# Test complete upload workflow (requires S3 credentials)
go test -v ./internal/media -run TestMediaUpload

# Test all media functionality
go test -v ./internal/media
```

**Prerequisites for Upload Test:**
- Valid S3 credentials in config
- Network access to S3 endpoint
- At least one dimension preset in database

---

## Performance Considerations

### Image Processing Performance

**Factors Affecting Speed:**
1. **Source image size** - Larger images take longer to decode
2. **Number of dimension presets** - More presets = more processing
3. **Scaling algorithm** - BiLinear is balanced; CatmullRom is slower but higher quality
4. **Image format** - PNG is slower than JPEG due to compression

**Optimization Strategies:**
- Limit number of dimension presets to essential sizes only
- Consider async processing for large uploads
- Use JPEG for photographs (smaller files, faster processing)
- Use PNG only when transparency is required

### S3 Upload Performance

**Factors:**
1. **Number of files** - Each dimension preset = one upload
2. **Network bandwidth** - Upload speed depends on connection
3. **S3 endpoint location** - Choose nearest region

**Optimization Strategies:**
- Consider parallel uploads (currently sequential)
- Implement upload progress tracking for large files
- Add retry logic for failed uploads

### Memory Usage

**Current Implementation:**
- Loads entire source image into memory
- Creates new image buffer for each dimension preset
- All images held in memory until encoding completes

**Memory-Intensive Operations:**
```go
dImg image.Image                // Full source image
images := []draw.Image{}        // All optimized versions
img := image.NewRGBA(dstRect)   // One per preset
```

**Large Image Warning:**
A 10MB source image with 5 dimension presets could consume 50MB+ of memory during processing.

---

## Common Workflows

### Adding a New Dimension Preset

**Step 1: Insert Preset into Database**

Via TUI, SQL, or API:
```sql
INSERT INTO media_dimensions (label, width, height, aspect_ratio)
VALUES ('mobile', 480, 320, '3:2');
```

**Step 2: Test with New Upload**

New uploads will automatically include the new dimension.

**Step 3: Re-process Existing Images (Optional)**

If you need existing images in the new size:
1. Download original from S3
2. Re-run `OptimizeUpload()` with updated presets
3. Re-run `HandleMediaUpload()` to upload new versions

**Note:** There is no built-in bulk re-processing feature yet.

---

### Implementing Custom Cropping Strategy

The current implementation uses **center cropping**. To implement custom cropping (e.g., focus on faces):

**Option 1: Modify OptimizeUpload()**

Replace center cropping logic in `internal/media/media_optomize.go:76-86`:

```go
// Current: Center cropping
centerX := (bounds.Min.X + bounds.Max.X) / 2
centerY := (bounds.Min.Y + bounds.Max.Y) / 2

// Alternative: Top-center cropping (for portraits)
centerX := (bounds.Min.X + bounds.Max.X) / 2
centerY := bounds.Min.Y + cropHeight/2
```

**Option 2: Add Crop Point to Media Table**

1. Add `crop_x` and `crop_y` columns to media table
2. Allow user to specify crop point in TUI/admin
3. Load crop point from database in `OptimizeUpload()`
4. Use custom crop point instead of center

**Option 3: Integrate Face Detection**

Use a library like `github.com/esimov/pigo` to detect faces and crop around detected face center.

---

### Supporting Additional Image Formats

To add support for AVIF, TIFF, or other formats:

**Step 1: Import Decoder/Encoder**
```go
import (
    "image/tiff"
    _ "golang.org/x/image/tiff" // TIFF support
)
```

**Step 2: Add to Decode Switch**

In `internal/media/media_optomize.go:52-63`:
```go
case ".tiff", ".tif":
    dImg, err = tiff.Decode(file)
```

**Step 3: Add to Encode Switch**

In `internal/media/media_optomize.go:109-119`:
```go
case ".tiff", ".tif":
    err = tiff.Encode(f, im, nil)
```

**Step 4: Update Format Support Table**

Update documentation to reflect new format support.

---

### Changing S3 Provider

To switch from Linode to DigitalOcean Spaces (or any S3-compatible provider):

**Step 1: Update Configuration**
```json
{
    "bucket_url": "nyc3.digitaloceanspaces.com",
    "bucket_endpoint": "modulacms.nyc3.digitaloceanspaces.com",
    "bucket_access_key": "NEW_ACCESS_KEY",
    "bucket_secret_key": "NEW_SECRET_KEY"
}
```

**Step 2: No Code Changes Required**

The bucket package uses `S3ForcePathStyle: true`, making it compatible with any S3-compatible provider.

**Step 3: Migrate Existing Media (Optional)**

Use AWS CLI or rclone to sync existing media:
```bash
rclone sync linode:modulacms digitalocean:modulacms
```

---

## Integration with Other Packages

### Database Package (internal/db)

**Used For:**
- Loading dimension presets: `d.ListMediaDimensions()`
- Creating media entries: `d.CreateMedia(params)`
- Querying media by name: `d.GetMediaByName(baseName)`
- Updating srcset: `d.UpdateMedia(params)`

**See Also:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/DB_PACKAGE.md`

### Config Package (internal/config)

**Used For:**
- Database connection settings
- S3 credentials and endpoint
- Bucket configuration

**Configuration Fields Used:**
```go
c.Bucket_Access_Key
c.Bucket_Secret_Key
c.Bucket_Endpoint
c.Bucket_Media  // Bucket name/directory
c.Db_Driver
c.Db_URL
```

### Bucket Package (internal/bucket)

**Used For:**
- S3 session creation: `s3Creds.GetBucket()`
- Upload preparation: `bucket.UploadPrep(path, bucket, file)`
- Object upload: `bucket.ObjectUpload(s3Session, prep)`

**Key Types:**
- `bucket.S3Credintials` - Credentials struct
- `*s3.S3` - AWS SDK session
- `*s3.PutObjectInput` - Upload payload

### Utility Package (internal/utility)

**Used For:**
- Logging: `utility.DefaultLogger.Debug()`, `.Error()`, `.Info()`
- Timestamps: `utility.TimestampReadable()`

---

## Future Enhancements

### Potential Improvements

**1. Async Processing**
- Move image processing to background worker
- Return immediately after creating media entry
- Update database when processing completes

**2. Progress Tracking**
- Track optimization progress (1/5 dimensions complete)
- Track upload progress (bytes uploaded / total bytes)
- Expose progress via WebSocket or polling endpoint

**3. Bulk Re-processing**
- Command to re-process all media with updated dimension presets
- Queue-based processing to avoid overload
- Progress reporting and error logging

**4. Smart Cropping**
- Face detection integration
- Content-aware cropping
- User-selectable crop focus points

**5. Format Conversion**
- Automatically convert to modern formats (WebP, AVIF)
- Serve best format based on browser support
- Keep original format as fallback

**6. CDN Integration**
- CloudFlare, Fastly, or Cloudinary integration
- Automatic cache invalidation on update
- On-the-fly image transformations

**7. Metadata Extraction**
- EXIF data parsing
- GPS location extraction
- Camera settings preservation

**8. Validation**
- Implement `media_assert.go` for upload validation
- File size limits
- Dimension requirements
- Format restrictions per content type

---

## Troubleshooting

### Image Processing Issues

**Problem:** "unsupported file extension"
**Cause:** File extension not in supported list (.png, .jpg, .jpeg, .gif, .webp)
**Solution:** Convert image to supported format or add format support

**Problem:** "error decoding image"
**Cause:** Corrupted file or incorrect extension
**Solution:** Verify file integrity; check actual format vs extension

**Problem:** "decoded image is nil"
**Cause:** Decode succeeded but returned nil (rare)
**Solution:** Check file is valid image file; try re-encoding

### Database Issues

**Problem:** "dimensions list is nil"
**Cause:** No dimension presets in media_dimensions table
**Solution:** Insert at least one dimension preset

**Problem:** "failed to list media dimensions"
**Cause:** Database connection error or table doesn't exist
**Solution:** Run migrations; check database connectivity

### S3 Upload Issues

**Problem:** "failed to upload"
**Cause:** Invalid credentials, network error, or bucket doesn't exist
**Solution:** Verify credentials; check bucket name; test connectivity

**Problem:** Partial uploads (some files uploaded, some failed)
**Cause:** Network interruption during upload loop
**Solution:** Check S3 for orphaned files; implement cleanup

### Performance Issues

**Problem:** Slow image processing
**Cause:** Large source images or many dimension presets
**Solution:** Reduce source image size; limit dimension presets; consider async processing

**Problem:** High memory usage
**Cause:** Processing multiple large images simultaneously
**Solution:** Process images sequentially; implement queue system

---

## Related Documentation

**Architecture:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/CONTENT_MODEL.md` - Understanding media references in content

**Packages:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/DB_PACKAGE.md` - Database operations for media
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/CLI_PACKAGE.md` - TUI for media management

**Database:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQL_DIRECTORY.md` - Media table queries
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQLC.md` - Type-safe media queries

**Domain:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/MEDIA_SYSTEM.md` - Media system domain concepts

**Workflows:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_TABLES.md` - Adding media-related tables
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/TESTING.md` - Testing media functionality

---

## Quick Reference

### Key Functions

```go
// Create media database entry
CreateMedia(name string, c config.Config) string

// Optimize image to multiple dimensions
OptimizeUpload(fSrc string, dstPath string, c config.Config) (*[]string, error)

// Complete upload workflow (optimize + upload to S3 + update database)
HandleMediaUpload(srcFile string, dstPath string, c config.Config) error
```

### Supported Formats

| Format | Read | Write | Transparency | Animation |
|--------|------|-------|--------------|-----------|
| PNG | ✅ | ✅ | ✅ | ❌ |
| JPEG | ✅ | ✅ | ❌ | ❌ |
| GIF | ✅ | ✅ | ✅ | ⚠️ (first frame) |
| WebP | ✅ | ❌ | ✅ | ❌ |

### Database Tables

```sql
-- Media entries
media (media_id, name, srcset, url, ...)

-- Dimension presets
media_dimensions (md_id, label, width, height, aspect_ratio)
```

### Configuration Fields

```go
c.Bucket_Access_Key    // S3 access key
c.Bucket_Secret_Key    // S3 secret key
c.Bucket_Endpoint      // S3 endpoint URL
c.Bucket_Media         // Bucket name
c.Db_Driver           // Database driver
c.Db_URL              // Database URL
```

### Directory Structure

```
/media/
└── {year}/
    └── {month}/
        ├── filename-150x150.jpg
        ├── filename-800x600.jpg
        └── filename-1920x1080.jpg
```

### Typical Workflow

1. **Create Entry:** `CreateMedia(filename, config)`
2. **Optimize:** Called internally by HandleMediaUpload
3. **Upload:** `HandleMediaUpload(filepath, destpath, config)`
4. **Result:** Database updated with srcset JSON

---

**Last Updated:** 2026-01-12
**Document Version:** 1.0
**Status:** Complete

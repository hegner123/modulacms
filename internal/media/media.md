# media

Package media provides media file upload, optimization, and S3 storage integration for ModulaCMS. It handles image validation, multi-resolution generation, and transactional upload pipelines with automatic rollback on failure.

## Overview

The media package processes user uploads through a pipeline: validate file type and size, create database record, optimize into multiple dimensions, upload to S3, update database with srcset. If any step fails, all side effects are rolled back including S3 uploads and database records.

Key features: MIME type validation using http.DetectContentType, configurable size limits, prevention of memory exhaustion attacks via dimension validation, WebP/PNG/JPEG/GIF support, center-crop resizing to predefined dimensions, transactional S3 uploads with rollback.

## Constants

### MaxUploadSize

```go
const MaxUploadSize = 10 << 20 // 10 MB
```

Maximum allowed file size for uploads. Files exceeding this limit trigger FileTooLargeError.

### MaxImageWidth

```go
const MaxImageWidth = 10000
```

Maximum allowed image width in pixels. Images exceeding this dimension are rejected to prevent memory exhaustion attacks.

### MaxImageHeight

```go
const MaxImageHeight = 10000
```

Maximum allowed image height in pixels. Images exceeding this dimension are rejected to prevent memory exhaustion attacks.

### MaxImagePixels

```go
const MaxImagePixels = 50000000
```

Maximum total pixel count for an image. Prevents decompression bombs where small file sizes expand to massive memory allocations.

### DefaultS3Region

```go
const DefaultS3Region = "us-southeast-1"
```

Default AWS region for S3 bucket operations when region is not specified in configuration.

### TempDirPrefix

```go
const TempDirPrefix = "modulacms-media"
```

Prefix for temporary directories created during upload processing. Used by os.MkdirTemp.

## Types

### MediaStore

```go
type MediaStore interface {
    GetMediaByName(name string) (*db.Media, error)
    CreateMedia(ctx context.Context, ac audited.AuditContext, params db.CreateMediaParams) (*db.Media, error)
}
```

Consumer-defined interface for media persistence. All three database drivers in ModulaCMS satisfy this interface implicitly. Used by ProcessMediaUpload to remain database-agnostic.

### UploadPipelineFunc

```go
type UploadPipelineFunc func(srcFile string, dstPath string) error
```

Processes a source file into optimized variants and uploads to S3. Callers close over configuration when constructing this function. Returned errors trigger rollback of database records.

### DuplicateMediaError

```go
type DuplicateMediaError struct {
    Name string
}
```

Indicates a media record with the same name already exists. Returned by ProcessMediaUpload when GetMediaByName succeeds before creating a new record.

#### Error

```go
func (e DuplicateMediaError) Error() string
```

Returns formatted error message including the duplicate filename.

### InvalidMediaTypeError

```go
type InvalidMediaTypeError struct {
    ContentType string
}
```

Indicates the uploaded file has an unsupported MIME type. Only image/png, image/jpeg, image/gif, and image/webp are allowed.

#### Error

```go
func (e InvalidMediaTypeError) Error() string
```

Returns formatted error message including the rejected content type.

### FileTooLargeError

```go
type FileTooLargeError struct {
    Size    int64
    MaxSize int64
}
```

Indicates the uploaded file exceeds MaxUploadSize. Returned before any processing begins.

#### Error

```go
func (e FileTooLargeError) Error() string
```

Returns formatted error message including actual size and maximum allowed size.

## Functions

### ProcessMediaUpload

```go
func ProcessMediaUpload(
    ctx context.Context,
    ac audited.AuditContext,
    file multipart.File,
    header *multipart.FileHeader,
    store MediaStore,
    pipeline UploadPipelineFunc,
) (*db.Media, error)
```

Validates, persists, and pipelines a media upload. Returns the created Media record or a typed error for HTTP status code mapping. Performs validation checks: file size against MaxUploadSize, MIME type detection via first 512 bytes, duplicate filename check. Creates database record before running optimization pipeline. On error, returns typed errors: FileTooLargeError, InvalidMediaTypeError, DuplicateMediaError.

### CreateMedia

```go
func CreateMedia(name string, c config.Config) (string, error)
```

Creates a new media record in the database with the given filename. Uses audited context with system user. Returns the created media name or error. Used internally by upload pipeline after validation.

### OptimizeUpload

```go
func OptimizeUpload(srcFile string, dstPath string, c config.Config) (*[]string, error)
```

Generates multiple image resolutions from source file. Loads MediaDimensions from database, decodes source image based on extension, validates dimensions against MaxImageWidth/MaxImageHeight/MaxImagePixels, copies original to dstPath, center-crops and scales to each dimension, encodes to original format, returns paths to all generated files. Skips dimensions larger than source to prevent upscaling. Uses BiLinear scaling algorithm. On encoding failure, deletes partial outputs and returns error.

Supported formats: PNG, JPEG, GIF, WebP. WebP encoding uses lossy compression with quality 80 and preset default.

### HandleMediaUpload

```go
func HandleMediaUpload(srcFile string, dstPath string, c config.Config) error
```

Orchestrates complete media upload workflow. Calls OptimizeUpload to generate variants, establishes S3 session using config credentials, uploads all files to S3 under bucket/year/month/filename structure, tracks uploaded keys for rollback, fetches media record by basename, updates srcset field with JSON array of S3 URLs, rolls back all S3 uploads on any failure. Uses ACL from config.Bucket_Default_ACL or defaults to public-read.

Returns error on optimization failure, S3 session creation failure, upload failure, or database update failure. All S3 uploads are atomic: either all succeed or all are deleted.

### MapMediaParams

```go
func MapMediaParams(a db.Media) db.UpdateMediaParams
```

Maps a Media record to UpdateMediaParams for database updates. Preserves all fields except DateModified which is set to current timestamp. Used by HandleMediaUpload to update srcset after S3 upload.

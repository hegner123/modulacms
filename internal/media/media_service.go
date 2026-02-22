// Package media provides media upload validation, persistence, and optimization pipeline coordination.
// It validates file types and sizes, creates database records, and orchestrates the upload pipeline
// for processing images into optimized variants stored in S3-compatible object storage.
package media

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// MediaStore is the consumer-defined interface for media persistence.
// All three DB drivers (Database, MysqlDatabase, PsqlDatabase) satisfy this implicitly.
type MediaStore interface {
	GetMediaByName(name string) (*db.Media, error)
	CreateMedia(ctx context.Context, ac audited.AuditContext, params db.CreateMediaParams) (*db.Media, error)
	DeleteMedia(ctx context.Context, ac audited.AuditContext, id types.MediaID) error
}

// UploadPipelineFunc processes a source file into optimized variants and uploads to S3.
// The caller closes over config when constructing this function.
type UploadPipelineFunc func(srcFile string, dstPath string) error

// UploadOriginalFunc uploads the original file to S3 and returns the URL and S3 key.
// The S3 key is returned for rollback purposes if subsequent steps fail.
type UploadOriginalFunc func(filePath string) (url string, s3Key string, err error)

// RollbackS3Func deletes an S3 object by key. Used to roll back the original upload on failure.
type RollbackS3Func func(s3Key string)

// DuplicateMediaError indicates a media record with the same name already exists.
type DuplicateMediaError struct {
	Name string
}

// Error returns a formatted error message for duplicate media entries.
func (e DuplicateMediaError) Error() string {
	return fmt.Sprintf("duplicate entry found for %s", e.Name)
}

// FileTooLargeError indicates the uploaded file exceeds the size limit.
type FileTooLargeError struct {
	Size    int64
	MaxSize int64
}

// Error returns a formatted error message for files exceeding size limits.
func (e FileTooLargeError) Error() string {
	return fmt.Sprintf("file size %d exceeds maximum %d", e.Size, e.MaxSize)
}

var imageMIMETypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/gif":  true,
	"image/webp": true,
}

// IsImageMIME reports whether the given content type is an image MIME type
// supported by the optimization pipeline.
func IsImageMIME(contentType string) bool {
	return imageMIMETypes[contentType]
}

// ProcessMediaUpload validates, uploads the original to S3, persists to DB, and
// conditionally runs the image optimization pipeline for supported image types.
//
// Flow:
//  1. Validate size and detect MIME type
//  2. Check for duplicate filename
//  3. Write uploaded file to temp directory
//  4. Upload original to S3 via uploadOriginal callback
//  5. Create DB record (with URL from step 4)
//  6. Run optimization/upload pipeline (images only)
//
// Rollback: S3 original is deleted if DB create or pipeline fails.
// DB record is deleted if pipeline fails after DB create succeeds.
func ProcessMediaUpload(
	ctx context.Context,
	ac audited.AuditContext,
	file multipart.File,
	header *multipart.FileHeader,
	store MediaStore,
	uploadOriginal UploadOriginalFunc,
	rollbackS3 RollbackS3Func,
	pipeline UploadPipelineFunc,
	maxUploadSize int64,
) (*db.Media, error) {
	// Step 1: Validate size
	if header.Size > maxUploadSize {
		return nil, FileTooLargeError{Size: header.Size, MaxSize: maxUploadSize}
	}

	// Step 1: Detect MIME type from first 512 bytes
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read file header: %w", err)
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("seek file: %w", err)
	}

	contentType := http.DetectContentType(buffer)

	// Step 2: Check for duplicate
	_, err = store.GetMediaByName(header.Filename)
	if err == nil {
		return nil, DuplicateMediaError{Name: header.Filename}
	}

	// Step 3: Write uploaded file to temp directory
	tmp, err := os.MkdirTemp("", TempDirPrefix)
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	dstPath := filepath.Join(tmp, header.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, fmt.Errorf("create destination file: %w", err)
	}

	written, copyErr := io.Copy(dst, file)
	if copyErr != nil {
		dst.Close()
		return nil, fmt.Errorf("copy file: %w", copyErr)
	}

	if err := dst.Close(); err != nil {
		return nil, fmt.Errorf("flush temp file: %w", err)
	}

	utility.DefaultLogger.Info("temp file written", "path", dstPath, "bytes", written)

	// Step 4: Upload original to S3
	originalURL, originalKey, err := uploadOriginal(dstPath)
	if err != nil {
		return nil, fmt.Errorf("upload original to S3: %w", err)
	}

	// Step 5: Create DB record with URL and mimetype
	params := db.CreateMediaParams{
		Name:         db.NewNullString(header.Filename),
		Mimetype:     db.NewNullString(contentType),
		URL:          types.URL(originalURL),
		AuthorID:     types.NullableUserID{ID: ac.UserID, Valid: ac.UserID != ""},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	}

	row, err := store.CreateMedia(ctx, ac, params)
	if err != nil {
		rollbackS3(originalKey)
		return nil, fmt.Errorf("create media record: %w", err)
	}

	// Step 6: Run optimization/upload pipeline (images only)
	if IsImageMIME(contentType) {
		if err := pipeline(dstPath, tmp); err != nil {
			rollbackS3(originalKey)
			deleteErr := store.DeleteMedia(ctx, ac, row.MediaID)
			if deleteErr != nil {
				utility.DefaultLogger.Error("failed to rollback media record", deleteErr)
			}
			return nil, fmt.Errorf("media pipeline: %w", err)
		}
	}

	return row, nil
}

// SanitizeMediaPath validates and normalizes a user-provided S3 key path prefix.
// It strips leading/trailing slashes and rejects path traversal or invalid characters.
// An empty input returns the default date-based path (YYYY/M).
func SanitizeMediaPath(raw string) (string, error) {
	trimmed := strings.Trim(raw, "/")
	if trimmed == "" {
		now := time.Now()
		return fmt.Sprintf("%d/%d", now.Year(), now.Month()), nil
	}

	if strings.Contains(trimmed, "..") {
		return "", fmt.Errorf("path traversal not allowed")
	}

	for _, ch := range trimmed {
		switch {
		case ch >= 'a' && ch <= 'z':
		case ch >= 'A' && ch <= 'Z':
		case ch >= '0' && ch <= '9':
		case ch == '/' || ch == '-' || ch == '_' || ch == '.':
		default:
			return "", fmt.Errorf("invalid character in path: %c", ch)
		}
	}

	return trimmed, nil
}

// MediaPathFromURL extracts the directory portion of an S3 key from a full media URL.
// Given URL "https://cdn.example.com/media/products/shoes/image.jpg" and
// prefix "https://cdn.example.com/media/", it returns "products/shoes".
func MediaPathFromURL(fullURL string, endpointPrefix string) string {
	key := strings.TrimPrefix(fullURL, endpointPrefix)
	dir := filepath.Dir(key)
	if dir == "." {
		return ""
	}
	return dir
}

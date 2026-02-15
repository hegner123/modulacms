// Package media provides media upload validation, persistence, and optimization pipeline coordination.
// It validates file types and sizes, creates database records, and orchestrates the upload pipeline
// for processing images into optimized variants stored in S3-compatible object storage.
package media

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

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
}

// UploadPipelineFunc processes a source file into optimized variants and uploads to S3.
// The caller closes over config when constructing this function.
type UploadPipelineFunc func(srcFile string, dstPath string) error

// DuplicateMediaError indicates a media record with the same name already exists.
type DuplicateMediaError struct {
	Name string
}

// Error returns a formatted error message for duplicate media entries.
func (e DuplicateMediaError) Error() string {
	return fmt.Sprintf("duplicate entry found for %s", e.Name)
}

// InvalidMediaTypeError indicates the uploaded file has an unsupported MIME type.
type InvalidMediaTypeError struct {
	ContentType string
}

// Error returns a formatted error message for invalid media types.
func (e InvalidMediaTypeError) Error() string {
	return fmt.Sprintf("invalid file type: %s. Only images allowed.", e.ContentType)
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

var validMIMETypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/gif":  true,
	"image/webp": true,
}

// ProcessMediaUpload validates, persists, and pipelines a media upload.
// Returns the created Media record or a typed error for the caller to map to HTTP status codes.
func ProcessMediaUpload(
	ctx context.Context,
	ac audited.AuditContext,
	file multipart.File,
	header *multipart.FileHeader,
	store MediaStore,
	pipeline UploadPipelineFunc,
) (*db.Media, error) {
	// Validate size
	if header.Size > MaxUploadSize {
		return nil, FileTooLargeError{Size: header.Size, MaxSize: MaxUploadSize}
	}

	// Detect MIME type from first 512 bytes
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
	if !validMIMETypes[contentType] {
		return nil, InvalidMediaTypeError{ContentType: contentType}
	}

	if contentType == "image/webp" {
		utility.DefaultLogger.Info("WebP upload detected - WebP encoding not supported, may fail during optimization")
	}

	// Check for duplicate
	_, err = store.GetMediaByName(header.Filename)
	if err == nil {
		return nil, DuplicateMediaError{Name: header.Filename}
	}

	// Create DB record
	params := db.CreateMediaParams{
		Name:         sql.NullString{String: header.Filename, Valid: true},
		AuthorID:     types.NullableUserID{ID: ac.UserID, Valid: ac.UserID != ""},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	}

	row, err := store.CreateMedia(ctx, ac, params)
	if err != nil {
		return nil, fmt.Errorf("create media record: %w", err)
	}

	// Write uploaded file to temp directory
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
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, fmt.Errorf("copy file: %w", err)
	}

	// Run optimization/upload pipeline
	if err := pipeline(dstPath, tmp); err != nil {
		return nil, fmt.Errorf("media pipeline: %w", err)
	}

	return row, nil
}

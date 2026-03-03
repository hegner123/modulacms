package remote

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hegner123/modulacms/internal/db"
	modula "github.com/hegner123/modulacms/sdks/go"
)

// UploadMedia uploads a file to the remote server via the SDK and returns the
// created media record. This method is not part of the DbDriver interface; the
// TUI discovers it via interface assertion (consumer-defined MediaUploader).
func (r *RemoteDriver) UploadMedia(ctx context.Context, filePath string) (*db.Media, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("remote: open file: %w", err)
	}
	defer f.Close() //nolint: read-only file
	filename := filepath.Base(filePath)
	result, err := r.client.MediaUpload.Upload(ctx, f, filename, nil)
	if err != nil {
		return nil, fmt.Errorf("remote: upload media: %w", err)
	}
	row := mediaToDb(result)
	return &row, nil
}

// UploadMediaWithProgress uploads a file with progress callback support.
// The progressFn receives (bytesSent, totalBytes) and is called during upload.
func (r *RemoteDriver) UploadMediaWithProgress(ctx context.Context, filePath string, progressFn modula.ProgressFunc) (*db.Media, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("remote: open file: %w", err)
	}
	defer f.Close() //nolint: read-only file

	var totalSize int64
	if info, statErr := f.Stat(); statErr == nil {
		totalSize = info.Size()
	}

	filename := filepath.Base(filePath)
	result, err := r.client.MediaUpload.UploadWithProgress(ctx, f, filename, totalSize, progressFn, nil)
	if err != nil {
		return nil, fmt.Errorf("remote: upload media: %w", err)
	}
	row := mediaToDb(result)
	return &row, nil
}

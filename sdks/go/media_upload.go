package modula

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// MediaUploadResource provides media file upload operations via multipart form POST.
// Files are uploaded to the CMS's configured storage backend (local filesystem or S3).
// After upload, the server generates optimized variants at configured dimensions.
// It is accessed via [Client].MediaUpload.
type MediaUploadResource struct {
	http *httpClient
}

// MediaUploadOptions configures optional parameters for media uploads.
type MediaUploadOptions struct {
	// Path sets the S3 key path prefix for organizing media files.
	// Segments are separated by "/". Leading and trailing slashes are stripped server-side.
	// Examples: "products/shoes", "blog/headers".
	// When empty, the server defaults to date-based organization (YYYY/M).
	Path string
}

// ProgressFunc is called during upload with bytes sent so far and total bytes.
// Total is -1 when the size is unknown.
type ProgressFunc func(bytesSent int64, total int64)

// UploadWithProgress uploads a file with progress reporting.
// The progressFn is called periodically with the number of bytes sent and total size.
// If totalSize is 0, the file size is unknown and total is passed as -1 to the callback.
func (m *MediaUploadResource) UploadWithProgress(ctx context.Context, r io.Reader, filename string, totalSize int64, progressFn ProgressFunc, opts *MediaUploadOptions) (*Media, error) {
	if progressFn == nil {
		return m.Upload(ctx, r, filename, opts)
	}
	pr := &progressReader{r: r, total: totalSize, fn: progressFn}
	return m.Upload(ctx, pr, filename, opts)
}

// progressReader wraps an io.Reader and calls fn on each Read with cumulative progress.
type progressReader struct {
	r     io.Reader
	total int64
	sent  int64
	fn    ProgressFunc
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	if n > 0 {
		pr.sent += int64(n)
		t := pr.total
		if t == 0 {
			t = -1
		}
		pr.fn(pr.sent, t)
	}
	return n, err
}

// Upload uploads a file and returns the created [Media] entity.
// The filename is used as the stored file name and for content-type detection.
// Pass nil for opts to use the server's default date-based path organization.
// Returns an [*ApiError] on failure (e.g., unsupported file type, storage full).
func (m *MediaUploadResource) Upload(ctx context.Context, r io.Reader, filename string, opts *MediaUploadOptions) (*Media, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("modula: create form file: %w", err)
	}

	if _, err := io.Copy(part, r); err != nil {
		return nil, fmt.Errorf("modula: copy file data: %w", err)
	}

	if opts != nil && opts.Path != "" {
		pathField, err := writer.CreateFormField("path")
		if err != nil {
			return nil, fmt.Errorf("modula: create path field: %w", err)
		}
		if _, err := pathField.Write([]byte(opts.Path)); err != nil {
			return nil, fmt.Errorf("modula: write path field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("modula: close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.http.baseURL+"/api/v1/media", &buf)
	if err != nil {
		return nil, fmt.Errorf("modula: create upload request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := m.http.doRaw(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body) //nolint: body must be drained
		resp.Body.Close()              //nolint: close on read-only response
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body) //nolint: best-effort read for error message
		return nil, &ApiError{
			StatusCode: resp.StatusCode,
			Body:       string(body),
		}
	}

	var result Media
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("modula: decode upload response: %w", err)
	}
	return &result, nil
}

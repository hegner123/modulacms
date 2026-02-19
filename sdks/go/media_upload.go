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

// MediaUploadResource provides media file upload operations.
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

// Upload uploads a file and returns the created media entity.
// An optional MediaUploadOptions can control the storage path.
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

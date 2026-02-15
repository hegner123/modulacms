package modulacms

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

// Upload uploads a file and returns the created media entity.
func (m *MediaUploadResource) Upload(ctx context.Context, r io.Reader, filename string) (*Media, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("modulacms: create form file: %w", err)
	}

	if _, err := io.Copy(part, r); err != nil {
		return nil, fmt.Errorf("modulacms: copy file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("modulacms: close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.http.baseURL+"/api/v1/media", &buf)
	if err != nil {
		return nil, fmt.Errorf("modulacms: create upload request: %w", err)
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
		return nil, fmt.Errorf("modulacms: decode upload response: %w", err)
	}
	return &result, nil
}

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

// ---------------------------------------------------------------------------
// Admin Media Types
// ---------------------------------------------------------------------------

// AdminMedia represents an uploaded media asset in the admin media library.
// Mirrors [Media] but uses [AdminMediaID] and [AdminMediaFolderID] for the
// admin-specific media namespace. Admin media assets are used for CMS-internal
// purposes (e.g., admin UI images, system icons) and are managed separately
// from user-uploaded public media.
type AdminMedia struct {
	AdminMediaID AdminMediaID        `json:"admin_media_id"`
	Name         *string             `json:"name"`
	DisplayName  *string             `json:"display_name"`
	Alt          *string             `json:"alt"`
	Caption      *string             `json:"caption"`
	Description  *string             `json:"description"`
	Class        *string             `json:"class"`
	Mimetype     *string             `json:"mimetype"`
	Dimensions   *string             `json:"dimensions"`
	URL          URL                 `json:"url"`
	Srcset       *string             `json:"srcset"`
	FocalX       *float64            `json:"focal_x"`
	FocalY       *float64            `json:"focal_y"`
	AuthorID     *UserID             `json:"author_id"`
	FolderID     *AdminMediaFolderID `json:"folder_id"`
	DateCreated  Timestamp           `json:"date_created"`
	DateModified Timestamp           `json:"date_modified"`
	DownloadURL  string              `json:"download_url"`
}

// CreateAdminMediaParams holds parameters for creating a new admin media asset.
// Admin media creation is handled via multipart upload; see [AdminMediaResource.Upload].
// This type is used as the Create type parameter for the generic Resource but is
// not directly used for creation (upload is multipart, not JSON).
type CreateAdminMediaParams struct {
	Name        *string             `json:"name"`
	DisplayName *string             `json:"display_name"`
	Alt         *string             `json:"alt"`
	Caption     *string             `json:"caption"`
	Description *string             `json:"description"`
	Class       *string             `json:"class"`
	FolderID    *AdminMediaFolderID `json:"folder_id"`
}

// UpdateAdminMediaParams contains fields for updating metadata on an existing
// admin media asset. Only metadata fields can be updated; the underlying file
// cannot be replaced. Pointer fields are optional; nil means no change.
type UpdateAdminMediaParams struct {
	AdminMediaID AdminMediaID `json:"admin_media_id"`
	Name         *string      `json:"name"`
	DisplayName  *string      `json:"display_name"`
	Alt          *string      `json:"alt"`
	Caption      *string      `json:"caption"`
	Description  *string      `json:"description"`
	Class        *string      `json:"class"`
	FocalX       *float64     `json:"focal_x"`
	FocalY       *float64     `json:"focal_y"`
}

// ---------------------------------------------------------------------------
// Admin Media Resource
// ---------------------------------------------------------------------------

// AdminMediaResource provides admin media upload and management operations.
// Upload creates new admin media assets via multipart form POST. Standard
// CRUD (list, get, update, delete) is available via [Client].AdminMediaData.
// It is accessed via [Client].AdminMediaUpload.
type AdminMediaResource struct {
	http *httpClient
}

// AdminMediaUploadOptions configures optional parameters for admin media uploads.
type AdminMediaUploadOptions struct {
	// Path sets the S3 key path prefix for organizing admin media files.
	// Segments are separated by "/". Leading and trailing slashes are stripped server-side.
	// When empty, the server defaults to date-based organization (YYYY/M).
	Path string
}

// Upload uploads a file to the admin media library and returns the created
// [AdminMedia] entity. The filename is used as the stored file name and for
// content-type detection. Pass nil for opts to use the server's default
// date-based path organization.
// Returns an [*ApiError] on failure (e.g., unsupported file type, storage full).
func (m *AdminMediaResource) Upload(ctx context.Context, r io.Reader, filename string, opts *AdminMediaUploadOptions) (*AdminMedia, error) {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.http.baseURL+"/api/v1/adminmedia", &buf)
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

	var result AdminMedia
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("modula: decode upload response: %w", err)
	}
	return &result, nil
}

// UploadWithProgress uploads a file to the admin media library with progress
// reporting. The progressFn is called periodically with the number of bytes
// sent and total size. If totalSize is 0, the file size is unknown and total
// is passed as -1 to the callback.
func (m *AdminMediaResource) UploadWithProgress(ctx context.Context, r io.Reader, filename string, totalSize int64, progressFn ProgressFunc, opts *AdminMediaUploadOptions) (*AdminMedia, error) {
	if progressFn == nil {
		return m.Upload(ctx, r, filename, opts)
	}
	pr := &progressReader{r: r, total: totalSize, fn: progressFn}
	return m.Upload(ctx, pr, filename, opts)
}

package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hegner123/modulacms/internal/bucket"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	mediapkg "github.com/hegner123/modulacms/internal/media"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// AdminMediaResponse wraps db.AdminMedia with a computed download URL.
type AdminMediaResponse struct {
	db.AdminMedia
	DownloadURL string `json:"download_url"`
}

func toAdminMediaResponse(m db.AdminMedia) AdminMediaResponse {
	return AdminMediaResponse{
		AdminMedia:  m,
		DownloadURL: "/api/v1/adminmedia/" + string(m.AdminMediaID) + "/download",
	}
}

func toAdminMediaListResponse(items []db.AdminMedia) []AdminMediaResponse {
	resp := make([]AdminMediaResponse, len(items))
	for i, m := range items {
		resp[i] = toAdminMediaResponse(m)
	}
	return resp
}

// --- Collection handler ---

// AdminMediasHandler dispatches collection-level operations for admin media.
func AdminMediasHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			apiListAdminMediaPaginated(w, r, svc)
		} else {
			apiListAdminMedia(w, r, svc)
		}
	case http.MethodPost:
		apiCreateAdminMedia(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// --- Item handler ---

// AdminMediaHandler dispatches item-level operations for admin media.
func AdminMediaHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetAdminMedia(w, r, svc)
	case http.MethodPut:
		apiUpdateAdminMedia(w, r, svc)
	case http.MethodDelete:
		apiDeleteAdminMedia(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// --- GET single ---

func apiGetAdminMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	rawID := r.PathValue("id")
	if rawID == "" {
		rawID = r.URL.Query().Get("q")
	}
	id := types.AdminMediaID(rawID)
	if err := id.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid admin media ID: %v", err), http.StatusBadRequest)
		return
	}

	record, err := d.GetAdminMedia(id)
	if err != nil {
		utility.DefaultLogger.Error("get admin media", err)
		http.Error(w, "admin media not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toAdminMediaResponse(*record))
}

// --- GET list ---

func apiListAdminMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	// Support folder_id filter
	folderFilter := r.URL.Query().Get("folder_id")
	if folderFilter != "" {
		if folderFilter == "unfiled" {
			items, err := d.ListAdminMediaUnfiled()
			if err != nil {
				utility.DefaultLogger.Error("list unfiled admin media", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			mediaItems := make([]db.AdminMedia, 0)
			if items != nil {
				mediaItems = *items
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(toAdminMediaListResponse(mediaItems))
			return
		}

		fid := types.AdminMediaFolderID(folderFilter)
		if err := fid.Validate(); err != nil {
			http.Error(w, fmt.Sprintf("invalid folder_id: %v", err), http.StatusBadRequest)
			return
		}
		folderNullable := types.NullableAdminMediaFolderID{ID: fid, Valid: true}
		items, err := d.ListAdminMediaByFolder(folderNullable)
		if err != nil {
			utility.DefaultLogger.Error("list admin media by folder", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		mediaItems := make([]db.AdminMedia, 0)
		if items != nil {
			mediaItems = *items
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toAdminMediaListResponse(mediaItems))
		return
	}

	items, err := d.ListAdminMedia()
	if err != nil {
		utility.DefaultLogger.Error("list admin media", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mediaItems := make([]db.AdminMedia, 0)
	if items != nil {
		mediaItems = *items
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toAdminMediaListResponse(mediaItems))
}

func apiListAdminMediaPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()
	params := ParsePaginationParams(r)

	// Support folder_id filter with pagination
	folderFilter := r.URL.Query().Get("folder_id")
	if folderFilter != "" {
		if folderFilter == "unfiled" {
			items, err := d.ListAdminMediaUnfiledPaginated(db.PaginationParams{Limit: params.Limit, Offset: params.Offset})
			if err != nil {
				utility.DefaultLogger.Error("list unfiled admin media paginated", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			total, err := d.CountAdminMediaUnfiled()
			if err != nil {
				utility.DefaultLogger.Error("count unfiled admin media", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			mediaItems := make([]db.AdminMedia, 0)
			if items != nil {
				mediaItems = *items
			}
			response := db.PaginatedResponse[AdminMediaResponse]{
				Data:   toAdminMediaListResponse(mediaItems),
				Total:  *total,
				Limit:  params.Limit,
				Offset: params.Offset,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		fid := types.AdminMediaFolderID(folderFilter)
		if err := fid.Validate(); err != nil {
			http.Error(w, fmt.Sprintf("invalid folder_id: %v", err), http.StatusBadRequest)
			return
		}
		folderNullable := types.NullableAdminMediaFolderID{ID: fid, Valid: true}
		items, err := d.ListAdminMediaByFolderPaginated(db.ListAdminMediaByFolderPaginatedParams{
			FolderID: folderNullable,
			Limit:    params.Limit,
			Offset:   params.Offset,
		})
		if err != nil {
			utility.DefaultLogger.Error("list admin media by folder paginated", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		total, err := d.CountAdminMediaByFolder(folderNullable)
		if err != nil {
			utility.DefaultLogger.Error("count admin media by folder", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		mediaItems := make([]db.AdminMedia, 0)
		if items != nil {
			mediaItems = *items
		}
		response := db.PaginatedResponse[AdminMediaResponse]{
			Data:   toAdminMediaListResponse(mediaItems),
			Total:  *total,
			Limit:  params.Limit,
			Offset: params.Offset,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	items, err := d.ListAdminMediaPaginated(db.PaginationParams{Limit: params.Limit, Offset: params.Offset})
	if err != nil {
		utility.DefaultLogger.Error("list admin media paginated", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	total, err := d.CountAdminMedia()
	if err != nil {
		utility.DefaultLogger.Error("count admin media", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mediaItems := make([]db.AdminMedia, 0)
	if items != nil {
		mediaItems = *items
	}
	response := db.PaginatedResponse[AdminMediaResponse]{
		Data:   toAdminMediaListResponse(mediaItems),
		Total:  *total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// --- POST upload ---

func apiCreateAdminMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	if parseErr := r.ParseMultipartForm(c.MaxUploadSize()); parseErr != nil {
		utility.DefaultLogger.Error("parse admin media form", parseErr)
		http.Error(w, "File too large or invalid multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		utility.DefaultLogger.Error("parse admin media file", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ac := middleware.AuditContextFromRequest(r, *c)

	// Parse optional folder_id from form data
	var folderID types.NullableAdminMediaFolderID
	if folderIDStr := r.PostFormValue("folder_id"); folderIDStr != "" {
		fid := types.AdminMediaFolderID(folderIDStr)
		if err := fid.Validate(); err == nil {
			folderID = types.NullableAdminMediaFolderID{ID: fid, Valid: true}
		}
	}

	// Validate S3 config is present (admin bucket falls back to shared bucket)
	accessKey := c.AdminBucketAccessKey()
	secretKey := c.AdminBucketSecretKey()
	if accessKey == "" || secretKey == "" {
		http.Error(w, "S3 storage must be configured for admin media uploads", http.StatusBadRequest)
		return
	}

	// Sanitize path
	mediaPath, pathErr := mediapkg.SanitizeMediaPath(r.PostFormValue("path"))
	if pathErr != nil {
		http.Error(w, pathErr.Error(), http.StatusBadRequest)
		return
	}

	// Create S3 session using admin bucket credentials
	s3Creds := bucket.S3Credentials{
		AccessKey:      accessKey,
		SecretKey:      secretKey,
		URL:            c.AdminBucketEndpointURL(),
		Region:         c.Bucket_Region,
		ForcePathStyle: c.Bucket_Force_Path_Style,
	}
	s3Session, err := s3Creds.GetBucket()
	if err != nil {
		utility.DefaultLogger.Error("admin media S3 session", err)
		http.Error(w, "storage unavailable", http.StatusServiceUnavailable)
		return
	}

	bucketDir := c.AdminBucketMedia()
	acl := c.Bucket_Default_ACL
	if acl == "" {
		acl = "public-read"
	}

	uploadOriginal := func(filePath string) (string, string, error) {
		f, fErr := os.Open(filePath)
		if fErr != nil {
			return "", "", fmt.Errorf("open file for S3 upload: %w", fErr)
		}
		defer f.Close()

		filename := filepath.Base(filePath)
		s3Key := fmt.Sprintf("%s/%s", mediaPath, filename)
		uploadURL := fmt.Sprintf("%s/%s/%s", c.AdminBucketPublicURL(), bucketDir, s3Key)

		prep, prepErr := bucket.UploadPrep(s3Key, bucketDir, f, acl)
		if prepErr != nil {
			return "", "", fmt.Errorf("upload prep: %w", prepErr)
		}

		_, upErr := bucket.ObjectUpload(s3Session, prep)
		if upErr != nil {
			return "", "", fmt.Errorf("S3 upload: %w", upErr)
		}

		return uploadURL, s3Key, nil
	}

	rollbackS3 := func(s3Key string) {
		_, delErr := s3Session.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(bucketDir),
			Key:    aws.String(s3Key),
		})
		if delErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("admin media rollback failed for key %s", s3Key), delErr)
		} else {
			utility.DefaultLogger.Info(fmt.Sprintf("rolled back admin media S3 upload: %s", s3Key))
		}
	}

	pipeline := func(srcFile string, dstPath string) error {
		return mediapkg.HandleAdminMediaUpload(srcFile, dstPath, *c)
	}

	d := svc.Driver()
	row, err := mediapkg.ProcessAdminMediaUpload(
		r.Context(), ac, file, header, d, uploadOriginal, rollbackS3, pipeline, c.MaxUploadSize(), folderID,
	)
	if err != nil {
		utility.DefaultLogger.Error("admin media upload", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toAdminMediaResponse(*row))
}

// --- PUT update metadata ---

func apiUpdateAdminMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	rawID := r.PathValue("id")
	if rawID == "" {
		rawID = r.URL.Query().Get("q")
	}

	var req struct {
		DisplayName string                `json:"display_name"`
		Alt         string                `json:"alt"`
		Caption     string                `json:"caption"`
		Description string                `json:"description"`
		FocalX      types.NullableFloat64 `json:"focal_x"`
		FocalY      types.NullableFloat64 `json:"focal_y"`
		FolderID    *string               `json:"folder_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := types.AdminMediaID(rawID)
	if err := id.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid admin media ID: %v", err), http.StatusBadRequest)
		return
	}

	existing, err := d.GetAdminMedia(id)
	if err != nil {
		http.Error(w, "admin media not found", http.StatusNotFound)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	updateParams := mediapkg.MapAdminMediaParams(*existing)
	updateParams.DisplayName = db.NewNullString(req.DisplayName)
	updateParams.Alt = db.NewNullString(req.Alt)
	updateParams.Caption = db.NewNullString(req.Caption)
	updateParams.Description = db.NewNullString(req.Description)
	updateParams.FocalX = req.FocalX
	updateParams.FocalY = req.FocalY

	_, err = d.UpdateAdminMedia(r.Context(), ac, updateParams)
	if err != nil {
		utility.DefaultLogger.Error("update admin media", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If folder_id was provided, move the media to the specified folder.
	if req.FolderID != nil {
		var folderNullable types.NullableAdminMediaFolderID
		if *req.FolderID != "" {
			fid := types.AdminMediaFolderID(*req.FolderID)
			if err := fid.Validate(); err != nil {
				http.Error(w, fmt.Sprintf("invalid folder_id: %v", err), http.StatusBadRequest)
				return
			}
			folderNullable = types.NullableAdminMediaFolderID{ID: fid, Valid: true}
		}

		moveErr := d.MoveAdminMediaToFolder(r.Context(), ac, db.MoveAdminMediaToFolderParams{
			FolderID:     folderNullable,
			DateModified: types.NewTimestamp(time.Now().UTC()),
			AdminMediaID: id,
		})
		if moveErr != nil {
			utility.DefaultLogger.Error("move admin media to folder during update", moveErr)
			http.Error(w, fmt.Sprintf("metadata updated but folder move failed: %v", moveErr), http.StatusInternalServerError)
			return
		}
	}

	// Re-fetch to reflect all changes
	updated, err := d.GetAdminMedia(id)
	if err != nil {
		utility.DefaultLogger.Error("fetch updated admin media", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toAdminMediaResponse(*updated))
}

// --- DELETE ---

func apiDeleteAdminMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	rawID := r.PathValue("id")
	if rawID == "" {
		rawID = r.URL.Query().Get("q")
	}
	id := types.AdminMediaID(rawID)
	if err := id.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid admin media ID: %v", err), http.StatusBadRequest)
		return
	}

	// Verify record exists
	if _, err := d.GetAdminMedia(id); err != nil {
		http.Error(w, "admin media not found", http.StatusNotFound)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	if err := d.DeleteAdminMedia(r.Context(), ac, id); err != nil {
		utility.DefaultLogger.Error("delete admin media", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// --- Download (pre-signed S3 redirect) ---

func apiDownloadAdminMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	rawID := r.PathValue("id")
	mediaID := types.AdminMediaID(rawID)
	if err := mediaID.Validate(); err != nil {
		http.Error(w, "invalid admin media ID", http.StatusBadRequest)
		return
	}

	m, err := d.GetAdminMedia(mediaID)
	if err != nil {
		utility.DefaultLogger.Error("get admin media for download", err)
		http.Error(w, "admin media not found", http.StatusNotFound)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	filename := adminMediaFilename(m)

	s3Key := extractAdminMediaS3Key(string(m.URL), *c)
	if s3Key == "" {
		http.Error(w, "unable to resolve storage key", http.StatusInternalServerError)
		return
	}

	s3Creds := bucket.S3Credentials{
		AccessKey:      c.AdminBucketAccessKey(),
		SecretKey:      c.AdminBucketSecretKey(),
		URL:            c.AdminBucketEndpointURL(),
		Region:         c.Bucket_Region,
		ForcePathStyle: c.Bucket_Force_Path_Style,
	}
	s3Client, err := s3Creds.GetBucket()
	if err != nil {
		http.Error(w, "storage unavailable", http.StatusServiceUnavailable)
		return
	}

	disposition := fmt.Sprintf(`attachment; filename="%s"`, sanitizeFilename(filename))

	req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket:                     aws.String(c.AdminBucketMedia()),
		Key:                        aws.String(s3Key),
		ResponseContentDisposition: aws.String(disposition),
	})

	presignedURL, err := req.Presign(15 * time.Minute)
	if err != nil {
		http.Error(w, "failed to generate download URL", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, presignedURL, http.StatusFound)
}

// --- Batch move ---

// adminBatchMoveMediaRequest is the JSON body for POST /api/v1/adminmedia/move.
type adminBatchMoveMediaRequest struct {
	MediaIDs []string `json:"media_ids"`
	FolderID *string  `json:"folder_id"`
}

func apiBatchMoveAdminMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	var req adminBatchMoveMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.MediaIDs) == 0 {
		http.Error(w, "media_ids is required and cannot be empty", http.StatusBadRequest)
		return
	}

	if len(req.MediaIDs) > 100 {
		http.Error(w, "batch size cannot exceed 100 items", http.StatusBadRequest)
		return
	}

	// Parse and validate target folder
	var folderID types.NullableAdminMediaFolderID
	if req.FolderID != nil && *req.FolderID != "" {
		fid := types.AdminMediaFolderID(*req.FolderID)
		if err := fid.Validate(); err != nil {
			http.Error(w, fmt.Sprintf("invalid folder_id: %v", err), http.StatusBadRequest)
			return
		}
		// Verify folder exists
		if _, err := d.GetAdminMediaFolder(fid); err != nil {
			http.Error(w, "target folder not found", http.StatusNotFound)
			return
		}
		folderID = types.NullableAdminMediaFolderID{ID: fid, Valid: true}
	}

	// Validate all media IDs upfront
	mediaIDs := make([]types.AdminMediaID, 0, len(req.MediaIDs))
	for _, idStr := range req.MediaIDs {
		mid := types.AdminMediaID(idStr)
		if err := mid.Validate(); err != nil {
			http.Error(w, fmt.Sprintf("invalid media_id %q: %v", idStr, err), http.StatusBadRequest)
			return
		}
		mediaIDs = append(mediaIDs, mid)
	}

	c, err := svc.Config()
	if err != nil {
		utility.DefaultLogger.Error("load config", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)
	now := types.NewTimestamp(time.Now().UTC())

	moved := 0
	for _, mid := range mediaIDs {
		err := d.MoveAdminMediaToFolder(r.Context(), ac, db.MoveAdminMediaToFolderParams{
			FolderID:     folderID,
			DateModified: now,
			AdminMediaID: mid,
		})
		if err != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("move admin media %s to folder", mid), err)
			continue
		}
		moved++
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(batchMoveMediaResponse{Moved: moved})
}

// --- Helpers ---

// adminMediaFilename returns the best available filename for download.
func adminMediaFilename(m *db.AdminMedia) string {
	if m.DisplayName.Valid && m.DisplayName.String != "" {
		return m.DisplayName.String
	}
	if m.Name.Valid && m.Name.String != "" {
		return m.Name.String
	}
	u := string(m.URL)
	if idx := strings.LastIndex(u, "/"); idx >= 0 {
		return u[idx+1:]
	}
	return "download"
}

// extractAdminMediaS3Key strips the public URL prefix and bucket name to recover the S3 key.
func extractAdminMediaS3Key(storedURL string, c config.Config) string {
	prefix := c.AdminBucketPublicURL() + "/" + c.AdminBucketMedia() + "/"
	if strings.HasPrefix(storedURL, prefix) {
		return storedURL[len(prefix):]
	}
	prefix = c.AdminBucketEndpointURL() + "/" + c.AdminBucketMedia() + "/"
	if strings.HasPrefix(storedURL, prefix) {
		return storedURL[len(prefix):]
	}
	return ""
}

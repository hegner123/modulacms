package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/bucket"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	mediapkg "github.com/hegner123/modulacms/internal/media"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// AdminMediaListHandler renders the admin media grid with pagination and folder support.
// When ?folder_id=<id> is set, filters media to that folder.
func AdminMediaListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset := ParsePagination(r)
		d := svc.Driver()

		folderIDStr := r.URL.Query().Get("folder_id")

		var mediaItems []db.AdminMedia
		var total int64

		if folderIDStr != "" {
			folderID := types.AdminMediaFolderID(folderIDStr)
			folderNullable := types.NullableAdminMediaFolderID{ID: folderID, Valid: true}

			items, err := d.ListAdminMediaByFolderPaginated(db.ListAdminMediaByFolderPaginatedParams{
				FolderID: folderNullable,
				Limit:    limit,
				Offset:   offset,
			})
			if err != nil {
				utility.DefaultLogger.Error("failed to list admin media by folder", err)
				http.Error(w, "Failed to load admin media", http.StatusInternalServerError)
				return
			}
			if items != nil {
				mediaItems = *items
			}

			count, err := d.CountAdminMediaByFolder(folderNullable)
			if err != nil {
				utility.DefaultLogger.Error("failed to count admin media by folder", err)
				http.Error(w, "Failed to load admin media", http.StatusInternalServerError)
				return
			}
			if count != nil {
				total = *count
			}
		} else {
			items, err := d.ListAdminMediaUnfiledPaginated(db.PaginationParams{Limit: limit, Offset: offset})
			if err != nil {
				utility.DefaultLogger.Error("failed to list unfiled admin media", err)
				http.Error(w, "Failed to load admin media", http.StatusInternalServerError)
				return
			}
			if items != nil {
				mediaItems = *items
			}
			count, err := d.CountAdminMediaUnfiled()
			if err != nil {
				utility.DefaultLogger.Error("failed to count unfiled admin media", err)
				http.Error(w, "Failed to load admin media", http.StatusInternalServerError)
				return
			}
			if count != nil {
				total = *count
			}
		}

		// In-memory sort
		sortBy := r.URL.Query().Get("sort")
		if sortBy == "" {
			sortBy = "newest"
		}
		sortAdminMediaItems(mediaItems, sortBy)

		baseURL := "/admin/admin-media"
		if folderIDStr != "" {
			baseURL = "/admin/admin-media?folder_id=" + folderIDStr
		}
		if sortBy != "newest" {
			if strings.Contains(baseURL, "?") {
				baseURL += "&sort=" + sortBy
			} else {
				baseURL += "?sort=" + sortBy
			}
		}

		pd := NewPaginationData(total, limit, offset, "#media-grid", baseURL)
		pg := partials.PaginationPageData{
			Current:    pd.Current,
			TotalPages: pd.TotalPages,
			Limit:      pd.Limit,
			Target:     pd.Target,
			BaseURL:    pd.BaseURL,
		}

		// Load child folders for the current directory
		var childFolders []db.AdminMediaFolder
		if folderIDStr != "" {
			folderID := types.AdminMediaFolderID(folderIDStr)
			cf, cfErr := d.ListAdminMediaFoldersByParent(folderID)
			if cfErr == nil && cf != nil {
				childFolders = *cf
			}
		} else {
			rf, rfErr := d.ListAdminMediaFoldersAtRoot()
			if rfErr == nil && rf != nil {
				childFolders = *rf
			}
		}

		// Build breadcrumb for current folder
		var breadcrumb []db.AdminMediaFolder
		var currentFolder *db.AdminMediaFolder
		if folderIDStr != "" {
			folderID := types.AdminMediaFolderID(folderIDStr)
			bc, err := d.GetAdminMediaFolderBreadcrumb(folderID)
			if err == nil {
				breadcrumb = bc
			}
			folder, err := d.GetAdminMediaFolder(folderID)
			if err == nil {
				currentFolder = folder
			}
		}

		if IsNavHTMX(r) {
			csrfToken := CSRFTokenFromContext(r.Context())
			w.Header().Set("HX-Trigger", `{"pageTitle": "Admin Media"}`)
			RenderWithOOB(w, r, pages.AdminMediaListContent(mediaItems, pg, childFolders, folderIDStr, breadcrumb, currentFolder, csrfToken),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.AdminMediaDialogs(csrfToken, folderIDStr)})
			return
		}

		if IsHTMX(r) {
			Render(w, r, pages.AdminMediaGridPartial(childFolders, mediaItems, pg, folderIDStr))
			return
		}

		layout := NewAdminData(r, "Admin Media")
		Render(w, r, pages.AdminMediaList(layout, mediaItems, pg, childFolders, folderIDStr, breadcrumb, currentFolder))
	}
}

// AdminMediaDetailHandler renders the detail/edit view for a single admin media item.
func AdminMediaDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Admin Media ID required", http.StatusBadRequest)
			return
		}

		d := svc.Driver()

		record, err := d.GetAdminMedia(types.AdminMediaID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to get admin media", err)
			http.NotFound(w, r)
			return
		}

		// Load all folders for the folder selector
		allFolders, folderErr := d.ListAdminMediaFolders()
		var folders []db.AdminMediaFolder
		if folderErr == nil && allFolders != nil {
			folders = *allFolders
		}

		// Get current folder info for breadcrumb
		var breadcrumb []db.AdminMediaFolder
		var currentFolder *db.AdminMediaFolder
		currentFolderID := ""
		if record.FolderID.Valid {
			currentFolderID = record.FolderID.ID.String()
			bc, bcErr := d.GetAdminMediaFolderBreadcrumb(record.FolderID.ID)
			if bcErr == nil {
				breadcrumb = bc
			}
			folder, fErr := d.GetAdminMediaFolder(record.FolderID.ID)
			if fErr == nil {
				currentFolder = folder
			}
		}

		layout := NewAdminData(r, "Admin Media Detail")
		csrfToken := CSRFTokenFromContext(r.Context())
		RenderNav(w, r, "Admin Media Detail",
			pages.AdminMediaDetailContent(*record, csrfToken, folders, currentFolderID, breadcrumb, currentFolder),
			pages.AdminMediaDetail(layout, *record, csrfToken, folders, currentFolderID, breadcrumb, currentFolder))
	}
}

// AdminMediaUploadHandler handles multipart file uploads for admin media.
// S3 storage must be configured.
func AdminMediaUploadHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		if parseErr := r.ParseMultipartForm(c.MaxUploadSize()); parseErr != nil {
			utility.DefaultLogger.Error("failed to parse multipart form", parseErr)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Upload failed: file too large", "type": "error"}}`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			http.Error(w, "Failed to parse upload", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		file, header, fileErr := r.FormFile("file")
		if fileErr != nil {
			utility.DefaultLogger.Error("no file in upload", fileErr)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "No file selected", "type": "error"}}`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			http.Error(w, "No file uploaded", http.StatusBadRequest)
			return
		}
		defer file.Close()

		ac := middleware.AuditContextFromRequest(r, *c)

		folderIDStr := r.FormValue("folder_id")
		utility.DefaultLogger.Info("admin media upload", "filename", header.Filename, "folder_id", folderIDStr)
		var folderID types.NullableAdminMediaFolderID
		if folderIDStr != "" {
			folderID = types.NullableAdminMediaFolderID{ID: types.AdminMediaFolderID(folderIDStr), Valid: true}
		}

		// Validate S3 config
		accessKey := c.AdminBucketAccessKey()
		secretKey := c.AdminBucketSecretKey()
		if accessKey == "" || secretKey == "" {
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "S3 storage must be configured", "type": "error"}}`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			http.Error(w, "S3 storage must be configured for admin media uploads", http.StatusBadRequest)
			return
		}

		// Sanitize path
		mediaPath, pathErr := mediapkg.SanitizeMediaPath(r.FormValue("path"))
		if pathErr != nil {
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Invalid path", "type": "error"}}`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
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
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Storage unavailable", "type": "error"}}`)
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
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
		created, err := mediapkg.ProcessAdminMediaUpload(
			r.Context(), ac, file, header, d, uploadOriginal, rollbackS3, pipeline, c.MaxUploadSize(), folderID,
		)
		if err != nil {
			utility.DefaultLogger.Error("admin media upload failed", err)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Upload failed", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.Error(w, "Failed to create admin media", http.StatusInternalServerError)
			return
		}

		w.Header().Set("X-Media-ID", created.AdminMediaID.String())
		w.Header().Set("X-Media-URL", created.URL.String())

		if IsHTMX(r) {
			Render(w, r, pages.AdminMediaUploadResult(*created, ""))
			return
		}
		http.Redirect(w, r, "/admin/admin-media/"+created.AdminMediaID.String(), http.StatusSeeOther)
	}
}

// AdminMediaUpdateHandler updates admin media metadata (alt, caption, description).
func AdminMediaUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Admin Media ID required", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		d := svc.Driver()
		ac := middleware.AuditContextFromRequest(r, *c)
		adminMediaID := types.AdminMediaID(id)

		existing, err := d.GetAdminMedia(adminMediaID)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		updateParams := mediapkg.MapAdminMediaParams(*existing)
		updateParams.DisplayName = db.NewNullString(r.FormValue("display_name"))
		updateParams.Alt = db.NewNullString(r.FormValue("alt"))
		updateParams.Caption = db.NewNullString(r.FormValue("caption"))
		updateParams.Description = db.NewNullString(r.FormValue("description"))

		if fxStr := r.FormValue("focal_x"); fxStr != "" {
			fx, fxErr := strconv.ParseFloat(fxStr, 64)
			if fxErr == nil {
				updateParams.FocalX = types.NullableFloat64{Float64: fx, Valid: true}
			}
		}
		if fyStr := r.FormValue("focal_y"); fyStr != "" {
			fy, fyErr := strconv.ParseFloat(fyStr, 64)
			if fyErr == nil {
				updateParams.FocalY = types.NullableFloat64{Float64: fy, Valid: true}
			}
		}

		if _, err := d.UpdateAdminMedia(r.Context(), ac, updateParams); err != nil {
			utility.DefaultLogger.Error("failed to update admin media", err)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to update admin media", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.Error(w, "Failed to update admin media", http.StatusInternalServerError)
			return
		}

		// Re-crop variants if focal point changed
		focalChanged := updateParams.FocalX.Valid || updateParams.FocalY.Valid
		if focalChanged {
			record, getErr := d.GetAdminMedia(adminMediaID)
			if getErr == nil && mediapkg.IsImageMIME(record.Mimetype.String) {
				// Re-process admin media variants (reuses shared pipeline)
				utility.DefaultLogger.Info("focal point changed, reprocessing admin media variants", "id", id)
			}
		}

		if IsHTMX(r) {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin media updated", "type": "success"}}`)
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/admin/admin-media/"+id, http.StatusSeeOther)
	}
}

// AdminMediaDeleteHandler deletes an admin media item.
// Only HTMX DELETE requests are supported; non-HTMX requests receive 405.
func AdminMediaDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Admin Media ID required", http.StatusBadRequest)
			return
		}

		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		d := svc.Driver()
		ac := middleware.AuditContextFromRequest(r, *c)

		if err := d.DeleteAdminMedia(r.Context(), ac, types.AdminMediaID(id)); err != nil {
			utility.DefaultLogger.Error("failed to delete admin media", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete admin media", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin media deleted", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/admin-media")
		w.WriteHeader(http.StatusOK)
	}
}

// AdminMediaBulkDeleteHandler deletes multiple admin media items by ID.
// Expects a JSON body: {"ids": ["id1", "id2", ...]}.
func AdminMediaBulkDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		var body struct {
			IDs []string `json:"ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if len(body.IDs) == 0 {
			http.Error(w, "No IDs provided", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		d := svc.Driver()
		ac := middleware.AuditContextFromRequest(r, *c)

		var failed int
		for _, idStr := range body.IDs {
			if err := d.DeleteAdminMedia(r.Context(), ac, types.AdminMediaID(idStr)); err != nil {
				utility.DefaultLogger.Error("bulk delete failed for admin media", err, "admin_media_id", idStr)
				failed++
			}
		}

		deleted := len(body.IDs) - failed
		msg := fmt.Sprintf("%d item(s) deleted", deleted)
		toastType := "success"
		if failed > 0 {
			msg = fmt.Sprintf("%d deleted, %d failed", deleted, failed)
			toastType = "warning"
		}

		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast": {"message": "%s", "type": "%s"}}`, msg, toastType))
		w.Header().Set("HX-Redirect", "/admin/admin-media")
		w.WriteHeader(http.StatusOK)
	}
}

// AdminMediaFolderCreateHandler handles POST /admin/admin-media-folders.
func AdminMediaFolderCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Folder name is required", "type": "error"}}`)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		d := svc.Driver()

		var parentID types.NullableAdminMediaFolderID
		if pidStr := r.FormValue("parent_id"); pidStr != "" {
			pid := types.AdminMediaFolderID(pidStr)
			if err := pid.Validate(); err != nil {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Invalid parent folder", "type": "error"}}`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if _, err := d.GetAdminMediaFolder(pid); err != nil {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Parent folder not found", "type": "error"}}`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			parentID = types.NullableAdminMediaFolderID{ID: pid, Valid: true}

			breadcrumb, err := d.GetAdminMediaFolderBreadcrumb(pid)
			if err != nil {
				utility.DefaultLogger.Error("check admin media folder depth", err)
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to validate folder depth", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if len(breadcrumb)+1 > 10 {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Maximum folder depth of 10 exceeded", "type": "error"}}`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		if err := d.ValidateAdminMediaFolderName(name, parentID); err != nil {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "A folder with that name already exists here", "type": "error"}}`)
			w.WriteHeader(http.StatusConflict)
			return
		}

		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ac := middleware.AuditContextFromRequest(r, *c)
		now := types.NewTimestamp(time.Now().UTC())

		if _, err := d.CreateAdminMediaFolder(r.Context(), ac, db.CreateAdminMediaFolderParams{
			Name:         name,
			ParentID:     parentID,
			DateCreated:  now,
			DateModified: now,
		}); err != nil {
			utility.DefaultLogger.Error("failed to create admin media folder", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to create folder", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Folder created", "type": "success"}}`)

		if IsHTMX(r) {
			w.Header().Set("HX-Redirect", "/admin/admin-media")
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/admin/admin-media", http.StatusSeeOther)
	}
}

// AdminMediaFolderUpdateHandler handles POST /admin/admin-media-folders/{id}.
func AdminMediaFolderUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Folder ID required", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		d := svc.Driver()
		folderID := types.AdminMediaFolderID(id)

		existing, err := d.GetAdminMediaFolder(folderID)
		if err != nil {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Folder not found", "type": "error"}}`)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			name = existing.Name
		}

		parentID := existing.ParentID
		parentChanged := false
		if r.Form.Has("parent_id") {
			parentChanged = true
			pidStr := r.FormValue("parent_id")
			if pidStr == "" {
				parentID = types.NullableAdminMediaFolderID{}
			} else {
				pid := types.AdminMediaFolderID(pidStr)
				if err := pid.Validate(); err != nil {
					w.Header().Set("HX-Trigger", `{"showToast": {"message": "Invalid parent folder", "type": "error"}}`)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				parentID = types.NullableAdminMediaFolderID{ID: pid, Valid: true}
			}
		}

		if parentChanged {
			if err := d.ValidateAdminMediaFolderMove(folderID, parentID); err != nil {
				w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast": {"message": "%s", "type": "error"}}`, err.Error()))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		nameChanged := name != existing.Name
		if nameChanged || parentChanged {
			if err := d.ValidateAdminMediaFolderName(name, parentID); err != nil {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "A folder with that name already exists here", "type": "error"}}`)
				w.WriteHeader(http.StatusConflict)
				return
			}
		}

		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ac := middleware.AuditContextFromRequest(r, *c)

		if _, err := d.UpdateAdminMediaFolder(r.Context(), ac, db.UpdateAdminMediaFolderParams{
			AdminFolderID: folderID,
			Name:          name,
			ParentID:      parentID,
			DateModified:  types.NewTimestamp(time.Now().UTC()),
		}); err != nil {
			utility.DefaultLogger.Error("failed to update admin media folder", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to update folder", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Folder updated", "type": "success"}}`)

		if IsHTMX(r) {
			w.Header().Set("HX-Redirect", "/admin/admin-media")
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/admin/admin-media", http.StatusSeeOther)
	}
}

// AdminMediaFolderDeleteHandler handles DELETE /admin/admin-media-folders/{id}.
func AdminMediaFolderDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Folder ID required", http.StatusBadRequest)
			return
		}

		d := svc.Driver()
		folderID := types.AdminMediaFolderID(id)

		if _, err := d.GetAdminMediaFolder(folderID); err != nil {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Folder not found", "type": "error"}}`)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		children, err := d.ListAdminMediaFoldersByParent(folderID)
		if err != nil {
			utility.DefaultLogger.Error("list admin media child folders", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to check folder contents", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if children != nil && len(*children) > 0 {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Cannot delete folder: contains subfolders", "type": "error"}}`)
			w.WriteHeader(http.StatusConflict)
			return
		}

		folderNullable := types.NullableAdminMediaFolderID{ID: folderID, Valid: true}
		mediaCount, err := d.CountAdminMediaByFolder(folderNullable)
		if err != nil {
			utility.DefaultLogger.Error("count admin media in folder", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to check folder contents", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if mediaCount != nil && *mediaCount > 0 {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Cannot delete folder: contains media items", "type": "error"}}`)
			w.WriteHeader(http.StatusConflict)
			return
		}

		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ac := middleware.AuditContextFromRequest(r, *c)

		if err := d.DeleteAdminMediaFolder(r.Context(), ac, folderID); err != nil {
			utility.DefaultLogger.Error("failed to delete admin media folder", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete folder", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Folder deleted", "type": "success"}}`)

		if IsHTMX(r) {
			w.Header().Set("HX-Redirect", "/admin/admin-media")
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/admin/admin-media", http.StatusSeeOther)
	}
}

// AdminMediaMoveToFolderHandler handles POST /admin/admin-media/move/{id}.
// Moves an admin media item to a folder (or root if folder_id is empty).
func AdminMediaMoveToFolderHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Admin Media ID required", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		d := svc.Driver()
		adminMediaID := types.AdminMediaID(id)

		var folderID types.NullableAdminMediaFolderID
		if fidStr := r.FormValue("folder_id"); fidStr != "" {
			fid := types.AdminMediaFolderID(fidStr)
			if err := fid.Validate(); err != nil {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Invalid folder", "type": "error"}}`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if _, err := d.GetAdminMediaFolder(fid); err != nil {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Folder not found", "type": "error"}}`)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			folderID = types.NullableAdminMediaFolderID{ID: fid, Valid: true}
		}

		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ac := middleware.AuditContextFromRequest(r, *c)

		utility.DefaultLogger.Info("moving admin media to folder", "admin_media_id", adminMediaID, "folder_id", folderID)
		if err := d.MoveAdminMediaToFolder(r.Context(), ac, db.MoveAdminMediaToFolderParams{
			FolderID:     folderID,
			DateModified: types.NewTimestamp(time.Now().UTC()),
			AdminMediaID: adminMediaID,
		}); err != nil {
			utility.DefaultLogger.Error("failed to move admin media to folder", err, "admin_media_id", adminMediaID, "folder_id", folderID)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to move admin media", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		utility.DefaultLogger.Info("admin media moved to folder successfully", "admin_media_id", adminMediaID, "folder_id", folderID)
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin media moved", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// sortAdminMediaItems sorts admin media items in-memory by the given key.
func sortAdminMediaItems(items []db.AdminMedia, sortBy string) {
	switch sortBy {
	case "oldest":
		sort.Slice(items, func(i, j int) bool {
			return items[i].DateCreated.String() < items[j].DateCreated.String()
		})
	case "name-asc":
		sort.Slice(items, func(i, j int) bool {
			return strings.ToLower(items[i].Name.String) < strings.ToLower(items[j].Name.String)
		})
	case "name-desc":
		sort.Slice(items, func(i, j int) bool {
			return strings.ToLower(items[i].Name.String) > strings.ToLower(items[j].Name.String)
		})
	case "type":
		sort.Slice(items, func(i, j int) bool {
			return items[i].Mimetype.String < items[j].Mimetype.String
		})
	default: // "newest"
		sort.Slice(items, func(i, j int) bool {
			return items[i].DateCreated.String() > items[j].DateCreated.String()
		})
	}
}

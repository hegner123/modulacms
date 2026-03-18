package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/media"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// MediaListHandler renders the media grid with pagination and folder support.
// When ?picker=true is set, returns a minimal grid without the full page layout
// for use in the media picker modal.
// When ?folder_id=<id> is set, filters media to that folder.
func MediaListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset := ParsePagination(r)
		d := svc.Driver()

		folderIDStr := r.URL.Query().Get("folder_id")
		picker := r.URL.Query().Get("picker") == "true"

		var mediaItems []db.Media
		var total int64

		if folderIDStr != "" {
			// Filter by folder
			folderID := types.MediaFolderID(folderIDStr)
			folderNullable := types.NullableMediaFolderID{ID: folderID, Valid: true}

			items, err := d.ListMediaByFolderPaginated(db.ListMediaByFolderPaginatedParams{
				FolderID: folderNullable,
				Limit:    limit,
				Offset:   offset,
			})
			if err != nil {
				utility.DefaultLogger.Error("failed to list media by folder", err)
				http.Error(w, "Failed to load media", http.StatusInternalServerError)
				return
			}
			if items != nil {
				mediaItems = *items
			}

			count, err := d.CountMediaByFolder(folderNullable)
			if err != nil {
				utility.DefaultLogger.Error("failed to count media by folder", err)
				http.Error(w, "Failed to load media", http.StatusInternalServerError)
				return
			}
			if count != nil {
				total = *count
			}
		} else {
			// Root: show only unfiled media (no folder assigned)
			items, err := d.ListMediaUnfiledPaginated(db.PaginationParams{Limit: limit, Offset: offset})
			if err != nil {
				utility.DefaultLogger.Error("failed to list unfiled media", err)
				http.Error(w, "Failed to load media", http.StatusInternalServerError)
				return
			}
			if items != nil {
				mediaItems = *items
			}
			count, err := d.CountMediaUnfiled()
			if err != nil {
				utility.DefaultLogger.Error("failed to count unfiled media", err)
				http.Error(w, "Failed to load media", http.StatusInternalServerError)
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
		sortMediaItems(mediaItems, sortBy)

		baseURL := "/admin/media"
		if picker {
			baseURL = "/admin/media?picker=true"
		} else if folderIDStr != "" {
			baseURL = "/admin/media?folder_id=" + folderIDStr
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
		var childFolders []db.MediaFolder
		if folderIDStr != "" {
			folderID := types.MediaFolderID(folderIDStr)
			cf, cfErr := d.ListMediaFoldersByParent(folderID)
			if cfErr == nil && cf != nil {
				childFolders = *cf
			}
		} else {
			rf, rfErr := d.ListMediaFoldersAtRoot()
			if rfErr == nil && rf != nil {
				childFolders = *rf
			}
		}

		// Build breadcrumb for current folder
		var breadcrumb []db.MediaFolder
		var currentFolder *db.MediaFolder
		if folderIDStr != "" {
			folderID := types.MediaFolderID(folderIDStr)
			bc, err := d.GetMediaFolderBreadcrumb(folderID)
			if err == nil {
				breadcrumb = bc
			}
			folder, err := d.GetMediaFolder(folderID)
			if err == nil {
				currentFolder = folder
			}
		}

		if IsNavHTMX(r) {
			csrfToken := CSRFTokenFromContext(r.Context())
			w.Header().Set("HX-Trigger", `{"pageTitle": "Media"}`)
			RenderWithOOB(w, r, pages.MediaListContent(mediaItems, pg, childFolders, folderIDStr, breadcrumb, currentFolder, csrfToken),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.MediaDialogs(csrfToken, folderIDStr)})
			return
		}

		if picker || IsHTMX(r) {
			Render(w, r, pages.MediaGridPartial(childFolders, mediaItems, pg, picker, folderIDStr))
			return
		}

		layout := NewAdminData(r, "Media")
		Render(w, r, pages.MediaList(layout, mediaItems, pg, childFolders, folderIDStr, breadcrumb, currentFolder))
	}
}

// MediaDetailHandler renders the detail/edit view for a single media item.
func MediaDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Media ID required", http.StatusBadRequest)
			return
		}

		d := svc.Driver()

		record, err := svc.Media.GetMedia(r.Context(), types.MediaID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to get media", err)
			http.NotFound(w, r)
			return
		}

		// Load all folders for the folder selector
		allFolders, folderErr := d.ListMediaFolders()
		var folders []db.MediaFolder
		if folderErr == nil && allFolders != nil {
			folders = *allFolders
		}

		// Get current folder info for breadcrumb
		var breadcrumb []db.MediaFolder
		var currentFolder *db.MediaFolder
		currentFolderID := ""
		if record.FolderID.Valid {
			currentFolderID = record.FolderID.ID.String()
			bc, bcErr := d.GetMediaFolderBreadcrumb(record.FolderID.ID)
			if bcErr == nil {
				breadcrumb = bc
			}
			folder, fErr := d.GetMediaFolder(record.FolderID.ID)
			if fErr == nil {
				currentFolder = folder
			}
		}

		layout := NewAdminData(r, "Media Detail")
		csrfToken := CSRFTokenFromContext(r.Context())
		RenderNav(w, r, "Media Detail",
			pages.MediaDetailContent(*record, csrfToken, folders, currentFolderID, breadcrumb, currentFolder),
			pages.MediaDetail(layout, *record, csrfToken, folders, currentFolderID, breadcrumb, currentFolder))
	}
}

// MediaUploadHandler handles multipart file uploads via the service layer.
// S3 storage must be configured -- placeholder DB-only records are no longer created.
func MediaUploadHandler(svc *service.Registry) http.HandlerFunc {
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

		// Extract display name from filename (without extension)
		filename := header.Filename
		displayName := filename
		if dotIdx := strings.LastIndex(filename, "."); dotIdx > 0 {
			displayName = filename[:dotIdx]
		}

		folderIDStr := r.FormValue("folder_id")
		utility.DefaultLogger.Info("media upload", "filename", header.Filename, "folder_id", folderIDStr)
		var folderID types.NullableMediaFolderID
		if folderIDStr != "" {
			folderID = types.NullableMediaFolderID{ID: types.MediaFolderID(folderIDStr), Valid: true}
		}

		created, err := svc.Media.Upload(r.Context(), ac, service.UploadMediaParams{
			File:        file,
			Header:      header,
			Path:        r.FormValue("path"),
			Alt:         r.FormValue("alt"),
			Description: r.FormValue("description"),
			DisplayName: displayName,
			FolderID:    folderID,
		})
		if err != nil {
			if service.IsValidation(err) || service.IsConflict(err) {
				if IsHTMX(r) {
					w.Header().Set("HX-Trigger", `{"showToast": {"message": "Upload failed", "type": "error"}}`)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			utility.DefaultLogger.Error("failed to upload media", err)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Upload failed", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.Error(w, "Failed to create media", http.StatusInternalServerError)
			return
		}

		w.Header().Set("X-Media-ID", created.MediaID.String())
		w.Header().Set("X-Media-URL", created.URL.String())

		if IsHTMX(r) {
			Render(w, r, pages.MediaUploadResult(*created, ""))
			return
		}
		http.Redirect(w, r, "/admin/media/"+created.MediaID.String(), http.StatusSeeOther)
	}
}

// MediaUpdateHandler updates media metadata (alt, caption, description).
func MediaUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Media ID required", http.StatusBadRequest)
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

		ac := middleware.AuditContextFromRequest(r, *c)

		params := service.UpdateMediaMetadataParams{
			MediaID:     types.MediaID(id),
			DisplayName: r.FormValue("display_name"),
			Alt:         r.FormValue("alt"),
			Caption:     r.FormValue("caption"),
			Description: r.FormValue("description"),
		}

		if fxStr := r.FormValue("focal_x"); fxStr != "" {
			fx, fxErr := strconv.ParseFloat(fxStr, 64)
			if fxErr == nil {
				params.FocalX = types.NullableFloat64{Float64: fx, Valid: true}
			}
		}
		if fyStr := r.FormValue("focal_y"); fyStr != "" {
			fy, fyErr := strconv.ParseFloat(fyStr, 64)
			if fyErr == nil {
				params.FocalY = types.NullableFloat64{Float64: fy, Valid: true}
			}
		}

		if _, err := svc.Media.UpdateMediaMetadata(r.Context(), ac, params); err != nil {
			if service.IsNotFound(err) {
				http.NotFound(w, r)
				return
			}
			utility.DefaultLogger.Error("failed to update media", err)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to update media", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.Error(w, "Failed to update media", http.StatusInternalServerError)
			return
		}

		// Re-crop variants if focal point changed
		focalChanged := params.FocalX.Valid || params.FocalY.Valid
		if focalChanged {
			record, getErr := svc.Media.GetMedia(r.Context(), types.MediaID(id))
			if getErr == nil && media.IsImageMIME(record.Mimetype.String) {
				if rpErr := svc.Media.ReprocessMediaVariants(r.Context(), ac, types.MediaID(id)); rpErr != nil {
					utility.DefaultLogger.Error("failed to reprocess media variants after focal point change", rpErr)
				}
			}
		}

		if IsHTMX(r) {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Media updated", "type": "success"}}`)
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/admin/media/"+id, http.StatusSeeOther)
	}
}

// MediaDeleteHandler deletes a media item including S3 cleanup.
// Only HTMX DELETE requests are supported; non-HTMX requests receive 405.
func MediaDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Media ID required", http.StatusBadRequest)
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

		if err := svc.Media.DeleteMedia(r.Context(), ac, types.MediaID(id)); err != nil {
			utility.DefaultLogger.Error("failed to delete media", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete media", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Media deleted", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/media")
		w.WriteHeader(http.StatusOK)
	}
}

// MediaBulkDeleteHandler deletes multiple media items by ID.
// Expects a JSON body: {"ids": ["id1", "id2", ...]}.
func MediaBulkDeleteHandler(svc *service.Registry) http.HandlerFunc {
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

		ac := middleware.AuditContextFromRequest(r, *c)

		var failed int
		for _, id := range body.IDs {
			if err := svc.Media.DeleteMedia(r.Context(), ac, types.MediaID(id)); err != nil {
				utility.DefaultLogger.Error("bulk delete failed for media", err, "media_id", id)
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
		w.Header().Set("HX-Redirect", "/admin/media")
		w.WriteHeader(http.StatusOK)
	}
}

// sortMediaItems sorts media items in-memory by the given key.
func sortMediaItems(items []db.Media, sortBy string) {
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

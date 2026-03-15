package handlers

import (
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
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
			// All media
			items, count, err := svc.Media.ListMediaPaginated(r.Context(), limit, offset)
			if err != nil {
				utility.DefaultLogger.Error("failed to list media", err)
				http.Error(w, "Failed to load media", http.StatusInternalServerError)
				return
			}
			if items != nil {
				mediaItems = *items
			}
			if count != nil {
				total = *count
			}
		}

		baseURL := "/admin/media"
		if picker {
			baseURL = "/admin/media?picker=true"
		} else if folderIDStr != "" {
			baseURL = "/admin/media?folder_id=" + folderIDStr
		}

		pd := NewPaginationData(total, limit, offset, "#media-grid", baseURL)
		pg := partials.PaginationPageData{
			Current:    pd.Current,
			TotalPages: pd.TotalPages,
			Limit:      pd.Limit,
			Target:     pd.Target,
			BaseURL:    pd.BaseURL,
		}

		// Load folders for the sidebar
		allFolders, folderErr := d.ListMediaFolders()
		var folders []db.MediaFolder
		if folderErr == nil && allFolders != nil {
			folders = *allFolders
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
			RenderWithOOB(w, r, pages.MediaListContent(mediaItems, pg, folders, folderIDStr, breadcrumb, currentFolder, csrfToken),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.MediaUploadDialog(csrfToken, folderIDStr)})
			return
		}

		if picker || IsHTMX(r) {
			Render(w, r, pages.MediaGridPartial(mediaItems, pg, picker))
			return
		}

		layout := NewAdminData(r, "Media")
		Render(w, r, pages.MediaList(layout, mediaItems, pg, folders, folderIDStr, breadcrumb, currentFolder))
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
					w.Header().Set("HX-Trigger", `{"showToast": {"message": "Upload failed: `+err.Error()+`", "type": "error"}}`)
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
		w.WriteHeader(http.StatusOK)
	}
}

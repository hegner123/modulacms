package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// MediaFolderCreateHandler handles POST /admin/media-folders.
// Creates a new media folder and returns the refreshed folder tree partial.
func MediaFolderCreateHandler(svc *service.Registry) http.HandlerFunc {
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

		var parentID types.NullableMediaFolderID
		if pidStr := r.FormValue("parent_id"); pidStr != "" {
			pid := types.MediaFolderID(pidStr)
			if err := pid.Validate(); err != nil {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Invalid parent folder", "type": "error"}}`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if _, err := d.GetMediaFolder(pid); err != nil {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Parent folder not found", "type": "error"}}`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			parentID = types.NullableMediaFolderID{ID: pid, Valid: true}

			breadcrumb, err := d.GetMediaFolderBreadcrumb(pid)
			if err != nil {
				utility.DefaultLogger.Error("check folder depth", err)
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

		if err := d.ValidateMediaFolderName(name, parentID); err != nil {
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

		if _, err := d.CreateMediaFolder(r.Context(), ac, db.CreateMediaFolderParams{
			Name:         name,
			ParentID:     parentID,
			DateCreated:  now,
			DateModified: now,
		}); err != nil {
			utility.DefaultLogger.Error("failed to create media folder", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to create folder", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Folder created", "type": "success"}}`)

		// Reload the media page to show the new folder in the grid
		if IsHTMX(r) {
			w.Header().Set("HX-Redirect", "/admin/media")
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
	}
}

// MediaFolderUpdateHandler handles POST /admin/media-folders/{id}.
// Renames or moves a media folder and returns the refreshed folder tree partial.
func MediaFolderUpdateHandler(svc *service.Registry) http.HandlerFunc {
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
		folderID := types.MediaFolderID(id)

		existing, err := d.GetMediaFolder(folderID)
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
				parentID = types.NullableMediaFolderID{}
			} else {
				pid := types.MediaFolderID(pidStr)
				if err := pid.Validate(); err != nil {
					w.Header().Set("HX-Trigger", `{"showToast": {"message": "Invalid parent folder", "type": "error"}}`)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				parentID = types.NullableMediaFolderID{ID: pid, Valid: true}
			}
		}

		if parentChanged {
			if err := d.ValidateMediaFolderMove(folderID, parentID); err != nil {
				w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast": {"message": "%s", "type": "error"}}`, err.Error()))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		nameChanged := name != existing.Name
		if nameChanged || parentChanged {
			if err := d.ValidateMediaFolderName(name, parentID); err != nil {
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

		if _, err := d.UpdateMediaFolder(r.Context(), ac, db.UpdateMediaFolderParams{
			FolderID:     folderID,
			Name:         name,
			ParentID:     parentID,
			DateModified: types.NewTimestamp(time.Now().UTC()),
		}); err != nil {
			utility.DefaultLogger.Error("failed to update media folder", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to update folder", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Folder updated", "type": "success"}}`)

		if IsHTMX(r) {
			w.Header().Set("HX-Redirect", "/admin/media")
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
	}
}

// MediaFolderDeleteHandler handles DELETE /admin/media-folders/{id}.
// Deletes a media folder (must be empty) and returns the refreshed folder tree partial.
func MediaFolderDeleteHandler(svc *service.Registry) http.HandlerFunc {
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
		folderID := types.MediaFolderID(id)

		if _, err := d.GetMediaFolder(folderID); err != nil {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Folder not found", "type": "error"}}`)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		children, err := d.ListMediaFoldersByParent(folderID)
		if err != nil {
			utility.DefaultLogger.Error("list child folders", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to check folder contents", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if children != nil && len(*children) > 0 {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Cannot delete folder: contains subfolders", "type": "error"}}`)
			w.WriteHeader(http.StatusConflict)
			return
		}

		folderNullable := types.NullableMediaFolderID{ID: folderID, Valid: true}
		mediaCount, err := d.CountMediaByFolder(folderNullable)
		if err != nil {
			utility.DefaultLogger.Error("count media in folder", err)
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

		if err := d.DeleteMediaFolder(r.Context(), ac, folderID); err != nil {
			utility.DefaultLogger.Error("failed to delete media folder", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete folder", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Folder deleted", "type": "success"}}`)

		if IsHTMX(r) {
			w.Header().Set("HX-Redirect", "/admin/media")
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
	}
}

// MediaMoveToFolderHandler handles POST /admin/media/move/{id}.
// Moves a media item to a folder (or root if folder_id is empty).
func MediaMoveToFolderHandler(svc *service.Registry) http.HandlerFunc {
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

		d := svc.Driver()
		mediaID := types.MediaID(id)

		var folderID types.NullableMediaFolderID
		if fidStr := r.FormValue("folder_id"); fidStr != "" {
			fid := types.MediaFolderID(fidStr)
			if err := fid.Validate(); err != nil {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Invalid folder", "type": "error"}}`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if _, err := d.GetMediaFolder(fid); err != nil {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Folder not found", "type": "error"}}`)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			folderID = types.NullableMediaFolderID{ID: fid, Valid: true}
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

		utility.DefaultLogger.Info("moving media to folder", "media_id", mediaID, "folder_id", folderID)
		if err := d.MoveMediaToFolder(r.Context(), ac, db.MoveMediaToFolderParams{
			FolderID:     folderID,
			DateModified: types.NewTimestamp(time.Now().UTC()),
			MediaID:      mediaID,
		}); err != nil {
			utility.DefaultLogger.Error("failed to move media to folder", err, "media_id", mediaID, "folder_id", folderID)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to move media", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		utility.DefaultLogger.Info("media moved to folder successfully", "media_id", mediaID, "folder_id", folderID)
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Media moved", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// renderFolderTree loads all folders and renders the folder tree partial.
func renderFolderTree(w http.ResponseWriter, r *http.Request, d db.DbDriver, activeFolderID string) {
	allFolders, err := d.ListMediaFolders()
	if err != nil {
		utility.DefaultLogger.Error("failed to list media folders", err)
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to load folders", "type": "error"}}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var folders []db.MediaFolder
	if allFolders != nil {
		folders = *allFolders
	}

	csrfToken := CSRFTokenFromContext(r.Context())
	Render(w, r, partials.MediaFolderTree(folders, activeFolderID, csrfToken))
}

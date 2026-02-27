package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// MediaListHandler renders the media grid with pagination.
// When ?picker=true is set, returns a minimal grid without the full page layout
// for use in the media picker modal.
func MediaListHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset := ParsePagination(r)

		items, err := driver.ListMediaPaginated(db.PaginationParams{
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to list media", err)
			http.Error(w, "Failed to load media", http.StatusInternalServerError)
			return
		}

		total, err := driver.CountMedia()
		if err != nil {
			utility.DefaultLogger.Error("failed to count media", err)
			http.Error(w, "Failed to load media", http.StatusInternalServerError)
			return
		}

		var mediaItems []db.Media
		if items != nil {
			mediaItems = *items
		}

		picker := r.URL.Query().Get("picker") == "true"
		baseURL := "/admin/media"
		if picker {
			baseURL = "/admin/media?picker=true"
		}

		pd := NewPaginationData(*total, limit, offset, "#media-grid", baseURL)
		pg := partials.PaginationPageData{
			Current:    pd.Current,
			TotalPages: pd.TotalPages,
			Limit:      pd.Limit,
			Target:     pd.Target,
			BaseURL:    pd.BaseURL,
		}

		if IsNavHTMX(r) {
			csrfToken := CSRFTokenFromContext(r.Context())
			w.Header().Set("HX-Trigger", `{"pageTitle": "Media"}`)
			RenderWithOOB(w, r, pages.MediaListContent(mediaItems, pg),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.MediaUploadDialog(csrfToken)})
			return
		}

		if picker || IsHTMX(r) {
			Render(w, r, pages.MediaGridPartial(mediaItems, pg, picker))
			return
		}

		layout := NewAdminData(r, "Media")
		Render(w, r, pages.MediaList(layout, mediaItems, pg))
	}
}

// MediaDetailHandler renders the detail/edit view for a single media item.
func MediaDetailHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Media ID required", http.StatusBadRequest)
			return
		}

		media, err := driver.GetMedia(types.MediaID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to get media", err)
			http.NotFound(w, r)
			return
		}

		layout := NewAdminData(r, "Media Detail")
		csrfToken := CSRFTokenFromContext(r.Context())
		RenderNav(w, r, "Media Detail", pages.MediaDetailContent(*media, csrfToken), pages.MediaDetail(layout, *media, csrfToken))
	}
}

// MediaUploadHandler handles multipart file uploads.
// Parses the uploaded file and creates a media record with metadata.
func MediaUploadHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 32 MB max upload size
		if parseErr := r.ParseMultipartForm(32 << 20); parseErr != nil {
			utility.DefaultLogger.Error("failed to parse multipart form", parseErr)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Upload failed: file too large", "type": "error"}}`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			http.Error(w, "Failed to parse upload", http.StatusBadRequest)
			return
		}

		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
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

		filename := header.Filename
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		// Extract display name from filename (without extension)
		displayName := filename
		if dotIdx := strings.LastIndex(filename, "."); dotIdx > 0 {
			displayName = filename[:dotIdx]
		}

		now := types.NewTimestamp(time.Now())
		ac := audited.Ctx(
			types.NodeID(cfg.Node_ID),
			user.UserID,
			middleware.RequestIDFromContext(r.Context()),
			clientIP(r),
		)

		// Create media record. URL will be set by the media upload pipeline
		// when S3/local storage is configured; for now store the filename as a placeholder.
		params := db.CreateMediaParams{
			Name:         db.NewNullString(filename),
			DisplayName:  db.NewNullString(displayName),
			Alt:          db.NewNullString(r.FormValue("alt")),
			Caption:      db.NullString{},
			Description:  db.NewNullString(r.FormValue("description")),
			Class:        db.NullString{},
			Mimetype:     db.NewNullString(contentType),
			Dimensions:   db.NullString{},
			URL:          types.URL("/uploads/" + filename),
			Srcset:       db.NullString{},
			FocalX:       types.NullableFloat64{},
			FocalY:       types.NullableFloat64{},
			AuthorID:     types.NullableUserID{ID: user.UserID, Valid: true},
			DateCreated:  now,
			DateModified: now,
		}

		created, createErr := driver.CreateMedia(r.Context(), ac, params)
		if createErr != nil {
			utility.DefaultLogger.Error("failed to create media record", createErr)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Upload failed", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.Error(w, "Failed to create media", http.StatusInternalServerError)
			return
		}

		if IsHTMX(r) {
			Render(w, r, pages.MediaUploadResult(*created, ""))
			return
		}
		http.Redirect(w, r, "/admin/media/"+created.MediaID.String(), http.StatusSeeOther)
	}
}

// MediaUpdateHandler updates media metadata (alt, caption, description).
func MediaUpdateHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
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

		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		existing, getErr := driver.GetMedia(types.MediaID(id))
		if getErr != nil {
			utility.DefaultLogger.Error("media not found for update", getErr)
			http.NotFound(w, r)
			return
		}

		ac := audited.Ctx(
			types.NodeID(cfg.Node_ID),
			user.UserID,
			middleware.RequestIDFromContext(r.Context()),
			clientIP(r),
		)

		params := db.UpdateMediaParams{
			MediaID:      existing.MediaID,
			Name:         existing.Name,
			DisplayName:  updateNullString(r.FormValue("display_name"), existing.DisplayName),
			Alt:          updateNullString(r.FormValue("alt"), existing.Alt),
			Caption:      updateNullString(r.FormValue("caption"), existing.Caption),
			Description:  updateNullString(r.FormValue("description"), existing.Description),
			Class:        existing.Class,
			Mimetype:     existing.Mimetype,
			Dimensions:   existing.Dimensions,
			URL:          existing.URL,
			Srcset:       existing.Srcset,
			FocalX:       existing.FocalX,
			FocalY:       existing.FocalY,
			AuthorID:     existing.AuthorID,
			DateCreated:  existing.DateCreated,
			DateModified: types.NewTimestamp(time.Now()),
		}

		if _, updateErr := driver.UpdateMedia(r.Context(), ac, params); updateErr != nil {
			utility.DefaultLogger.Error("failed to update media", updateErr)
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

// MediaDeleteHandler deletes a media item.
// Only HTMX DELETE requests are supported; non-HTMX requests receive 405.
func MediaDeleteHandler(driver db.DbDriver, mgr *config.Manager) http.HandlerFunc {
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

		cfg, err := mgr.Config()
		if err != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ac := audited.Ctx(
			types.NodeID(cfg.Node_ID),
			user.UserID,
			middleware.RequestIDFromContext(r.Context()),
			clientIP(r),
		)

		if deleteErr := driver.DeleteMedia(r.Context(), ac, types.MediaID(id)); deleteErr != nil {
			utility.DefaultLogger.Error("failed to delete media", deleteErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete media", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Media deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// updateNullString returns a NullString from the form value if non-empty,
// otherwise falls back to the existing value.
func updateNullString(formVal string, existing db.NullString) db.NullString {
	trimmed := strings.TrimSpace(formVal)
	if trimmed != "" {
		return db.NewNullString(trimmed)
	}
	return existing
}

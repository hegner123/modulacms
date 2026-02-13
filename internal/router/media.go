package router

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/media"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// MediasHandler handles CRUD operations that do not require a specific media ID.
func MediasHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			apiListMediaPaginated(w, r, c)
		} else {
			apiListMedia(w, c)
		}
	case http.MethodPost:
		apiCreateMedia(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// MediaHandler handles CRUD operations for specific media items.
func MediaHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetMedia(w, r, c)
	case http.MethodPut:
		apiUpdateMedia(w, r, c)
	case http.MethodDelete:
		apiDeleteMedia(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetMedia handles GET requests for a single media item
func apiGetMedia(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	mID := types.MediaID(q)
	if err := mID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	media, err := d.GetMedia(mID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(media)
	return nil
}

// apiListMedia handles GET requests for listing media items
func apiListMedia(w http.ResponseWriter, c config.Config) error {
	d := db.ConfigDB(c)

	mediaList, err := d.ListMedia()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(mediaList)
	return nil
}

// apiCreateMedia handles POST requests to upload and create a new media item.
// Accepts multipart form with a "file" field. Delegates validation, DB creation,
// and pipeline execution to media.ProcessMediaUpload.
func apiCreateMedia(w http.ResponseWriter, r *http.Request, c config.Config) {
	err := r.ParseMultipartForm(media.MaxUploadSize)
	if err != nil {
		utility.DefaultLogger.Error("parse form", err)
		http.Error(w, "File too large or invalid multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		utility.DefaultLogger.Error("parse file", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	d := db.ConfigDB(c)
	ac := middleware.AuditContextFromRequest(r, c)

	pipeline := func(srcFile string, dstPath string) error {
		return media.HandleMediaUpload(srcFile, dstPath, c)
	}

	row, err := media.ProcessMediaUpload(r.Context(), ac, file, header, d, pipeline)
	if err != nil {
		var dupErr media.DuplicateMediaError
		var mimeErr media.InvalidMediaTypeError
		var sizeErr media.FileTooLargeError

		switch {
		case errors.As(err, &dupErr):
			utility.DefaultLogger.Error("duplicate media", err)
			http.Error(w, err.Error(), http.StatusConflict)
		case errors.As(err, &mimeErr):
			utility.DefaultLogger.Error("invalid content type", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.As(err, &sizeErr):
			utility.DefaultLogger.Error("file too large", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			utility.DefaultLogger.Error("create media", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(row)
}

// apiUpdateMedia handles PUT requests to update an existing media item
func apiUpdateMedia(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var updateMedia db.UpdateMediaParams
	err := json.NewDecoder(r.Body).Decode(&updateMedia)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	updatedMedia, err := d.UpdateMedia(r.Context(), ac, updateMedia)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedMedia)
	return nil
}

// apiDeleteMedia handles DELETE requests for media items
func apiDeleteMedia(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	mID := types.MediaID(q)
	if err := mID.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	ac := middleware.AuditContextFromRequest(r, c)
	err := d.DeleteMedia(r.Context(), ac, mID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

// apiListMediaPaginated handles GET requests for listing media with pagination.
func apiListMediaPaginated(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	params := ParsePaginationParams(r)

	items, err := d.ListMediaPaginated(params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	total, err := d.CountMedia()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	response := db.PaginatedResponse[db.Media]{
		Data:   *items,
		Total:  *total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}

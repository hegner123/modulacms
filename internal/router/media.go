package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// MediasHandler handles CRUD operations that do not require a specific media ID.
func MediasHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			apiListMediaPaginated(w, r, svc)
		} else {
			apiListMedia(w, r, svc)
		}
	case http.MethodPost:
		apiCreateMedia(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// MediaHandler handles CRUD operations for specific media items.
func MediaHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetMedia(w, r, svc)
	case http.MethodPut:
		apiUpdateMedia(w, r, svc)
	case http.MethodDelete:
		apiDeleteMedia(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// MediaFullHandler handles requests for the media list with author names.
func MediaFullHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiListMediaFull(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func apiListMediaFull(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	views, err := svc.Media.ListMediaFull(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(views)
}

func apiGetMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	mID := types.MediaID(q)
	if err := mID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	record, err := svc.Media.GetMedia(r.Context(), mID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(record)
}

func apiListMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	mediaList, err := svc.Media.ListMedia(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(mediaList)
}

// apiCreateMedia handles POST requests to upload and create a new media item.
func apiCreateMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	if parseErr := r.ParseMultipartForm(c.MaxUploadSize()); parseErr != nil {
		utility.DefaultLogger.Error("parse form", parseErr)
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

	ac := middleware.AuditContextFromRequest(r, *c)

	row, err := svc.Media.Upload(r.Context(), ac, service.UploadMediaParams{
		File:   file,
		Header: header,
		Path:   r.PostFormValue("path"),
	})
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(row)
}

func apiUpdateMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var params service.UpdateMediaMetadataParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	updated, err := svc.Media.UpdateMediaMetadata(r.Context(), ac, params)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

// apiDeleteMedia handles DELETE requests for media items.
// Retains clean_refs logic in handler: calls svc.Driver() for reference scan
// before calling svc.Media.DeleteMedia.
func apiDeleteMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	mID := types.MediaID(q)
	if err := mID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	// If clean_refs=true, blank out content fields referencing this media.
	if r.URL.Query().Get("clean_refs") == "true" {
		record, getErr := svc.Media.GetMedia(r.Context(), mID)
		if getErr == nil {
			cleaned, cleanErr := cleanMediaReferences(svc.Driver(), r, *c, mID, string(record.URL))
			if cleanErr != nil {
				utility.DefaultLogger.Error("clean refs failed", cleanErr)
			} else if cleaned > 0 {
				utility.DefaultLogger.Info(fmt.Sprintf("cleaned %d media references for %s", cleaned, mID))
			}
		}
	}

	if err := svc.Media.DeleteMedia(r.Context(), ac, mID); err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// MediaHealthHandler checks for orphaned files in the media S3 bucket.
func MediaHealthHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	result, err := svc.Media.MediaHealth(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// MediaCleanupHandler deletes orphaned files from the media S3 bucket.
func MediaCleanupHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	result, err := svc.Media.MediaCleanup(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// MediaReferenceInfo describes a content field that references a media asset.
type MediaReferenceInfo struct {
	ContentFieldID types.ContentFieldID    `json:"content_field_id"`
	ContentDataID  types.NullableContentID `json:"content_data_id"`
	FieldID        types.NullableFieldID   `json:"field_id"`
}

// MediaReferenceScanResponse is the JSON response for GET /api/v1/media/references.
type MediaReferenceScanResponse struct {
	MediaID        types.MediaID        `json:"media_id"`
	References     []MediaReferenceInfo `json:"references"`
	ReferenceCount int                  `json:"reference_count"`
}

// MediaReferencesHandler handles GET /api/v1/media/references?q=<media_id>.
// Uses svc.Driver() for the reference scan (unmigrated helper pattern).
func MediaReferencesHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	q := r.URL.Query().Get("q")
	mID := types.MediaID(q)
	if err := mID.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid media_id: %v", err), http.StatusBadRequest)
		return
	}

	mediaRecord, err := d.GetMedia(mID)
	if err != nil {
		http.Error(w, fmt.Sprintf("media not found: %v", err), http.StatusNotFound)
		return
	}

	searchTerms := []string{string(mID)}
	if string(mediaRecord.URL) != "" {
		searchTerms = append(searchTerms, string(mediaRecord.URL))
	}

	allFields, err := d.ListContentFields()
	if err != nil {
		utility.DefaultLogger.Error("media references: failed to list content fields", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var refs []MediaReferenceInfo
	if allFields != nil {
		for _, cf := range *allFields {
			if cf.FieldValue == "" {
				continue
			}
			for _, term := range searchTerms {
				if strings.Contains(cf.FieldValue, term) {
					refs = append(refs, MediaReferenceInfo{
						ContentFieldID: cf.ContentFieldID,
						ContentDataID:  cf.ContentDataID,
						FieldID:        cf.FieldID,
					})
					break
				}
			}
		}
	}

	if refs == nil {
		refs = make([]MediaReferenceInfo, 0)
	}

	writeJSON(w, MediaReferenceScanResponse{
		MediaID:        mID,
		References:     refs,
		ReferenceCount: len(refs),
	})
}

// cleanMediaReferences blanks content field values that reference the given media.
// Stays as a handler-level concern per the service layer design.
func cleanMediaReferences(d db.DbDriver, r *http.Request, c config.Config, mID types.MediaID, mediaURL string) (int, error) {
	searchTerms := []string{string(mID)}
	if mediaURL != "" {
		searchTerms = append(searchTerms, mediaURL)
	}

	allFields, err := d.ListContentFields()
	if err != nil {
		return 0, fmt.Errorf("list content fields: %w", err)
	}

	cleaned := 0
	if allFields != nil {
		ctx := r.Context()
		ac := middleware.AuditContextFromRequest(r, c)
		now := types.TimestampNow()

		for _, cf := range *allFields {
			if cf.FieldValue == "" {
				continue
			}
			for _, term := range searchTerms {
				if strings.Contains(cf.FieldValue, term) {
					_, uErr := d.UpdateContentField(ctx, ac, db.UpdateContentFieldParams{
						ContentFieldID: cf.ContentFieldID,
						RouteID:        cf.RouteID,
						ContentDataID:  cf.ContentDataID,
						FieldID:        cf.FieldID,
						FieldValue:     "",
						Locale:         cf.Locale,
						AuthorID:       cf.AuthorID,
						DateCreated:    cf.DateCreated,
						DateModified:   now,
					})
					if uErr != nil {
						utility.DefaultLogger.Error(fmt.Sprintf("clean ref: failed to clear field %s", cf.ContentFieldID), uErr)
					} else {
						cleaned++
					}
					break
				}
			}
		}
	}

	return cleaned, nil
}

func apiListMediaPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	params := ParsePaginationParams(r)

	items, total, err := svc.Media.ListMediaPaginated(r.Context(), params.Limit, params.Offset)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
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
}

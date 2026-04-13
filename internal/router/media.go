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
	json.NewEncoder(w).Encode(toMediaResponse(*record))
}

func apiListMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	// Support folder_id filter: "unfiled" returns media with no folder,
	// a valid ULID returns media in that folder, absent returns all media.
	folderFilter := r.URL.Query().Get("folder_id")
	if folderFilter != "" {
		d := svc.Driver()
		if folderFilter == "unfiled" {
			items, err := d.ListMediaUnfiled()
			if err != nil {
				utility.DefaultLogger.Error("list unfiled media", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			mediaItems := make([]db.Media, 0)
			if items != nil {
				mediaItems = *items
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(toMediaListResponse(mediaItems))
			return
		}

		fid := types.MediaFolderID(folderFilter)
		if err := fid.Validate(); err != nil {
			http.Error(w, fmt.Sprintf("invalid folder_id: %v", err), http.StatusBadRequest)
			return
		}
		folderNullable := types.NullableMediaFolderID{ID: fid, Valid: true}
		items, err := d.ListMediaByFolder(folderNullable)
		if err != nil {
			utility.DefaultLogger.Error("list media by folder", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		mediaItems := make([]db.Media, 0)
		if items != nil {
			mediaItems = *items
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toMediaListResponse(mediaItems))
		return
	}

	mediaList, err := svc.Media.ListMedia(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toMediaListResponse(*mediaList))
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

	// If folder_id was provided, move the newly uploaded media into that folder.
	if folderIDStr := r.PostFormValue("folder_id"); folderIDStr != "" {
		fid := types.MediaFolderID(folderIDStr)
		if err := fid.Validate(); err == nil {
			moveErr := svc.Driver().MoveMediaToFolder(r.Context(), ac, db.MoveMediaToFolderParams{
				FolderID:     types.NullableMediaFolderID{ID: fid, Valid: true},
				DateModified: row.DateModified,
				MediaID:      row.MediaID,
			})
			if moveErr != nil {
				utility.DefaultLogger.Error("move uploaded media to folder", moveErr)
				// Upload succeeded, folder move failed — still return the created media
			} else {
				// Re-fetch to reflect updated folder_id
				updated, getErr := svc.Media.GetMedia(r.Context(), row.MediaID)
				if getErr == nil {
					row = updated
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toMediaResponse(*row))
}

func apiUpdateMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	// Decode into a struct that includes folder_id alongside metadata fields.
	var req struct {
		service.UpdateMediaMetadataParams
		FolderID *string `json:"folder_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	updated, err := svc.Media.UpdateMediaMetadata(r.Context(), ac, req.UpdateMediaMetadataParams)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	// If folder_id was provided, move the media to the specified folder.
	if req.FolderID != nil {
		var folderNullable types.NullableMediaFolderID
		if *req.FolderID != "" {
			fid := types.MediaFolderID(*req.FolderID)
			if err := fid.Validate(); err != nil {
				http.Error(w, fmt.Sprintf("invalid folder_id: %v", err), http.StatusBadRequest)
				return
			}
			folderNullable = types.NullableMediaFolderID{ID: fid, Valid: true}
		}
		// Empty string means move to root (unfiled)

		moveErr := svc.Driver().MoveMediaToFolder(r.Context(), ac, db.MoveMediaToFolderParams{
			FolderID:     folderNullable,
			DateModified: updated.DateModified,
			MediaID:      updated.MediaID,
		})
		if moveErr != nil {
			utility.DefaultLogger.Error("move media to folder during update", moveErr)
			http.Error(w, fmt.Sprintf("metadata updated but folder move failed: %v", moveErr), http.StatusInternalServerError)
			return
		}

		// Re-fetch to reflect updated folder_id
		refreshed, getErr := svc.Media.GetMedia(r.Context(), updated.MediaID)
		if getErr == nil {
			updated = refreshed
		}
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

// MediaReprocessStatusHandler returns the current bulk reprocess job status.
func MediaReprocessStatusHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	status := svc.Media.GetReprocessStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// MediaReprocessTriggerHandler manually triggers a bulk reprocess of all media variants.
func MediaReprocessTriggerHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	started := svc.Media.TriggerReprocess()

	w.Header().Set("Content-Type", "application/json")
	if started {
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]any{"reprocess_started": true, "message": "Bulk reprocess started"})
	} else {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"reprocess_started": false, "message": "Reprocess already running, restart queued"})
	}
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

	// Support folder_id filter with pagination.
	folderFilter := r.URL.Query().Get("folder_id")
	if folderFilter != "" {
		d := svc.Driver()
		if folderFilter == "unfiled" {
			items, err := d.ListMediaUnfiledPaginated(db.PaginationParams{Limit: params.Limit, Offset: params.Offset})
			if err != nil {
				utility.DefaultLogger.Error("list unfiled media paginated", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			total, err := d.CountMediaUnfiled()
			if err != nil {
				utility.DefaultLogger.Error("count unfiled media", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			mediaItems := make([]db.Media, 0)
			if items != nil {
				mediaItems = *items
			}
			response := db.PaginatedResponse[MediaResponse]{
				Data:   toMediaListResponse(mediaItems),
				Total:  *total,
				Limit:  params.Limit,
				Offset: params.Offset,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		fid := types.MediaFolderID(folderFilter)
		if err := fid.Validate(); err != nil {
			http.Error(w, fmt.Sprintf("invalid folder_id: %v", err), http.StatusBadRequest)
			return
		}
		folderNullable := types.NullableMediaFolderID{ID: fid, Valid: true}
		items, err := d.ListMediaByFolderPaginated(db.ListMediaByFolderPaginatedParams{
			FolderID: folderNullable,
			Limit:    params.Limit,
			Offset:   params.Offset,
		})
		if err != nil {
			utility.DefaultLogger.Error("list media by folder paginated", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		total, err := d.CountMediaByFolder(folderNullable)
		if err != nil {
			utility.DefaultLogger.Error("count media by folder", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		mediaItems := make([]db.Media, 0)
		if items != nil {
			mediaItems = *items
		}
		response := db.PaginatedResponse[MediaResponse]{
			Data:   toMediaListResponse(mediaItems),
			Total:  *total,
			Limit:  params.Limit,
			Offset: params.Offset,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	items, total, err := svc.Media.ListMediaPaginated(r.Context(), params.Limit, params.Offset)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	response := db.PaginatedResponse[MediaResponse]{
		Data:   toMediaListResponse(*items),
		Total:  *total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

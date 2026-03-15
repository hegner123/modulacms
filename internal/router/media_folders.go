package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// --- Request/Response types ---

// createMediaFolderRequest is the JSON body for POST /api/v1/media-folders.
type createMediaFolderRequest struct {
	Name     string `json:"name"`
	ParentID string `json:"parent_id"`
}

// updateMediaFolderRequest is the JSON body for PUT /api/v1/media-folders/{id}.
type updateMediaFolderRequest struct {
	Name     *string `json:"name"`
	ParentID *string `json:"parent_id"`
}

// batchMoveMediaRequest is the JSON body for POST /api/v1/media/move.
type batchMoveMediaRequest struct {
	MediaIDs []string `json:"media_ids"`
	FolderID *string  `json:"folder_id"`
}

// batchMoveMediaResponse is the JSON response for POST /api/v1/media/move.
type batchMoveMediaResponse struct {
	Moved int `json:"moved"`
}

// mediaFolderTreeNode represents a folder in the tree response.
type mediaFolderTreeNode struct {
	FolderID     types.MediaFolderID         `json:"folder_id"`
	Name         string                      `json:"name"`
	ParentID     types.NullableMediaFolderID `json:"parent_id"`
	DateCreated  types.Timestamp             `json:"date_created"`
	DateModified types.Timestamp             `json:"date_modified"`
	Children     []mediaFolderTreeNode       `json:"children"`
}

// --- Handlers ---

// apiListMediaFolders handles GET /api/v1/media-folders.
// Returns root folders by default, or children of a given parent_id.
func apiListMediaFolders(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	parentIDStr := r.URL.Query().Get("parent_id")
	if parentIDStr != "" {
		parentID := types.MediaFolderID(parentIDStr)
		if err := parentID.Validate(); err != nil {
			http.Error(w, fmt.Sprintf("invalid parent_id: %v", err), http.StatusBadRequest)
			return
		}

		folders, err := d.ListMediaFoldersByParent(parentID)
		if err != nil {
			utility.DefaultLogger.Error("list media folders by parent", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(folders)
		return
	}

	folders, err := d.ListMediaFoldersAtRoot()
	if err != nil {
		utility.DefaultLogger.Error("list media folders at root", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(folders)
}

// apiGetMediaFolder handles GET /api/v1/media-folders/{id}.
func apiGetMediaFolder(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	id := types.MediaFolderID(r.PathValue("id"))
	if err := id.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid folder id: %v", err), http.StatusBadRequest)
		return
	}

	folder, err := d.GetMediaFolder(id)
	if err != nil {
		utility.DefaultLogger.Error("get media folder", err)
		http.Error(w, "media folder not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(folder)
}

// apiCreateMediaFolder handles POST /api/v1/media-folders.
func apiCreateMediaFolder(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	var req createMediaFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	var parentID types.NullableMediaFolderID
	if req.ParentID != "" {
		pid := types.MediaFolderID(req.ParentID)
		if err := pid.Validate(); err != nil {
			http.Error(w, fmt.Sprintf("invalid parent_id: %v", err), http.StatusBadRequest)
			return
		}
		// Verify parent exists
		if _, err := d.GetMediaFolder(pid); err != nil {
			http.Error(w, "parent folder not found", http.StatusBadRequest)
			return
		}
		parentID = types.NullableMediaFolderID{ID: pid, Valid: true}

		// Validate depth won't exceed 10
		breadcrumb, err := d.GetMediaFolderBreadcrumb(pid)
		if err != nil {
			utility.DefaultLogger.Error("check folder depth", err)
			http.Error(w, "failed to validate folder depth", http.StatusInternalServerError)
			return
		}
		// breadcrumb includes from root to parent; new folder adds 1 more level
		if len(breadcrumb)+1 > 10 {
			http.Error(w, "creating this folder would exceed maximum folder depth of 10", http.StatusBadRequest)
			return
		}
	}

	// Validate name uniqueness within parent
	if err := d.ValidateMediaFolderName(name, parentID); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	c, err := svc.Config()
	if err != nil {
		utility.DefaultLogger.Error("load config", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)
	now := types.NewTimestamp(time.Now().UTC())

	folder, err := d.CreateMediaFolder(r.Context(), ac, db.CreateMediaFolderParams{
		Name:         name,
		ParentID:     parentID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		utility.DefaultLogger.Error("create media folder", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(folder)
}

// apiUpdateMediaFolder handles PUT /api/v1/media-folders/{id}.
func apiUpdateMediaFolder(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	id := types.MediaFolderID(r.PathValue("id"))
	if err := id.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid folder id: %v", err), http.StatusBadRequest)
		return
	}

	existing, err := d.GetMediaFolder(id)
	if err != nil {
		http.Error(w, "media folder not found", http.StatusNotFound)
		return
	}

	var req updateMediaFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	name := existing.Name
	parentID := existing.ParentID

	// Handle name change
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			http.Error(w, "name cannot be empty", http.StatusBadRequest)
			return
		}
		name = trimmed
	}

	// Handle parent change
	parentChanged := false
	if req.ParentID != nil {
		parentChanged = true
		if *req.ParentID == "" {
			// Move to root
			parentID = types.NullableMediaFolderID{}
		} else {
			pid := types.MediaFolderID(*req.ParentID)
			if err := pid.Validate(); err != nil {
				http.Error(w, fmt.Sprintf("invalid parent_id: %v", err), http.StatusBadRequest)
				return
			}
			parentID = types.NullableMediaFolderID{ID: pid, Valid: true}
		}
	}

	// Validate move if parent changed
	if parentChanged {
		if err := d.ValidateMediaFolderMove(id, parentID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Validate name uniqueness (check if name or parent changed)
	nameChanged := req.Name != nil && name != existing.Name
	if nameChanged || parentChanged {
		if err := d.ValidateMediaFolderName(name, parentID); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
	}

	c, err := svc.Config()
	if err != nil {
		utility.DefaultLogger.Error("load config", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	_, err = d.UpdateMediaFolder(r.Context(), ac, db.UpdateMediaFolderParams{
		FolderID:     id,
		Name:         name,
		ParentID:     parentID,
		DateModified: types.NewTimestamp(time.Now().UTC()),
	})
	if err != nil {
		utility.DefaultLogger.Error("update media folder", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updated, err := d.GetMediaFolder(id)
	if err != nil {
		utility.DefaultLogger.Error("fetch updated media folder", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

// apiDeleteMediaFolder handles DELETE /api/v1/media-folders/{id}.
// Rejects deletion if the folder has child folders or media items.
func apiDeleteMediaFolder(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	id := types.MediaFolderID(r.PathValue("id"))
	if err := id.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid folder id: %v", err), http.StatusBadRequest)
		return
	}

	// Verify folder exists
	if _, err := d.GetMediaFolder(id); err != nil {
		http.Error(w, "media folder not found", http.StatusNotFound)
		return
	}

	// Check for child folders
	children, err := d.ListMediaFoldersByParent(id)
	if err != nil {
		utility.DefaultLogger.Error("list child folders", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	childCount := 0
	if children != nil {
		childCount = len(*children)
	}

	// Check for media items in folder
	folderNullable := types.NullableMediaFolderID{ID: id, Valid: true}
	mediaCount, err := d.CountMediaByFolder(folderNullable)
	if err != nil {
		utility.DefaultLogger.Error("count media in folder", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if childCount > 0 || (mediaCount != nil && *mediaCount > 0) {
		mc := int64(0)
		if mediaCount != nil {
			mc = *mediaCount
		}
		msg := fmt.Sprintf("cannot delete folder: contains %d child folder(s) and %d media item(s)", childCount, mc)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]any{
			"error":         msg,
			"child_folders": childCount,
			"media_items":   mc,
		})
		return
	}

	c, err := svc.Config()
	if err != nil {
		utility.DefaultLogger.Error("load config", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	if err := d.DeleteMediaFolder(r.Context(), ac, id); err != nil {
		utility.DefaultLogger.Error("delete media folder", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// apiMediaFolderMedia handles GET /api/v1/media-folders/{id}/media.
// Returns paginated media items within a folder.
func apiMediaFolderMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	id := types.MediaFolderID(r.PathValue("id"))
	if err := id.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid folder id: %v", err), http.StatusBadRequest)
		return
	}

	// Verify folder exists
	if _, err := d.GetMediaFolder(id); err != nil {
		http.Error(w, "media folder not found", http.StatusNotFound)
		return
	}

	folderNullable := types.NullableMediaFolderID{ID: id, Valid: true}

	if HasPaginationParams(r) {
		params := ParsePaginationParams(r)

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
}

// apiMediaFolderTree handles GET /api/v1/media-folders/tree.
// Returns the full folder hierarchy as a nested tree.
func apiMediaFolderTree(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	allFolders, err := d.ListMediaFolders()
	if err != nil {
		utility.DefaultLogger.Error("list all media folders", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tree := buildFolderTree(allFolders)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tree)
}

// apiBatchMoveMedia handles POST /api/v1/media/move.
// Moves multiple media items to a folder (or to root if folder_id is null/empty).
func apiBatchMoveMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	var req batchMoveMediaRequest
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
	var folderID types.NullableMediaFolderID
	if req.FolderID != nil && *req.FolderID != "" {
		fid := types.MediaFolderID(*req.FolderID)
		if err := fid.Validate(); err != nil {
			http.Error(w, fmt.Sprintf("invalid folder_id: %v", err), http.StatusBadRequest)
			return
		}
		// Verify folder exists
		if _, err := d.GetMediaFolder(fid); err != nil {
			http.Error(w, "target folder not found", http.StatusNotFound)
			return
		}
		folderID = types.NullableMediaFolderID{ID: fid, Valid: true}
	}

	// Validate all media IDs upfront
	mediaIDs := make([]types.MediaID, 0, len(req.MediaIDs))
	for _, idStr := range req.MediaIDs {
		mid := types.MediaID(idStr)
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
		err := d.MoveMediaToFolder(r.Context(), ac, db.MoveMediaToFolderParams{
			FolderID:     folderID,
			DateModified: now,
			MediaID:      mid,
		})
		if err != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("move media %s to folder", mid), err)
			continue
		}
		moved++
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(batchMoveMediaResponse{Moved: moved})
}

// --- Tree building ---

// buildFolderTree assembles a flat list of folders into a nested tree structure.
// Walks from root folders downward, recursively attaching children.
func buildFolderTree(folders *[]db.MediaFolder) []mediaFolderTreeNode {
	if folders == nil || len(*folders) == 0 {
		return []mediaFolderTreeNode{}
	}

	// Index by ID for O(1) lookup
	folderByID := make(map[types.MediaFolderID]db.MediaFolder, len(*folders))
	for _, f := range *folders {
		folderByID[f.FolderID] = f
	}

	// Group children by parent ID
	childrenOf := make(map[types.MediaFolderID][]types.MediaFolderID)
	var rootIDs []types.MediaFolderID

	for _, f := range *folders {
		if !f.ParentID.Valid {
			rootIDs = append(rootIDs, f.FolderID)
		} else {
			pid := types.MediaFolderID(f.ParentID.ID)
			childrenOf[pid] = append(childrenOf[pid], f.FolderID)
		}
	}

	var buildNode func(id types.MediaFolderID) mediaFolderTreeNode
	buildNode = func(id types.MediaFolderID) mediaFolderTreeNode {
		f := folderByID[id]
		node := mediaFolderTreeNode{
			FolderID:     f.FolderID,
			Name:         f.Name,
			ParentID:     f.ParentID,
			DateCreated:  f.DateCreated,
			DateModified: f.DateModified,
			Children:     make([]mediaFolderTreeNode, 0, len(childrenOf[id])),
		}
		for _, childID := range childrenOf[id] {
			node.Children = append(node.Children, buildNode(childID))
		}
		return node
	}

	roots := make([]mediaFolderTreeNode, 0, len(rootIDs))
	for _, rid := range rootIDs {
		roots = append(roots, buildNode(rid))
	}
	return roots
}

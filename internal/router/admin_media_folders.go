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

// createAdminMediaFolderRequest is the JSON body for POST /api/v1/adminmedia-folders.
type createAdminMediaFolderRequest struct {
	Name     string `json:"name"`
	ParentID string `json:"parent_id"`
}

// updateAdminMediaFolderRequest is the JSON body for PUT /api/v1/adminmedia-folders/{id}.
type updateAdminMediaFolderRequest struct {
	Name     *string `json:"name"`
	ParentID *string `json:"parent_id"`
}

// adminMediaFolderTreeNode represents a folder in the tree response.
type adminMediaFolderTreeNode struct {
	FolderID     types.AdminMediaFolderID         `json:"folder_id"`
	Name         string                           `json:"name"`
	ParentID     types.NullableAdminMediaFolderID `json:"parent_id"`
	DateCreated  types.Timestamp                  `json:"date_created"`
	DateModified types.Timestamp                  `json:"date_modified"`
	Children     []adminMediaFolderTreeNode       `json:"children"`
}

// --- Handlers ---

// apiListAdminMediaFolders handles GET /api/v1/adminmedia-folders.
// Returns root folders by default, or children of a given parent_id.
func apiListAdminMediaFolders(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	parentIDStr := r.URL.Query().Get("parent_id")
	if parentIDStr != "" {
		parentID := types.AdminMediaFolderID(parentIDStr)
		if err := parentID.Validate(); err != nil {
			http.Error(w, fmt.Sprintf("invalid parent_id: %v", err), http.StatusBadRequest)
			return
		}

		folders, err := d.ListAdminMediaFoldersByParent(parentID)
		if err != nil {
			utility.DefaultLogger.Error("list admin media folders by parent", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(folders)
		return
	}

	folders, err := d.ListAdminMediaFoldersAtRoot()
	if err != nil {
		utility.DefaultLogger.Error("list admin media folders at root", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(folders)
}

// apiGetAdminMediaFolder handles GET /api/v1/adminmedia-folders/{id}.
func apiGetAdminMediaFolder(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	id := types.AdminMediaFolderID(r.PathValue("id"))
	if err := id.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid folder id: %v", err), http.StatusBadRequest)
		return
	}

	folder, err := d.GetAdminMediaFolder(id)
	if err != nil {
		utility.DefaultLogger.Error("get admin media folder", err)
		http.Error(w, "admin media folder not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(folder)
}

// apiCreateAdminMediaFolder handles POST /api/v1/adminmedia-folders.
func apiCreateAdminMediaFolder(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	var req createAdminMediaFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	var parentID types.NullableAdminMediaFolderID
	if req.ParentID != "" {
		pid := types.AdminMediaFolderID(req.ParentID)
		if err := pid.Validate(); err != nil {
			http.Error(w, fmt.Sprintf("invalid parent_id: %v", err), http.StatusBadRequest)
			return
		}
		// Verify parent exists
		if _, err := d.GetAdminMediaFolder(pid); err != nil {
			http.Error(w, "parent folder not found", http.StatusBadRequest)
			return
		}
		parentID = types.NullableAdminMediaFolderID{ID: pid, Valid: true}

		// Validate depth won't exceed 10
		breadcrumb, err := d.GetAdminMediaFolderBreadcrumb(pid)
		if err != nil {
			utility.DefaultLogger.Error("check admin folder depth", err)
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
	if err := d.ValidateAdminMediaFolderName(name, parentID); err != nil {
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

	folder, err := d.CreateAdminMediaFolder(r.Context(), ac, db.CreateAdminMediaFolderParams{
		Name:         name,
		ParentID:     parentID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		utility.DefaultLogger.Error("create admin media folder", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(folder)
}

// apiUpdateAdminMediaFolder handles PUT /api/v1/adminmedia-folders/{id}.
func apiUpdateAdminMediaFolder(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	id := types.AdminMediaFolderID(r.PathValue("id"))
	if err := id.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid folder id: %v", err), http.StatusBadRequest)
		return
	}

	existing, err := d.GetAdminMediaFolder(id)
	if err != nil {
		http.Error(w, "admin media folder not found", http.StatusNotFound)
		return
	}

	var req updateAdminMediaFolderRequest
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
			parentID = types.NullableAdminMediaFolderID{}
		} else {
			pid := types.AdminMediaFolderID(*req.ParentID)
			if err := pid.Validate(); err != nil {
				http.Error(w, fmt.Sprintf("invalid parent_id: %v", err), http.StatusBadRequest)
				return
			}
			parentID = types.NullableAdminMediaFolderID{ID: pid, Valid: true}
		}
	}

	// Validate move if parent changed
	if parentChanged {
		if err := d.ValidateAdminMediaFolderMove(id, parentID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Validate name uniqueness (check if name or parent changed)
	nameChanged := req.Name != nil && name != existing.Name
	if nameChanged || parentChanged {
		if err := d.ValidateAdminMediaFolderName(name, parentID); err != nil {
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

	_, err = d.UpdateAdminMediaFolder(r.Context(), ac, db.UpdateAdminMediaFolderParams{
		AdminFolderID: id,
		Name:          name,
		ParentID:      parentID,
		DateModified:  types.NewTimestamp(time.Now().UTC()),
	})
	if err != nil {
		utility.DefaultLogger.Error("update admin media folder", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updated, err := d.GetAdminMediaFolder(id)
	if err != nil {
		utility.DefaultLogger.Error("fetch updated admin media folder", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

// apiDeleteAdminMediaFolder handles DELETE /api/v1/adminmedia-folders/{id}.
// Rejects deletion if the folder has child folders or media items.
func apiDeleteAdminMediaFolder(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	id := types.AdminMediaFolderID(r.PathValue("id"))
	if err := id.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid folder id: %v", err), http.StatusBadRequest)
		return
	}

	// Verify folder exists
	if _, err := d.GetAdminMediaFolder(id); err != nil {
		http.Error(w, "admin media folder not found", http.StatusNotFound)
		return
	}

	// Check for child folders
	children, err := d.ListAdminMediaFoldersByParent(id)
	if err != nil {
		utility.DefaultLogger.Error("list admin child folders", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	childCount := 0
	if children != nil {
		childCount = len(*children)
	}

	// Check for media items in folder
	folderNullable := types.NullableAdminMediaFolderID{ID: id, Valid: true}
	mediaCount, err := d.CountAdminMediaByFolder(folderNullable)
	if err != nil {
		utility.DefaultLogger.Error("count admin media in folder", err)
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

	if err := d.DeleteAdminMediaFolder(r.Context(), ac, id); err != nil {
		utility.DefaultLogger.Error("delete admin media folder", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// apiAdminMediaFolderMedia handles GET /api/v1/adminmedia-folders/{id}/media.
// Returns paginated admin media items within a folder.
func apiAdminMediaFolderMedia(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	id := types.AdminMediaFolderID(r.PathValue("id"))
	if err := id.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("invalid folder id: %v", err), http.StatusBadRequest)
		return
	}

	// Verify folder exists
	if _, err := d.GetAdminMediaFolder(id); err != nil {
		http.Error(w, "admin media folder not found", http.StatusNotFound)
		return
	}

	folderNullable := types.NullableAdminMediaFolderID{ID: id, Valid: true}

	if HasPaginationParams(r) {
		params := ParsePaginationParams(r)

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
}

// apiAdminMediaFolderTree handles GET /api/v1/adminmedia-folders/tree.
// Returns the full folder hierarchy as a nested tree.
func apiAdminMediaFolderTree(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	d := svc.Driver()

	allFolders, err := d.ListAdminMediaFolders()
	if err != nil {
		utility.DefaultLogger.Error("list all admin media folders", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tree := buildAdminFolderTree(allFolders)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tree)
}

// --- Tree building ---

// buildAdminFolderTree assembles a flat list of admin media folders into a nested tree.
func buildAdminFolderTree(folders *[]db.AdminMediaFolder) []adminMediaFolderTreeNode {
	if folders == nil || len(*folders) == 0 {
		return []adminMediaFolderTreeNode{}
	}

	// Index by ID for O(1) lookup
	folderByID := make(map[types.AdminMediaFolderID]db.AdminMediaFolder, len(*folders))
	for _, f := range *folders {
		folderByID[f.AdminFolderID] = f
	}

	// Group children by parent ID
	childrenOf := make(map[types.AdminMediaFolderID][]types.AdminMediaFolderID)
	var rootIDs []types.AdminMediaFolderID

	for _, f := range *folders {
		if !f.ParentID.Valid {
			rootIDs = append(rootIDs, f.AdminFolderID)
		} else {
			pid := types.AdminMediaFolderID(f.ParentID.ID)
			childrenOf[pid] = append(childrenOf[pid], f.AdminFolderID)
		}
	}

	var buildNode func(id types.AdminMediaFolderID) adminMediaFolderTreeNode
	buildNode = func(id types.AdminMediaFolderID) adminMediaFolderTreeNode {
		f := folderByID[id]
		node := adminMediaFolderTreeNode{
			FolderID:     f.AdminFolderID,
			Name:         f.Name,
			ParentID:     f.ParentID,
			DateCreated:  f.DateCreated,
			DateModified: f.DateModified,
			Children:     make([]adminMediaFolderTreeNode, 0, len(childrenOf[id])),
		}
		for _, childID := range childrenOf[id] {
			node.Children = append(node.Children, buildNode(childID))
		}
		return node
	}

	roots := make([]adminMediaFolderTreeNode, 0, len(rootIDs))
	for _, rid := range rootIDs {
		roots = append(roots, buildNode(rid))
	}
	return roots
}

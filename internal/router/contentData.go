package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/tree/ops"
)

// ContentDatasHandler handles CRUD operations that do not require a specific data ID.
func ContentDatasHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			apiListContentDataPaginated(w, r, svc)
		} else {
			apiListContentData(w, r, svc)
		}
	case http.MethodPost:
		apiCreateContentData(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ContentDataHandler handles CRUD operations for specific content data items.
func ContentDataHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetContentData(w, r, svc)
	case http.MethodPost:
		apiCreateContentData(w, r, svc)
	case http.MethodPut:
		apiUpdateContentData(w, r, svc)
	case http.MethodDelete:
		apiDeleteContentData(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ContentDataFullHandler handles requests for the composed content data view.
func ContentDataFullHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetContentDataFull(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetContentDataFull handles GET requests for a content data item with author, datatype, and fields.
func apiGetContentDataFull(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	cdID := types.ContentID(q)
	if err := cdID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	view, err := svc.Content.GetFull(r.Context(), cdID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(view)
}

// apiListContentData handles GET requests for listing content data
func apiListContentData(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	contentDataList, err := svc.Content.List(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, contentDataList)
}

// apiCreateContentData handles POST requests to create new content data
func apiCreateContentData(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var newContentData db.CreateContentDataParams
	if err := json.NewDecoder(r.Body).Decode(&newContentData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	created, err := svc.Content.Create(r.Context(), ac, newContentData)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// updateContentDataRequest wraps UpdateContentDataParams with an
// optional revision field for optimistic locking.
type updateContentDataRequest struct {
	db.UpdateContentDataParams
	Revision int64 `json:"revision"`
}

// apiUpdateContentData handles PUT requests to update existing content data
func apiUpdateContentData(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var req updateContentDataRequest
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

	updated, err := svc.Content.Update(r.Context(), ac, req.UpdateContentDataParams, req.Revision)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, updated)
}

// apiDeleteContentData handles DELETE requests for content data.
// When recursive=true, collects and deletes all descendants first (leaves first).
func apiDeleteContentData(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	cdID := types.ContentID(q)
	if err := cdID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	recursive := r.URL.Query().Get("recursive") == "true"

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	deletedIDs, err := svc.Content.Delete(r.Context(), ac, cdID, recursive)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	if recursive {
		writeJSON(w, RecursiveDeleteResponse{
			DeletedRoot:  cdID,
			TotalDeleted: len(deletedIDs),
			DeletedIDs:   deletedIDs,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// apiListContentDataPaginated handles GET requests for listing content data with pagination.
func apiListContentDataPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	params := ParsePaginationParams(r)

	result, err := svc.Content.ListPaginated(r.Context(), params)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, result)
}

// MoveContentDataRequest is the JSON body for POST /api/v1/admincontentdatas/move.
type MoveContentDataRequest struct {
	NodeID      types.ContentID         `json:"node_id"`
	NewParentID types.NullableContentID `json:"new_parent_id"`
	Position    int                     `json:"position"`
}

// MoveContentDataResponse is the JSON response for POST /api/v1/admincontentdatas/move.
type MoveContentDataResponse struct {
	NodeID      types.ContentID         `json:"node_id"`
	OldParentID types.NullableContentID `json:"old_parent_id"`
	NewParentID types.NullableContentID `json:"new_parent_id"`
	Position    int                     `json:"position"`
}

// ContentDataMoveHandler handles POST requests to move content data to a new parent.
func ContentDataMoveHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	apiMoveContentData(w, r, svc)
}

// apiMoveContentData moves a content data node to a new parent at a given position.
func apiMoveContentData(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req MoveContentDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := req.NodeID.Validate(); err != nil {
		http.Error(w, "invalid node_id", http.StatusBadRequest)
		return
	}
	if req.NewParentID.Valid {
		if err := req.NewParentID.ID.Validate(); err != nil {
			http.Error(w, "invalid new_parent_id", http.StatusBadRequest)
			return
		}
	}
	if req.Position < 0 {
		http.Error(w, "position must be >= 0", http.StatusBadRequest)
		return
	}

	c, cfgErr := svc.Config()
	if cfgErr != nil {
		service.HandleServiceError(w, r, cfgErr)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	moveParams := ops.MoveParams[types.ContentID]{
		NodeID:      req.NodeID,
		NewParentID: contentIDToOpsNullable(req.NewParentID),
		Position:    req.Position,
	}

	result, err := svc.Content.Move(r.Context(), ac, moveParams)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, MoveContentDataResponse{
		NodeID:      req.NodeID,
		OldParentID: opsNullableToContentID(result.OldParentID),
		NewParentID: req.NewParentID,
		Position:    req.Position,
	})
}

// ReorderContentDataRequest is the JSON body for POST /api/v1/admincontentdatas/reorder.
type ReorderContentDataRequest struct {
	ParentID   types.NullableContentID `json:"parent_id"`
	OrderedIDs []types.ContentID       `json:"ordered_ids"`
}

// ReorderContentDataResponse is the JSON response for POST /api/v1/admincontentdatas/reorder.
type ReorderContentDataResponse struct {
	Updated  int                     `json:"updated"`
	ParentID types.NullableContentID `json:"parent_id"`
}

// ContentDataReorderHandler handles POST requests to reorder content data siblings.
func ContentDataReorderHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	apiReorderContentData(w, r, svc)
}

// apiReorderContentData atomically reorders sibling content data nodes under a parent.
func apiReorderContentData(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req ReorderContentDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if len(req.OrderedIDs) == 0 {
		http.Error(w, "ordered_ids must not be empty", http.StatusBadRequest)
		return
	}

	seen := make(map[string]struct{}, len(req.OrderedIDs))
	for _, id := range req.OrderedIDs {
		if err := id.Validate(); err != nil {
			http.Error(w, "invalid content_data_id", http.StatusBadRequest)
			return
		}
		s := string(id)
		if _, exists := seen[s]; exists {
			http.Error(w, "duplicate id", http.StatusBadRequest)
			return
		}
		seen[s] = struct{}{}
	}

	c, cfgErr := svc.Config()
	if cfgErr != nil {
		service.HandleServiceError(w, r, cfgErr)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	result, err := svc.Content.Reorder(r.Context(), ac, contentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderContentDataResponse{
		Updated:  result.Updated,
		ParentID: req.ParentID,
	})
}

// opsNullableToContentID converts ops.NullableID to types.NullableContentID for router use.
func opsNullableToContentID(n ops.NullableID[types.ContentID]) types.NullableContentID {
	if !n.Valid {
		return types.NullableContentID{}
	}
	return types.NullableContentID{ID: n.Value, Valid: true}
}

// ContentDataByRouteHandler handles requests for content under a specific route.
func ContentDataByRouteHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiListContentDataByRoute(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
func apiListContentDataByRoute(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	routeID := types.RouteID(q)
	if err := routeID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	views, err := svc.Content.ListByRoute(r.Context(), routeID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(views)
}

// apiGetContentData handles GET requests for a single content data
func apiGetContentData(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	cdID := types.ContentID(q)
	if err := cdID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cd, err := svc.Content.Get(r.Context(), cdID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, cd)
}

// RecursiveDeleteResponse is the JSON response for DELETE with recursive=true.
type RecursiveDeleteResponse struct {
	DeletedRoot  types.ContentID   `json:"deleted_root"`
	TotalDeleted int               `json:"total_deleted"`
	DeletedIDs   []types.ContentID `json:"deleted_ids"`
}

// contentIDToOpsNullable converts types.NullableContentID to ops.NullableID for router use.
func contentIDToOpsNullable(n types.NullableContentID) ops.NullableID[types.ContentID] {
	if !n.Valid {
		return ops.EmptyID[types.ContentID]()
	}
	return ops.NullID(n.ID)
}

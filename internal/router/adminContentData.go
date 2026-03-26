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

// AdminContentDatasHandler handles CRUD operations that do not require a specific data ID.
func AdminContentDatasHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		if HasPaginationParams(r) {
			apiListAdminContentDataPaginated(w, r, svc)
		} else {
			apiListAdminContentData(w, r, svc)
		}
	case http.MethodPost:
		apiCreateAdminContentData(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminContentDataHandler handles CRUD operations for specific admin content data items.
func AdminContentDataHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodPost:
		apiCreateAdminContentData(w, r, svc)
	case http.MethodPut:
		apiUpdateAdminContentData(w, r, svc)
	case http.MethodDelete:
		apiDeleteAdminContentData(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminContentDataFullHandler handles requests for the composed admin content data view.
func AdminContentDataFullHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetAdminContentDataFull(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func apiGetAdminContentDataFull(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	cdID := types.AdminContentID(q)
	if err := cdID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	view, err := svc.AdminContent.GetFull(r.Context(), cdID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(view)
}

// apiListAdminContentData handles GET requests for listing admin content data
func apiListAdminContentData(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	list, err := svc.AdminContent.List(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	writeJSON(w, list)
}

// apiCreateAdminContentData handles POST requests to create new admin content data
func apiCreateAdminContentData(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var params db.CreateAdminContentDataParams
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

	created, err := svc.AdminContent.Create(r.Context(), ac, params)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// updateAdminContentDataRequest wraps UpdateAdminContentDataParams with an
// optional revision field for optimistic locking.
type updateAdminContentDataRequest struct {
	db.UpdateAdminContentDataParams
	Revision int64 `json:"revision"`
}

// apiUpdateAdminContentData handles PUT requests to update existing admin content data
func apiUpdateAdminContentData(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var req updateAdminContentDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.DateModified = types.TimestampNow()

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	updated, err := svc.AdminContent.Update(r.Context(), ac, req.UpdateAdminContentDataParams, req.Revision)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, updated)
}

// apiDeleteAdminContentData handles DELETE requests for admin content data
func apiDeleteAdminContentData(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	acdID := types.AdminContentID(q)
	if err := acdID.Validate(); err != nil {
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

	deletedIDs, err := svc.AdminContent.Delete(r.Context(), ac, acdID, recursive)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	if recursive {
		writeJSON(w, map[string]any{
			"deleted_root":  acdID,
			"total_deleted": len(deletedIDs),
			"deleted_ids":   deletedIDs,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// apiListAdminContentDataPaginated handles GET requests for listing admin content data with pagination
func apiListAdminContentDataPaginated(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	params := ParsePaginationParams(r)

	result, err := svc.AdminContent.ListPaginated(r.Context(), params)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, result)
}

// MoveAdminContentDataRequest is the JSON body for POST /api/v1/admincontentdatas/move.
type MoveAdminContentDataRequest struct {
	NodeID      types.AdminContentID         `json:"node_id"`
	NewParentID types.NullableAdminContentID `json:"new_parent_id"`
	Position    int                          `json:"position"`
}

// MoveAdminContentDataResponse is the JSON response for POST /api/v1/admincontentdatas/move.
type MoveAdminContentDataResponse struct {
	NodeID      types.AdminContentID         `json:"node_id"`
	OldParentID types.NullableAdminContentID `json:"old_parent_id"`
	NewParentID types.NullableAdminContentID `json:"new_parent_id"`
	Position    int                          `json:"position"`
}

// AdminContentDataMoveHandler handles POST requests to move admin content data to a new parent.
func AdminContentDataMoveHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	apiMoveAdminContentData(w, r, svc)
}

// apiMoveAdminContentData moves an admin content data node to a new parent at a given position.
func apiMoveAdminContentData(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req MoveAdminContentDataRequest
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

	moveParams := ops.MoveParams[types.AdminContentID]{
		NodeID:      req.NodeID,
		NewParentID: adminContentIDToOpsNullable(req.NewParentID),
		Position:    req.Position,
	}

	result, err := svc.AdminContent.Move(r.Context(), ac, moveParams)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, MoveAdminContentDataResponse{
		NodeID:      req.NodeID,
		OldParentID: opsNullableToAdminContentID(result.OldParentID),
		NewParentID: req.NewParentID,
		Position:    req.Position,
	})
}

// ReorderAdminContentDataRequest is the JSON body for POST /api/v1/admincontentdatas/reorder.
type ReorderAdminContentDataRequest struct {
	ParentID   types.NullableAdminContentID `json:"parent_id"`
	OrderedIDs []types.AdminContentID       `json:"ordered_ids"`
}

// ReorderAdminContentDataResponse is the JSON response for POST /api/v1/admincontentdatas/reorder.
type ReorderAdminContentDataResponse struct {
	Updated  int                          `json:"updated"`
	ParentID types.NullableAdminContentID `json:"parent_id"`
}

// AdminContentDataReorderHandler handles POST requests to reorder admin content data siblings.
func AdminContentDataReorderHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	apiReorderAdminContentData(w, r, svc)
}

// apiReorderAdminContentData atomically reorders sibling admin content data nodes under a parent.
func apiReorderAdminContentData(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req ReorderAdminContentDataRequest
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
			http.Error(w, "invalid admin_content_data_id", http.StatusBadRequest)
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

	result, err := svc.AdminContent.Reorder(r.Context(), ac, adminContentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  result.Updated,
		ParentID: req.ParentID,
	})
}

// adminContentIDToOpsNullable converts types.NullableAdminContentID to ops.NullableID for router use.
func adminContentIDToOpsNullable(n types.NullableAdminContentID) ops.NullableID[types.AdminContentID] {
	if !n.Valid {
		return ops.EmptyID[types.AdminContentID]()
	}
	return ops.NullID(n.ID)
}

// opsNullableToAdminContentID converts ops.NullableID to types.NullableAdminContentID for router use.
func opsNullableToAdminContentID(n ops.NullableID[types.AdminContentID]) types.NullableAdminContentID {
	if !n.Valid {
		return types.NullableAdminContentID{}
	}
	return types.NullableAdminContentID{ID: n.Value, Valid: true}
}

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

	updated, err := svc.AdminContent.Reorder(r.Context(), ac, adminContentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  updated,
		ParentID: req.ParentID,
	})
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

	updated, err := svc.AdminContent.Reorder(r.Context(), ac, adminContentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  updated,
		ParentID: req.ParentID,
	})
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

	updated, err := svc.AdminContent.Reorder(r.Context(), ac, adminContentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  updated,
		ParentID: req.ParentID,
	})
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

	updated, err := svc.AdminContent.Reorder(r.Context(), ac, adminContentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  updated,
		ParentID: req.ParentID,
	})
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

	updated, err := svc.AdminContent.Reorder(r.Context(), ac, adminContentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  updated,
		ParentID: req.ParentID,
	})
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

	updated, err := svc.AdminContent.Reorder(r.Context(), ac, adminContentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  updated,
		ParentID: req.ParentID,
	})
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

	updated, err := svc.AdminContent.Reorder(r.Context(), ac, adminContentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  updated,
		ParentID: req.ParentID,
	})
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

	updated, err := svc.AdminContent.Reorder(r.Context(), ac, adminContentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  updated,
		ParentID: req.ParentID,
	})
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

	updated, err := svc.AdminContent.Reorder(r.Context(), ac, adminContentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  updated,
		ParentID: req.ParentID,
	})
}

// updateAdminContentDataRequest wraps UpdateAdminContentDataParams with an optional
// revision field for optimistic locking. If Revision is zero (omitted by the
// client), the update falls through to the non-revision path for backward
// compatibility.
type updateAdminContentDataRequest struct {
	db.UpdateAdminContentDataParams
	Revision int64 `json:"revision"`
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

	updated, err := svc.AdminContent.Reorder(r.Context(), ac, adminContentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  updated,
		ParentID: req.ParentID,
	})
}

// RecursiveDeleteResponse is the JSON response for DELETE with recursive=true.
type RecursiveDeleteResponse struct {
	DeletedRoot  types.AdminContentID   `json:"deleted_root"`
	TotalDeleted int                    `json:"total_deleted"`
	DeletedIDs   []types.AdminContentID `json:"deleted_ids"`
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

	updated, err := svc.AdminContent.Reorder(r.Context(), ac, adminContentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  updated,
		ParentID: req.ParentID,
	})
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

	updated, err := svc.AdminContent.Reorder(r.Context(), ac, adminContentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  updated,
		ParentID: req.ParentID,
	})
}

// MoveAdminContentDataRequest is the JSON body for POST /api/v1/contentdata/move.
type MoveAdminContentDataRequest struct {
	NodeID      types.AdminContentID         `json:"node_id"`
	NewParentID types.NullableAdminContentID `json:"new_parent_id"`
	Position    int                          `json:"position"`
}

// MoveAdminContentDataResponse is the JSON response for POST /api/v1/contentdata/move.
type MoveAdminContentDataResponse struct {
	NodeID      types.AdminContentID         `json:"node_id"`
	OldParentID types.NullableAdminContentID `json:"old_parent_id"`
	NewParentID types.NullableAdminContentID `json:"new_parent_id"`
	Position    int                          `json:"position"`
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

	updated, err := svc.AdminContent.Reorder(r.Context(), ac, adminContentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  updated,
		ParentID: req.ParentID,
	})
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

	updated, err := svc.AdminContent.Reorder(r.Context(), ac, adminContentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  updated,
		ParentID: req.ParentID,
	})
}

// ReorderAdminContentDataRequest is the JSON body for POST /api/v1/contentdata/reorder.
type ReorderAdminContentDataRequest struct {
	ParentID   types.NullableAdminContentID `json:"parent_id"`
	OrderedIDs []types.AdminContentID       `json:"ordered_ids"`
}

// ReorderAdminContentDataResponse is the JSON response for POST /api/v1/contentdata/reorder.
type ReorderAdminContentDataResponse struct {
	Updated  int                          `json:"updated"`
	ParentID types.NullableAdminContentID `json:"parent_id"`
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

	updated, err := svc.AdminContent.Reorder(r.Context(), ac, adminContentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  updated,
		ParentID: req.ParentID,
	})
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

	updated, err := svc.AdminContent.Reorder(r.Context(), ac, adminContentIDToOpsNullable(req.ParentID), req.OrderedIDs)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, ReorderAdminContentDataResponse{
		Updated:  updated,
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

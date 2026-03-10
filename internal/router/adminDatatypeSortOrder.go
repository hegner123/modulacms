package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// AdminDatatypeSortOrderHandler updates the sort order for a specific admin datatype.
// Registered as: PUT /api/v1/admindatatypes/{id}/sort-order
func AdminDatatypeSortOrderHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "admin datatype id is required", http.StatusBadRequest)
		return
	}

	adminDatatypeID := types.AdminDatatypeID(idStr)
	if err := adminDatatypeID.Validate(); err != nil {
		utility.DefaultLogger.Error("invalid admin datatype id", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		SortOrder int64 `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.DefaultLogger.Error("invalid request body", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ac, err := svc.AuditCtx(r.Context())
	if err != nil {
		utility.DefaultLogger.Error("failed to build audit context", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = svc.Schema.UpdateAdminDatatypeSortOrder(r.Context(), ac, db.UpdateAdminDatatypeSortOrderParams{
		AdminDatatypeID: adminDatatypeID,
		SortOrder:       req.SortOrder,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// AdminDatatypeMaxSortOrderHandler returns the maximum sort order for admin datatypes under a parent.
// When parent_id is omitted, returns the max for root-level admin datatypes.
// Registered as: GET /api/v1/admindatatypes/max-sort-order
func AdminDatatypeMaxSortOrderHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	parentIDStr := r.URL.Query().Get("parent_id")

	var parentID types.NullableAdminDatatypeID
	if parentIDStr != "" {
		pid := types.AdminDatatypeID(parentIDStr)
		if err := pid.Validate(); err != nil {
			utility.DefaultLogger.Error("invalid parent_id", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		parentID = types.NullableAdminDatatypeID{ID: pid, Valid: true}
	}

	maxSortOrder, err := svc.Schema.GetMaxAdminDatatypeSortOrder(r.Context(), parentID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int64{"max_sort_order": maxSortOrder})
}

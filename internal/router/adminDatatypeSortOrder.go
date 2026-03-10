package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

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

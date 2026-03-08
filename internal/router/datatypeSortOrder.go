package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// DatatypeSortOrderHandler updates the sort order for a specific datatype.
// Registered as: PUT /api/v1/datatype/{id}/sort-order
func DatatypeSortOrderHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "datatype id is required", http.StatusBadRequest)
		return
	}

	datatypeID := types.DatatypeID(idStr)
	if err := datatypeID.Validate(); err != nil {
		utility.DefaultLogger.Error("invalid datatype id", err)
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

	err = svc.Schema.UpdateDatatypeSortOrder(r.Context(), ac, db.UpdateDatatypeSortOrderParams{
		DatatypeID: datatypeID,
		SortOrder:  req.SortOrder,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// DatatypeMaxSortOrderHandler returns the maximum sort order for datatypes under a parent.
// When parent_id is omitted, returns the max for root-level datatypes.
// Registered as: GET /api/v1/datatype/max-sort-order
func DatatypeMaxSortOrderHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	parentIDStr := r.URL.Query().Get("parent_id")

	var parentID types.NullableDatatypeID
	if parentIDStr != "" {
		pid := types.DatatypeID(parentIDStr)
		if err := pid.Validate(); err != nil {
			utility.DefaultLogger.Error("invalid parent_id", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		parentID = types.NullableDatatypeID{ID: pid, Valid: true}
	}

	maxSortOrder, err := svc.Schema.GetMaxDatatypeSortOrder(r.Context(), parentID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int64{"max_sort_order": maxSortOrder})
}

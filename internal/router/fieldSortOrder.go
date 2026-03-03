package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// FieldSortOrderHandler updates the sort order for a specific field.
// Registered as: PUT /api/v1/fields/{id}/sort-order
func FieldSortOrderHandler(w http.ResponseWriter, r *http.Request, d db.DbDriver, c config.Config) {
	fieldIDStr := r.PathValue("id")
	if fieldIDStr == "" {
		http.Error(w, "field id is required", http.StatusBadRequest)
		return
	}

	fieldID := types.FieldID(fieldIDStr)
	if err := fieldID.Validate(); err != nil {
		utility.DefaultLogger.Error("invalid field id", err)
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

	ac := middleware.AuditContextFromRequest(r, c)
	err := d.UpdateFieldSortOrder(r.Context(), ac, db.UpdateFieldSortOrderParams{
		FieldID:   fieldID,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		utility.DefaultLogger.Error("failed to update field sort order", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// FieldMaxSortOrderHandler returns the maximum sort order for fields under a parent datatype.
// Registered as: GET /api/v1/fields/max-sort-order
func FieldMaxSortOrderHandler(w http.ResponseWriter, r *http.Request, d db.DbDriver) {
	parentIDStr := r.URL.Query().Get("parent_id")
	if parentIDStr == "" {
		http.Error(w, "parent_id query parameter is required", http.StatusBadRequest)
		return
	}

	parentID := types.DatatypeID(parentIDStr)
	if err := parentID.Validate(); err != nil {
		utility.DefaultLogger.Error("invalid parent_id", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nullableParentID := types.NullableDatatypeID{ID: parentID, Valid: true}
	maxSortOrder, err := d.GetMaxSortOrderByParentID(nullableParentID)
	if err != nil {
		utility.DefaultLogger.Error("failed to get max sort order", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int64{"max_sort_order": maxSortOrder})
}

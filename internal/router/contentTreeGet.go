package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// ContentTreeGetHandler returns the content tree for a route.
// Registered as: GET /api/v1/content/tree/{routeID}
func ContentTreeGetHandler(w http.ResponseWriter, r *http.Request, d db.DbDriver) {
	routeIDStr := r.PathValue("routeID")
	if routeIDStr == "" {
		http.Error(w, "routeID is required", http.StatusBadRequest)
		return
	}

	routeID := types.RouteID(routeIDStr)
	if err := routeID.Validate(); err != nil {
		utility.DefaultLogger.Error("invalid routeID", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nullableRouteID := types.NullableRouteID{ID: routeID, Valid: true}
	tree, err := d.GetContentTreeByRoute(nullableRouteID)
	if err != nil {
		utility.DefaultLogger.Error("failed to get content tree by route", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tree)
}

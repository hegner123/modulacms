package router

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// ContentTreeGetHandler returns the content tree for a route.
// Registered as: GET /api/v1/content/tree/{routeID}
func ContentTreeGetHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
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
	tree, err := svc.Content.GetTree(r.Context(), nullableRouteID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, tree)
}

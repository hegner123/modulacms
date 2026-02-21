package router

import (
	"encoding/json"
	"math"
	"net/http"
	"sort"
	"strconv"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// AdminRoutesHandler handles CRUD operations that do not require a specific admin route ID.
func AdminRoutesHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		if r.URL.Query().Get("ordered") == "true" {
			apiListOrderedAdminRoutes(w, r, c)
		} else if HasPaginationParams(r) {
			apiListAdminRoutesPaginated(w, r, c)
		} else {
			apiListAdminRoutes(w, c)
		}
	case http.MethodPost:
		apiCreateAdminRoute(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminRouteHandler handles CRUD operations for specific admin route items.
func AdminRouteHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetAdminRoute(w, r, c)
	case http.MethodPut:
		apiUpdateAdminRoute(w, r, c)
	case http.MethodDelete:
		apiDeleteAdminRoute(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetAdminRoute handles GET requests for a single admin route
func apiGetAdminRoute(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	slug := types.Slug(q)
	if err := slug.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	adminRoute, err := d.GetAdminRoute(slug)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminRoute)
	return nil
}

// apiListAdminRoutes handles GET requests for listing admin routes
func apiListAdminRoutes(w http.ResponseWriter, c config.Config) error {
	d := db.ConfigDB(c)

	adminRoutes, err := d.ListAdminRoutes()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminRoutes)
	return nil
}

// apiListOrderedAdminRoutes handles GET requests for listing admin routes sorted by the "Order" field
// on each route's root admin content node. Routes without an Order value appear last.
func apiListOrderedAdminRoutes(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	routes, err := d.ListAdminRoutes()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Find the "Order" admin field ID
	allFields, err := d.ListAdminFields()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	var orderFieldID types.AdminFieldID
	for _, f := range *allFields {
		if f.Label == "Order" {
			orderFieldID = f.AdminFieldID
			break
		}
	}

	// If no Order field exists, return routes unsorted
	if orderFieldID.IsZero() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(routes)
		return nil
	}

	type orderedRoute struct {
		Route db.AdminRoutes
		Order int
	}
	ordered := make([]orderedRoute, 0, len(*routes))

	for _, route := range *routes {
		nullableRouteID := types.NullableAdminRouteID{ID: route.AdminRouteID, Valid: true}
		order := math.MaxInt32

		contentData, err := d.ListAdminContentDataByRoute(nullableRouteID)
		if err != nil {
			utility.DefaultLogger.Error("failed to list admin content data for route", err)
			ordered = append(ordered, orderedRoute{Route: route, Order: order})
			continue
		}

		// Find root node (null parent)
		var rootContentDataID string
		for _, cd := range *contentData {
			if !cd.ParentID.Valid {
				rootContentDataID = cd.AdminContentDataID.String()
				break
			}
		}

		if rootContentDataID != "" {
			contentFields, err := d.ListAdminContentFieldsByRoute(nullableRouteID)
			if err != nil {
				utility.DefaultLogger.Error("failed to list admin content fields for route", err)
				ordered = append(ordered, orderedRoute{Route: route, Order: order})
				continue
			}

			for _, cf := range *contentFields {
				if cf.AdminContentDataID.String() == rootContentDataID &&
					cf.AdminFieldID.Valid &&
					cf.AdminFieldID.ID == orderFieldID {
					parsed, parseErr := strconv.Atoi(cf.AdminFieldValue)
					if parseErr == nil {
						order = parsed
					}
					break
				}
			}
		}

		ordered = append(ordered, orderedRoute{Route: route, Order: order})
	}

	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Order < ordered[j].Order
	})

	result := make([]db.AdminRoutes, len(ordered))
	for i, o := range ordered {
		result[i] = o.Route
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
	return nil
}

// apiCreateAdminRoute handles POST requests to create a new admin route
func apiCreateAdminRoute(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var newAdminRoute db.CreateAdminRouteParams
	err := json.NewDecoder(r.Body).Decode(&newAdminRoute)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	createdAdminRoute, err := d.CreateAdminRoute(r.Context(), ac, newAdminRoute)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdAdminRoute)
	return nil
}

// apiUpdateAdminRoute handles PUT requests to update an existing admin route
func apiUpdateAdminRoute(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var updateAdminRoute db.UpdateAdminRouteParams
	err := json.NewDecoder(r.Body).Decode(&updateAdminRoute)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	_, err = d.UpdateAdminRoute(r.Context(), ac, updateAdminRoute)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	updated, err := d.GetAdminRoute(updateAdminRoute.Slug)
	if err != nil {
		utility.DefaultLogger.Error("failed to fetch updated admin route", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteAdminRoute handles DELETE requests for admin routes
func apiDeleteAdminRoute(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	id := types.AdminRouteID(q)
	if err := id.Validate(); err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	ac := middleware.AuditContextFromRequest(r, c)
	err := d.DeleteAdminRoute(r.Context(), ac, id)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

// apiListAdminRoutesPaginated handles GET requests for listing admin routes with pagination
func apiListAdminRoutesPaginated(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)
	params := ParsePaginationParams(r)

	items, err := d.ListAdminRoutesPaginated(params)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	total, err := d.CountAdminRoutes()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	response := db.PaginatedResponse[db.AdminRoutes]{
		Data:   *items,
		Total:  *total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}

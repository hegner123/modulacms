package router

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// ImportContentfulHandler handles importing Contentful format to Modula
func ImportContentfulHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiImportContent(w, r, svc, config.FormatContentful)
}

// ImportSanityHandler handles importing Sanity format to Modula
func ImportSanityHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiImportContent(w, r, svc, config.FormatSanity)
}

// ImportStrapiHandler handles importing Strapi format to Modula
func ImportStrapiHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiImportContent(w, r, svc, config.FormatStrapi)
}

// ImportWordPressHandler handles importing WordPress format to Modula
func ImportWordPressHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiImportContent(w, r, svc, config.FormatWordPress)
}

// ImportCleanHandler handles importing Clean Modula format
func ImportCleanHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiImportContent(w, r, svc, config.FormatClean)
}

// ImportBulkHandler handles bulk import with format specified in request
func ImportBulkHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get format from query parameter
	format := r.URL.Query().Get("format")
	if format == "" {
		http.Error(w, "format query parameter required", http.StatusBadRequest)
		return
	}

	if !config.IsValidOutputFormat(format) {
		http.Error(w, "Invalid format. Valid options: contentful, sanity, strapi, wordpress, clean", http.StatusBadRequest)
		return
	}

	apiImportContent(w, r, svc, config.OutputFormat(format))
}

// apiImportContent handles the core import logic
func apiImportContent(w http.ResponseWriter, r *http.Request, svc *service.Registry, format config.OutputFormat) {
	cfg, cfgErr := svc.Config()
	if cfgErr != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utility.DefaultLogger.Error("failed to read request body", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	routeID := types.NullableRouteID{Valid: false}
	routeIDStr := r.URL.Query().Get("route_id")
	if routeIDStr != "" {
		routeID = types.NullableRouteID{
			ID:    types.RouteID(routeIDStr),
			Valid: true,
		}
	}

	ac := middleware.AuditContextFromRequest(r, *cfg)
	result, err := svc.Import.ImportContent(r.Context(), ac, service.ImportContentInput{
		Format:  format,
		Body:    body,
		RouteID: routeID,
	})
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

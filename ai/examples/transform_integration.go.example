package transform

// Complete integration example showing how to use transformers in HTTP handlers
// This file is for documentation and is not compiled

/*

COMPLETE EXAMPLE: HTTP Handlers with Transform Package

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/model"
	"github.com/hegner123/modulacms/internal/transform"
)

// ============================================================================
// EXAMPLE 1: GET Request (Read) - Transform ModulaCMS → CMS Format
// ============================================================================

func GetContentHandler(cfg *config.Config, driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract ID from URL
		id := r.URL.Path[len("/content/"):]

		// 2. Query database for content (simplified)
		root, err := queryContentByID(driver, id)
		if err != nil {
			http.Error(w, "Content not found", http.StatusNotFound)
			return
		}

		// 3. Allow format override via query parameter
		format := cfg.Output_Format
		if queryFormat := r.URL.Query().Get("format"); queryFormat != "" {
			if config.IsValidOutputFormat(queryFormat) {
				format = config.OutputFormat(queryFormat)
			} else {
				http.Error(w, "Invalid format", http.StatusBadRequest)
				return
			}
		}

		// 4. Create transform config
		transformCfg := transform.NewTransformConfig(
			format,
			cfg.Client_Site,
			cfg.Space_ID,
			driver,
		)

		// 5. Transform and write response
		if err := transformCfg.TransformAndWrite(w, root); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// ============================================================================
// EXAMPLE 2: POST Request (Create) - Parse CMS Format → ModulaCMS
// ============================================================================

func CreateContentHandler(cfg *config.Config, driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Determine input format
		inputFormat := cfg.Output_Format
		if headerFormat := r.Header.Get("Content-Type-Format"); headerFormat != "" {
			if config.IsValidOutputFormat(headerFormat) {
				inputFormat = config.OutputFormat(headerFormat)
			}
		}

		// 2. Create transform config
		transformCfg := transform.NewTransformConfig(
			inputFormat,
			cfg.Client_Site,
			cfg.Space_ID,
			driver,
		)

		// 3. Parse request body from CMS format → ModulaCMS
		root, err := transformCfg.ParseRequest(r)
		if err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// 4. Save to database (simplified)
		newID, err := saveContentToDatabase(driver, root)
		if err != nil {
			http.Error(w, "Failed to save content", http.StatusInternalServerError)
			return
		}

		// 5. Update root with new ID
		if root.Node != nil {
			root.Node.Datatype.Content.ContentDataID = newID
		}

		// 6. Return created content in same format
		w.WriteHeader(http.StatusCreated)
		if err := transformCfg.TransformAndWrite(w, root); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// ============================================================================
// EXAMPLE 3: PUT Request (Update) - Parse CMS Format → ModulaCMS
// ============================================================================

func UpdateContentHandler(cfg *config.Config, driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract ID from URL
		id := r.URL.Path[len("/content/"):]

		// 2. Check if content exists
		existing, err := queryContentByID(driver, id)
		if err != nil {
			http.Error(w, "Content not found", http.StatusNotFound)
			return
		}

		// 3. Determine input format
		inputFormat := cfg.Output_Format
		if headerFormat := r.Header.Get("Content-Type-Format"); headerFormat != "" {
			if config.IsValidOutputFormat(headerFormat) {
				inputFormat = config.OutputFormat(headerFormat)
			}
		}

		// 4. Create transform config
		transformCfg := transform.NewTransformConfig(
			inputFormat,
			cfg.Client_Site,
			cfg.Space_ID,
			driver,
		)

		// 5. Parse request body
		updated, err := transformCfg.ParseRequest(r)
		if err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// 6. Merge with existing (preserve ID, dates, etc.)
		if updated.Node != nil {
			updated.Node.Datatype.Content.ContentDataID = existing.Node.Datatype.Content.ContentDataID
			updated.Node.Datatype.Content.DateCreated = existing.Node.Datatype.Content.DateCreated
		}

		// 7. Update database
		if err := updateContentInDatabase(driver, updated); err != nil {
			http.Error(w, "Failed to update content", http.StatusInternalServerError)
			return
		}

		// 8. Return updated content
		if err := transformCfg.TransformAndWrite(w, updated); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// ============================================================================
// EXAMPLE 4: DELETE Request (Delete)
// ============================================================================

func DeleteContentHandler(cfg *config.Config, driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract ID from URL
		id := r.URL.Path[len("/content/"):]

		// 2. Delete from database
		if err := deleteContentFromDatabase(driver, id); err != nil {
			http.Error(w, "Failed to delete content", http.StatusInternalServerError)
			return
		}

		// 3. Return 204 No Content
		w.WriteHeader(http.StatusNoContent)
	}
}

// ============================================================================
// EXAMPLE 5: List Request (GET with query) - Multiple items
// ============================================================================

func ListContentHandler(cfg *config.Config, driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Parse query parameters
		contentType := r.URL.Query().Get("type")
		limit := getQueryInt(r, "limit", 10)
		offset := getQueryInt(r, "offset", 0)

		// 2. Query database for list
		items, err := queryContentList(driver, contentType, limit, offset)
		if err != nil {
			http.Error(w, "Failed to query content", http.StatusInternalServerError)
			return
		}

		// 3. Determine format
		format := cfg.Output_Format
		if queryFormat := r.URL.Query().Get("format"); queryFormat != "" {
			if config.IsValidOutputFormat(queryFormat) {
				format = config.OutputFormat(queryFormat)
			}
		}

		// 4. Create transform config
		transformCfg := transform.NewTransformConfig(
			format,
			cfg.Client_Site,
			cfg.Space_ID,
			driver,
		)

		// 5. Transform each item
		transformer, err := transformCfg.GetTransformer()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		transformedItems := make([]any, len(items))
		for i, item := range items {
			transformed, err := transformer.Transform(item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			transformedItems[i] = transformed
		}

		// 6. Wrap in list response
		response := map[string]any{
			"items":  transformedItems,
			"total":  len(transformedItems),
			"limit":  limit,
			"offset": offset,
		}

		// 7. Write response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// ============================================================================
// EXAMPLE 6: Using Config Constants
// ============================================================================

func SetupRoutes(cfg *config.Config, driver db.DbDriver) {
	// Public API - Always Contentful format
	http.HandleFunc("/api/v1/content/", func(w http.ResponseWriter, r *http.Request) {
		transformCfg := transform.NewTransformConfig(
			config.FormatContentful,  // Constant
			cfg.Client_Site,
			cfg.Space_ID,
			driver,
		)

		root := getContent(r)
		transformCfg.TransformAndWrite(w, root)
	})

	// Admin API - Always Clean format
	http.HandleFunc("/admin/api/content/", func(w http.ResponseWriter, r *http.Request) {
		transformCfg := transform.NewTransformConfig(
			config.FormatClean,  // Constant
			cfg.Admin_Site,
			cfg.Space_ID,
			driver,
		)

		root := getContent(r)
		transformCfg.TransformAndWrite(w, root)
	})

	// Debug API - Always Raw format
	http.HandleFunc("/debug/content/", func(w http.ResponseWriter, r *http.Request) {
		transformCfg := transform.NewTransformConfig(
			config.FormatRaw,  // Constant
			cfg.Client_Site,
			cfg.Space_ID,
			driver,
		)

		root := getContent(r)
		transformCfg.TransformAndWrite(w, root)
	})
}

// ============================================================================
// EXAMPLE 7: Accept Header Content Negotiation
// ============================================================================

func ContentNegotiationHandler(cfg *config.Config, driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		root := getContent(r)

		// Determine format from Accept header
		format := cfg.Output_Format // Default
		accept := r.Header.Get("Accept")

		switch accept {
		case "application/vnd.contentful+json":
			format = config.FormatContentful
		case "application/vnd.sanity+json":
			format = config.FormatSanity
		case "application/vnd.strapi+json":
			format = config.FormatStrapi
		case "application/vnd.wordpress+json":
			format = config.FormatWordPress
		case "application/vnd.modulacms.clean+json":
			format = config.FormatClean
		}

		transformCfg := transform.NewTransformConfig(
			format,
			cfg.Client_Site,
			cfg.Space_ID,
			driver,
		)

		transformCfg.TransformAndWrite(w, root)
	}
}

// ============================================================================
// Helper Functions (Simplified - implement with actual DB logic)
// ============================================================================

func queryContentByID(driver db.DbDriver, id string) (model.Root, error) {
	// TODO: Implement actual database query
	// This is a placeholder
	return model.Root{}, nil
}

func queryContentList(driver db.DbDriver, contentType string, limit, offset int) ([]model.Root, error) {
	// TODO: Implement actual database query
	return []model.Root{}, nil
}

func saveContentToDatabase(driver db.DbDriver, root model.Root) (int64, error) {
	// TODO: Implement actual database insert
	return 42, nil
}

func updateContentInDatabase(driver db.DbDriver, root model.Root) error {
	// TODO: Implement actual database update
	return nil
}

func deleteContentFromDatabase(driver db.DbDriver, id string) error {
	// TODO: Implement actual database delete
	return nil
}

func getContent(r *http.Request) model.Root {
	// TODO: Implement actual content retrieval
	return model.Root{}
}

func getQueryInt(r *http.Request, key string, defaultValue int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}
	// TODO: Parse int from string
	return defaultValue
}

*/

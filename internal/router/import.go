package router

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/model"
	"github.com/hegner123/modulacms/internal/transform"
	"github.com/hegner123/modulacms/internal/utility"
)

// ImportContentfulHandler handles importing Contentful format to ModulaCMS
func ImportContentfulHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiImportContent(w, r, c, config.FormatContentful)
}

// ImportSanityHandler handles importing Sanity format to ModulaCMS
func ImportSanityHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiImportContent(w, r, c, config.FormatSanity)
}

// ImportStrapiHandler handles importing Strapi format to ModulaCMS
func ImportStrapiHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiImportContent(w, r, c, config.FormatStrapi)
}

// ImportWordPressHandler handles importing WordPress format to ModulaCMS
func ImportWordPressHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiImportContent(w, r, c, config.FormatWordPress)
}

// ImportCleanHandler handles importing Clean ModulaCMS format
func ImportCleanHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiImportContent(w, r, c, config.FormatClean)
}

// ImportBulkHandler handles bulk import with format specified in request
func ImportBulkHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
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

	apiImportContent(w, r, c, config.OutputFormat(format))
}

// apiImportContent handles the core import logic
func apiImportContent(w http.ResponseWriter, r *http.Request, c config.Config, format config.OutputFormat) {
	d := db.ConfigDB(c)
	con, _, err := d.GetConnection()
	if err != nil {
		utility.DefaultLogger.Error("Database connection error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer con.Close()

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utility.DefaultLogger.Error("Failed to read request body", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Create transformer for the specified format
	transformCfg := transform.NewTransformConfig(
		format,
		c.Client_Site,
		c.Space_ID,
		d,
	)

	transformer, err := transformCfg.GetTransformer()
	if err != nil {
		utility.DefaultLogger.Error("Failed to get transformer", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse CMS format to ModulaCMS
	root, err := transformer.Parse(body)
	if err != nil {
		utility.DefaultLogger.Error("Failed to parse input", err)
		http.Error(w, "Failed to parse input: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Import to database (STUB - to be implemented)
	result, err := importRootToDatabase(d, root)
	if err != nil {
		utility.DefaultLogger.Error("Failed to import to database", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	Success         bool     `json:"success"`
	DatatypesCreated int     `json:"datatypes_created"`
	FieldsCreated    int     `json:"fields_created"`
	ContentCreated   int     `json:"content_created"`
	Errors          []string `json:"errors,omitempty"`
	Message         string   `json:"message"`
}

// importRootToDatabase imports a parsed Root structure to the database
// STUB: This is a placeholder for the actual implementation
func importRootToDatabase(driver db.DbDriver, root model.Root) (*ImportResult, error) {
	result := &ImportResult{
		Success: false,
		Message: "Import functionality not yet implemented (stub)",
	}

	if root.Node == nil {
		result.Errors = append(result.Errors, "No content to import")
		return result, nil
	}

	// TODO: Implement actual database insertion logic
	// Steps to implement:
	// 1. Extract or create Datatype from root.Node.Datatype
	// 2. Create/update Fields from root.Node.Fields
	// 3. Create ContentData from root.Node.Datatype.Content
	// 4. Create ContentFields linking content to fields
	// 5. Recursively process child nodes (root.Node.Nodes)
	// 6. Handle relationships and references

	// STUB: Simulate successful import
	result.Success = true
	result.Message = "Import stubbed - database insertion not implemented"
	result.DatatypesCreated = countDatatypes(root)
	result.FieldsCreated = countFields(root)
	result.ContentCreated = countContent(root)

	utility.DefaultLogger.Info("Import stub called",
		"datatypes", result.DatatypesCreated,
		"fields", result.FieldsCreated,
		"content", result.ContentCreated,
	)

	return result, nil
}

// Helper functions to count items in the tree (for stub response)

func countDatatypes(root model.Root) int {
	if root.Node == nil {
		return 0
	}
	count := 1 // Current node's datatype
	for _, child := range root.Node.Nodes {
		count += countDatatypes(model.Root{Node: child})
	}
	return count
}

func countFields(root model.Root) int {
	if root.Node == nil {
		return 0
	}
	count := len(root.Node.Fields)
	for _, child := range root.Node.Nodes {
		count += countFields(model.Root{Node: child})
	}
	return count
}

func countContent(root model.Root) int {
	if root.Node == nil {
		return 0
	}
	count := 1 // Current node's content
	for _, child := range root.Node.Nodes {
		count += countContent(model.Root{Node: child})
	}
	return count
}

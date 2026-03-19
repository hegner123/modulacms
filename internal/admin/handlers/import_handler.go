package handlers

import (
	"io"
	"net/http"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// ImportPageHandler renders the import wizard form.
// Shows a file upload form with format selector.
func ImportPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		layout := NewAdminData(r, "Import")
		RenderNav(w, r, "Import", pages.ImportContent(layout.CSRFToken), pages.Import(layout))
	}
}

// formatMap maps form values to config.OutputFormat.
var formatMap = map[string]config.OutputFormat{
	"wordpress":  config.FormatWordPress,
	"strapi":     config.FormatStrapi,
	"sanity":     config.FormatSanity,
	"contentful": config.FormatContentful,
	"clean":      config.FormatClean,
}

// ImportSubmitHandler processes an import file upload.
// Reads the uploaded file, delegates to the import service, and returns a result partial.
func ImportSubmitHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Limit upload to 32MB
		r.Body = http.MaxBytesReader(w, r.Body, 32<<20)

		if parseErr := r.ParseMultipartForm(32 << 20); parseErr != nil {
			utility.DefaultLogger.Error("failed to parse import form", parseErr)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "File too large or invalid form data", "type": "error"}}`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			http.Error(w, "File too large or invalid form data", http.StatusBadRequest)
			return
		}

		format := r.FormValue("format")
		outputFormat, ok := formatMap[format]
		if !ok || format == "" {
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Invalid import format", "type": "error"}}`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			http.Error(w, "Invalid import format", http.StatusBadRequest)
			return
		}

		file, _, fileErr := r.FormFile("import_file")
		if fileErr != nil {
			utility.DefaultLogger.Error("failed to read import file", fileErr)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "No file uploaded", "type": "error"}}`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			http.Error(w, "No file uploaded", http.StatusBadRequest)
			return
		}
		defer file.Close()

		body, readErr := io.ReadAll(file)
		if readErr != nil {
			utility.DefaultLogger.Error("failed to read import file body", readErr)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to read file", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		result, importErr := svc.Import.ImportContent(r.Context(), ac, service.ImportContentInput{
			Format: outputFormat,
			Body:   body,
			RouteID: types.NullableRouteID{
				Valid: false,
			},
		})
		if importErr != nil {
			utility.DefaultLogger.Error("import failed", importErr)
			failResult := &service.ImportResult{
				Success: false,
				Message: importErr.Error(),
			}
			Render(w, r, pages.ImportResultPartial(failResult))
			return
		}

		Render(w, r, pages.ImportResultPartial(result))
	}
}

package handlers

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/admin/pages"
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

// ImportSubmitHandler processes an import file upload.
// Reads the uploaded file and format selection, then delegates to the import logic.
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
		if format == "" {
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Import format is required", "type": "error"}}`)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			http.Error(w, "Import format is required", http.StatusBadRequest)
			return
		}

		file, header, fileErr := r.FormFile("import_file")
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

		_ = header // file header available for future use (filename, size, etc.)

		// TODO: Implement actual import processing based on format.
		// For now, acknowledge the upload and return success.

		if IsHTMX(r) {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Import received. Processing is not yet implemented.", "type": "info"}}`)
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/admin/import", http.StatusSeeOther)
	}
}

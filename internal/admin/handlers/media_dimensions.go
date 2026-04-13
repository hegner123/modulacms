package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// MediaDimensionsListHandler lists all media dimension presets.
func MediaDimensionsListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := svc.Media.ListMediaDimensions(r.Context())
		if err != nil {
			utility.DefaultLogger.Error("failed to list media dimensions", err)
			http.Error(w, "failed to load media dimensions", http.StatusInternalServerError)
			return
		}

		var dimensions []db.MediaDimensions
		if items != nil {
			dimensions = *items
		}

		status := svc.Media.GetReprocessStatus()

		if IsNavHTMX(r) {
			csrfToken := CSRFTokenFromContext(r.Context())
			w.Header().Set("HX-Trigger", `{"pageTitle": "Media Dimensions"}`)
			Render(w, r, pages.MediaDimensionsListContent(dimensions, csrfToken, status))
			return
		}

		if IsHTMX(r) {
			Render(w, r, partials.MediaDimensionsTableRows(dimensions))
			return
		}

		layout := NewAdminData(r, "Media Dimensions")
		Render(w, r, pages.MediaDimensionsList(layout, dimensions, status))
	}
}

// MediaDimensionReprocessStatusHandler returns the current reprocess status
// as an HTMX partial for polling.
func MediaDimensionReprocessStatusHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := svc.Media.GetReprocessStatus()
		Render(w, r, partials.ReprocessStatus(status))
	}
}

// MediaDimensionCreateHandler creates a new dimension preset.
func MediaDimensionCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		label := r.FormValue("label")
		widthStr := r.FormValue("width")
		heightStr := r.FormValue("height")
		aspectRatio := r.FormValue("aspect_ratio")

		width, err := strconv.ParseInt(widthStr, 10, 64)
		if err != nil || width <= 0 {
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Width must be a positive number", "type": "error"}}`)
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}
			http.Error(w, "Width must be a positive number", http.StatusBadRequest)
			return
		}

		height, err := strconv.ParseInt(heightStr, 10, 64)
		if err != nil || height <= 0 {
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Height must be a positive number", "type": "error"}}`)
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}
			http.Error(w, "Height must be a positive number", http.StatusBadRequest)
			return
		}

		input := db.CreateMediaDimensionParams{
			Label:       db.NullString{NullString: sql.NullString{String: label, Valid: label != ""}},
			Width:       types.NullableInt64{Int64: width, Valid: true},
			Height:      types.NullableInt64{Int64: height, Valid: true},
			AspectRatio: db.NullString{NullString: sql.NullString{String: aspectRatio, Valid: aspectRatio != ""}},
		}

		ac := middleware.AuditContextFromRequest(r, *cfg)
		_, createErr := svc.Media.CreateMediaDimension(r.Context(), ac, input)
		if createErr != nil {
			utility.DefaultLogger.Error("failed to create media dimension", createErr)
			if IsHTMX(r) {
				msg := fmt.Sprintf(`{"showToast": {"message": "failed to create dimension: %s", "type": "error"}}`, createErr.Error())
				w.Header().Set("HX-Trigger", msg)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.Error(w, "failed to create media dimension", http.StatusInternalServerError)
			return
		}

		svc.Media.TriggerReprocess()

		if IsHTMX(r) {
			items, listErr := svc.Media.ListMediaDimensions(r.Context())
			if listErr != nil {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Dimension created but failed to reload list", "type": "warning"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			var dimensions []db.MediaDimensions
			if items != nil {
				dimensions = *items
			}
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Dimension saved. Regenerating image variants in background...", "type": "success"}}`)
			Render(w, r, partials.MediaDimensionsTableRows(dimensions))
			return
		}
		http.Redirect(w, r, "/admin/media/dimensions", http.StatusSeeOther)
	}
}

// MediaDimensionUpdateHandler updates an existing dimension preset.
func MediaDimensionUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Dimension ID required", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		label := r.FormValue("label")
		widthStr := r.FormValue("width")
		heightStr := r.FormValue("height")
		aspectRatio := r.FormValue("aspect_ratio")

		width, err := strconv.ParseInt(widthStr, 10, 64)
		if err != nil || width <= 0 {
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Width must be a positive number", "type": "error"}}`)
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}
			http.Error(w, "Width must be a positive number", http.StatusBadRequest)
			return
		}

		height, err := strconv.ParseInt(heightStr, 10, 64)
		if err != nil || height <= 0 {
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Height must be a positive number", "type": "error"}}`)
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}
			http.Error(w, "Height must be a positive number", http.StatusBadRequest)
			return
		}

		input := db.UpdateMediaDimensionParams{
			MdID:        id,
			Label:       db.NullString{NullString: sql.NullString{String: label, Valid: label != ""}},
			Width:       types.NullableInt64{Int64: width, Valid: true},
			Height:      types.NullableInt64{Int64: height, Valid: true},
			AspectRatio: db.NullString{NullString: sql.NullString{String: aspectRatio, Valid: aspectRatio != ""}},
		}

		ac := middleware.AuditContextFromRequest(r, *cfg)
		_, updateErr := svc.Media.UpdateMediaDimension(r.Context(), ac, input)
		if updateErr != nil {
			utility.DefaultLogger.Error("failed to update media dimension", updateErr)
			if IsHTMX(r) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to update dimension", "type": "error"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.Error(w, "failed to update media dimension", http.StatusInternalServerError)
			return
		}

		svc.Media.TriggerReprocess()

		if IsHTMX(r) {
			items, listErr := svc.Media.ListMediaDimensions(r.Context())
			if listErr != nil {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Dimension updated but failed to reload list", "type": "warning"}}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			var dimensions []db.MediaDimensions
			if items != nil {
				dimensions = *items
			}
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Dimension saved. Regenerating image variants in background...", "type": "success"}}`)
			Render(w, r, partials.MediaDimensionsTableRows(dimensions))
			return
		}
		http.Redirect(w, r, "/admin/media/dimensions", http.StatusSeeOther)
	}
}

// MediaDimensionDeleteHandler deletes a dimension preset.
func MediaDimensionDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Dimension ID required", http.StatusBadRequest)
			return
		}

		ac := middleware.AuditContextFromRequest(r, *cfg)

		if err := svc.Media.DeleteMediaDimension(r.Context(), ac, id); err != nil {
			utility.DefaultLogger.Error("failed to delete media dimension", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to delete dimension", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		svc.Media.TriggerReprocess()

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Dimension deleted. Regenerating image variants in background...", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

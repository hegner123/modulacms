package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// RoutesListHandler handles GET /admin/routes.
// Lists all routes with pagination. HTMX requests receive partial table rows only.
func RoutesListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset := ParsePagination(r)

		items, total, err := svc.Routes.ListRoutesPaginated(r.Context(), limit, offset)
		if err != nil {
			utility.DefaultLogger.Error("failed to list routes", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		list := make([]db.Routes, 0)
		if items != nil {
			list = *items
		}

		pd := NewPaginationData(*total, limit, offset, "#routes-table-body", "/admin/routes")
		pg := partials.PaginationPageData{
			Current:    pd.Current,
			TotalPages: pd.TotalPages,
			Limit:      pd.Limit,
			Target:     pd.Target,
			BaseURL:    pd.BaseURL,
		}

		csrfToken := CSRFTokenFromContext(r.Context())

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Routes"}`)
			RenderWithOOB(w, r, pages.RoutesListContent(list, pg),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.RouteCreateDialog(csrfToken)})
			return
		}

		if IsHTMX(r) {
			Render(w, r, partials.RoutesTableRows(list, pg))
			return
		}

		layout := NewAdminData(r, "Routes")
		Render(w, r, pages.RoutesList(layout, list, pg))
	}
}

// RouteDetailHandler handles GET /admin/routes/{id}.
// Shows the full route detail with author info and content tree.
func RouteDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.NotFound(w, r)
			return
		}

		view, err := svc.Routes.GetRouteFull(r.Context(), types.RouteID(id))
		if err != nil {
			if service.IsNotFound(err) {
				http.NotFound(w, r)
				return
			}
			utility.DefaultLogger.Error("failed to get route", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		hasEdit := HasPermission(r, "routes:update")
		hasDelete := HasPermission(r, "routes:delete")

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Route: `+view.Title+`"}`)
			Render(w, r, pages.RouteDetailContent(view, csrfToken, hasEdit, hasDelete))
			return
		}

		layout := NewAdminData(r, "Route: "+view.Title)
		Render(w, r, pages.RouteDetail(layout, view, csrfToken, hasEdit, hasDelete))
	}
}

// RouteCreateHandler handles POST /admin/routes.
// Validates slug (must start with /) and title are required, creates via service layer.
func RouteCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		slug := strings.TrimSpace(r.FormValue("slug"))
		title := strings.TrimSpace(r.FormValue("title"))
		statusStr := strings.TrimSpace(r.FormValue("status"))

		var status int64
		if statusStr != "" {
			parsed, parseErr := strconv.ParseInt(statusStr, 10, 64)
			if parseErr == nil {
				status = parsed
			}
		}

		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}
		ac := middleware.AuditContextFromRequest(r, *c)

		user := middleware.AuthenticatedUser(r.Context())
		input := service.CreateRouteInput{
			Slug:     slug,
			Title:    title,
			Status:   status,
			AuthorID: types.NullableUserID{ID: user.UserID, Valid: true},
		}

		_, err := svc.Routes.CreateRoute(r.Context(), ac, input)
		if err != nil {
			if service.IsValidation(err) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				csrfToken := CSRFTokenFromContext(r.Context())
				errs := serviceValidationToMap(err)
				Render(w, r, partials.RouteForm(slug, title, statusStr, errs, csrfToken))
				return
			}
			if service.IsConflict(err) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				csrfToken := CSRFTokenFromContext(r.Context())
				Render(w, r, partials.RouteForm(slug, title, statusStr, map[string]string{"slug": "Slug already exists"}, csrfToken))
				return
			}
			utility.DefaultLogger.Error("failed to create route", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.RouteForm(slug, title, statusStr, map[string]string{"_": "failed to create route"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/routes", http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Route created", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/routes")
		w.WriteHeader(http.StatusOK)
	}
}

// RouteUpdateHandler handles POST /admin/routes/{id}.
// Updates an existing route via service layer.
func RouteUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing route ID", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		slug := strings.TrimSpace(r.FormValue("slug"))
		title := strings.TrimSpace(r.FormValue("title"))
		statusStr := strings.TrimSpace(r.FormValue("status"))

		var status int64
		if statusStr != "" {
			parsed, parseErr := strconv.ParseInt(statusStr, 10, 64)
			if parseErr == nil {
				status = parsed
			}
		}

		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}
		ac := middleware.AuditContextFromRequest(r, *c)

		user := middleware.AuthenticatedUser(r.Context())
		input := service.UpdateRouteInput{
			RouteID:  types.RouteID(id),
			Slug:     slug,
			Title:    title,
			Status:   status,
			AuthorID: types.NullableUserID{ID: user.UserID, Valid: true},
		}

		_, err := svc.Routes.UpdateRoute(r.Context(), ac, input)
		if err != nil {
			if service.IsValidation(err) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				csrfToken := CSRFTokenFromContext(r.Context())
				errs := serviceValidationToMap(err)
				Render(w, r, partials.RouteEditForm(id, slug, title, statusStr, errs, csrfToken))
				return
			}
			if service.IsConflict(err) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				csrfToken := CSRFTokenFromContext(r.Context())
				Render(w, r, partials.RouteEditForm(id, slug, title, statusStr, map[string]string{"slug": "Slug already exists"}, csrfToken))
				return
			}
			if service.IsNotFound(err) {
				http.Error(w, "Route not found", http.StatusNotFound)
				return
			}
			utility.DefaultLogger.Error("failed to update route", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.RouteEditForm(id, slug, title, statusStr, map[string]string{"_": "failed to update route"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/routes", http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Route updated", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/routes")
		w.WriteHeader(http.StatusOK)
	}
}

// RouteDeleteHandler handles DELETE /admin/routes/{id}.
// HTMX-only endpoint. Non-HTMX requests receive 405.
func RouteDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing route ID", http.StatusBadRequest)
			return
		}

		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}
		ac := middleware.AuditContextFromRequest(r, *c)

		err := svc.Routes.DeleteRoute(r.Context(), ac, types.RouteID(id))
		if err != nil {
			if service.IsConflict(err) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Route has content attached and cannot be deleted", "type": "error"}}`)
				w.WriteHeader(http.StatusConflict)
				return
			}
			utility.DefaultLogger.Error("failed to delete route", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to delete route", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Route deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// AdminRoutesListHandler handles GET /admin/routes/admin.
// Lists admin-specific routes (internal CMS routes). Read-only view.
func AdminRoutesListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset := ParsePagination(r)

		items, total, err := svc.Routes.ListAdminRoutesPaginated(r.Context(), limit, offset)
		if err != nil {
			utility.DefaultLogger.Error("failed to list admin routes", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		list := make([]db.AdminRoutes, 0)
		if items != nil {
			list = *items
		}

		pd2 := NewPaginationData(*total, limit, offset, "#admin-routes-table-body", "/admin/routes/admin")
		pg2 := partials.PaginationPageData{
			Current:    pd2.Current,
			TotalPages: pd2.TotalPages,
			Limit:      pd2.Limit,
			Target:     pd2.Target,
			BaseURL:    pd2.BaseURL,
		}

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Admin Routes"}`)
			Render(w, r, pages.AdminRoutesListContent(list, pg2))
			return
		}

		if IsHTMX(r) {
			Render(w, r, partials.AdminRoutesTableRows(list, pg2))
			return
		}

		layout := NewAdminData(r, "Admin Routes")
		Render(w, r, pages.AdminRoutesList(layout, list, pg2))
	}
}

// serviceValidationToMap converts a service.ValidationError to map[string]string
// for template rendering (first error per field wins).
func serviceValidationToMap(err error) map[string]string {
	var ve *service.ValidationError
	if !errors.As(err, &ve) {
		return nil
	}
	result := make(map[string]string, len(ve.Errors))
	for _, fe := range ve.Errors {
		if _, exists := result[fe.Field]; !exists {
			result[fe.Field] = fe.Message
		}
	}
	return result
}

package handlers

import (
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

// AdminSchemaRoutesListHandler handles GET /admin/admin-schema/routes.
// Lists all admin routes with pagination and create dialog OOB swap.
func AdminSchemaRoutesListHandler(svc *service.Registry) http.HandlerFunc {
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

		pd := NewPaginationData(*total, limit, offset, "#admin-routes-table-body", "/admin/admin-schema/routes")
		pg := partials.PaginationPageData{
			Current:    pd.Current,
			TotalPages: pd.TotalPages,
			Limit:      pd.Limit,
			Target:     pd.Target,
			BaseURL:    pd.BaseURL,
		}

		csrfToken := CSRFTokenFromContext(r.Context())

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Admin Routes"}`)
			RenderWithOOB(w, r, pages.AdminRoutesListContent(list, pg),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.AdminRouteCreateDialog(csrfToken)})
			return
		}

		if IsHTMX(r) {
			Render(w, r, partials.AdminSchemaRoutesTableRows(list, pg))
			return
		}

		layout := NewAdminData(r, "Admin Routes")
		Render(w, r, pages.AdminRoutesList(layout, list, pg))
	}
}

// AdminRouteCreateHandler handles POST /admin/admin-schema/routes.
// Validates slug (must start with /) and title are required, creates via service layer.
func AdminRouteCreateHandler(svc *service.Registry) http.HandlerFunc {
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
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
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

		_, err := svc.Routes.CreateAdminRoute(r.Context(), ac, input)
		if err != nil {
			if service.IsValidation(err) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				csrfToken := CSRFTokenFromContext(r.Context())
				errs := serviceValidationToMap(err)
				Render(w, r, partials.AdminRouteForm(slug, title, statusStr, errs, csrfToken))
				return
			}
			if service.IsConflict(err) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				csrfToken := CSRFTokenFromContext(r.Context())
				Render(w, r, partials.AdminRouteForm(slug, title, statusStr, map[string]string{"slug": "Slug already exists"}, csrfToken))
				return
			}
			utility.DefaultLogger.Error("failed to create admin route", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.AdminRouteForm(slug, title, statusStr, map[string]string{"_": "Failed to create route"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/admin-schema/routes", http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin route created", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/admin-schema/routes")
		w.WriteHeader(http.StatusOK)
	}
}

// AdminRouteUpdateHandler handles POST /admin/admin-schema/routes/{id}.
// Updates an existing admin route via the service layer.
func AdminRouteUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing route ID", http.StatusBadRequest)
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
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
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

		_, err := svc.Routes.UpdateAdminRoute(r.Context(), ac, input)
		if err != nil {
			if service.IsValidation(err) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				csrfToken := CSRFTokenFromContext(r.Context())
				errs := serviceValidationToMap(err)
				Render(w, r, partials.AdminRouteEditForm(id, slug, title, statusStr, errs, csrfToken))
				return
			}
			if service.IsConflict(err) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				csrfToken := CSRFTokenFromContext(r.Context())
				Render(w, r, partials.AdminRouteEditForm(id, slug, title, statusStr, map[string]string{"slug": "Slug already exists"}, csrfToken))
				return
			}
			if service.IsNotFound(err) {
				http.Error(w, "Admin route not found", http.StatusNotFound)
				return
			}
			utility.DefaultLogger.Error("failed to update admin route", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.AdminRouteEditForm(id, slug, title, statusStr, map[string]string{"_": "Failed to update route"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/admin-schema/routes", http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin route updated", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/admin-schema/routes")
		w.WriteHeader(http.StatusOK)
	}
}

// AdminRouteDeleteHandler handles DELETE /admin/admin-schema/routes/{id}.
// HTMX-only endpoint. Non-HTMX requests receive 405.
func AdminRouteDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing route ID", http.StatusBadRequest)
			return
		}

		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}
		ac := middleware.AuditContextFromRequest(r, *c)

		err := svc.Routes.DeleteAdminRoute(r.Context(), ac, types.AdminRouteID(id))
		if err != nil {
			if service.IsConflict(err) {
				w.Header().Set("HX-Trigger", `{"showToast": {"message": "Route has content attached and cannot be deleted", "type": "error"}}`)
				w.WriteHeader(http.StatusConflict)
				return
			}
			utility.DefaultLogger.Error("failed to delete admin route", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete route", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Admin route deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

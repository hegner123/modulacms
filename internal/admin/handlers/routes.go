package handlers

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// RoutesListHandler handles GET /admin/routes.
// Lists all routes with pagination. HTMX requests receive partial table rows only.
func RoutesListHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset := ParsePagination(r)

		items, err := driver.ListRoutesPaginated(db.PaginationParams{Limit: limit, Offset: offset})
		if err != nil {
			utility.DefaultLogger.Error("failed to list routes", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		total, err := driver.CountRoutes()
		if err != nil {
			utility.DefaultLogger.Error("failed to count routes", err)
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

// RouteCreateHandler handles POST /admin/routes.
// Validates slug (must start with /) and title are required, creates via audited context.
func RouteCreateHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		slug := strings.TrimSpace(r.FormValue("slug"))
		title := strings.TrimSpace(r.FormValue("title"))
		statusStr := strings.TrimSpace(r.FormValue("status"))

		errs := make(map[string]string)
		if slug == "" {
			errs["slug"] = "Slug is required"
		} else {
			if validateErr := types.Slug(slug).Validate(); validateErr != nil {
				errs["slug"] = validateErr.Error()
			}
		}
		if title == "" {
			errs["title"] = "Title is required"
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.RouteForm(slug, title, statusStr, errs, csrfToken))
			return
		}

		var status int64
		if statusStr != "" {
			parsed, parseErr := strconv.ParseInt(statusStr, 10, 64)
			if parseErr == nil {
				status = parsed
			}
		}

		user := middleware.AuthenticatedUser(r.Context())
		ip, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			ip = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID("0"), user.UserID, middleware.RequestIDFromContext(r.Context()), ip)

		now := types.NewTimestamp(time.Now())
		_, err := driver.CreateRoute(r.Context(), ac, db.CreateRouteParams{
			Slug:         types.Slug(slug),
			Title:        title,
			Status:       status,
			AuthorID:     types.NullableUserID{ID: user.UserID, Valid: true},
			DateCreated:  now,
			DateModified: now,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to create route", err)
			w.WriteHeader(http.StatusInternalServerError)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.RouteForm(slug, title, statusStr, map[string]string{"_": "Failed to create route"}, csrfToken))
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
// Updates an existing route via audited context.
func RouteUpdateHandler(driver db.DbDriver) http.HandlerFunc {
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

		errs := make(map[string]string)
		if slug == "" {
			errs["slug"] = "Slug is required"
		} else {
			if validateErr := types.Slug(slug).Validate(); validateErr != nil {
				errs["slug"] = validateErr.Error()
			}
		}
		if title == "" {
			errs["title"] = "Title is required"
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.RouteEditForm(id, slug, title, statusStr, errs, csrfToken))
			return
		}

		existing, err := driver.GetRoute(types.RouteID(id))
		if err != nil {
			http.Error(w, "Route not found", http.StatusNotFound)
			return
		}

		var status int64
		if statusStr != "" {
			parsed, parseErr := strconv.ParseInt(statusStr, 10, 64)
			if parseErr == nil {
				status = parsed
			}
		}

		user := middleware.AuthenticatedUser(r.Context())
		ip, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			ip = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID("0"), user.UserID, middleware.RequestIDFromContext(r.Context()), ip)

		_, err = driver.UpdateRoute(r.Context(), ac, db.UpdateRouteParams{
			Slug:         types.Slug(slug),
			Title:        title,
			Status:       status,
			AuthorID:     existing.AuthorID,
			DateCreated:  existing.DateCreated,
			DateModified: types.NewTimestamp(time.Now()),
			Slug_2:       existing.Slug,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to update route", err)
			w.WriteHeader(http.StatusInternalServerError)
			csrfToken := CSRFTokenFromContext(r.Context())
			Render(w, r, partials.RouteEditForm(id, slug, title, statusStr, map[string]string{"_": "Failed to update route"}, csrfToken))
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
func RouteDeleteHandler(driver db.DbDriver) http.HandlerFunc {
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

		user := middleware.AuthenticatedUser(r.Context())
		ip, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			ip = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID("0"), user.UserID, middleware.RequestIDFromContext(r.Context()), ip)

		err := driver.DeleteRoute(r.Context(), ac, types.RouteID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to delete route", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete route", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Route deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// AdminRoutesListHandler handles GET /admin/routes/admin.
// Lists admin-specific routes (internal CMS routes). Read-only view.
func AdminRoutesListHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset := ParsePagination(r)

		items, err := driver.ListAdminRoutesPaginated(db.PaginationParams{Limit: limit, Offset: offset})
		if err != nil {
			utility.DefaultLogger.Error("failed to list admin routes", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		total, err := driver.CountAdminRoutes()
		if err != nil {
			utility.DefaultLogger.Error("failed to count admin routes", err)
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

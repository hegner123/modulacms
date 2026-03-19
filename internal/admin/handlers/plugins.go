package handlers

import (
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// PluginsListHandler lists installed plugins with real data from the plugin system.
func PluginsListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		plugins, err := svc.Plugins.List(r.Context())
		if err != nil {
			utility.DefaultLogger.Error("failed to list plugins", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Plugins"}`)
			Render(w, r, pages.PluginsListContent(plugins))
			return
		}

		layout := NewAdminData(r, "Plugins")
		Render(w, r, pages.PluginsList(layout, plugins))
	}
}

// PluginDetailHandler shows plugin detail with real data from the plugin system.
func PluginDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			http.NotFound(w, r)
			return
		}

		detail, err := svc.Plugins.Get(r.Context(), name)
		if err != nil {
			if service.IsNotFound(err) {
				http.NotFound(w, r)
				return
			}
			utility.DefaultLogger.Error("failed to get plugin", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Fetch routes and hooks for this plugin.
		routes, routesErr := svc.Plugins.ListRoutes(r.Context())
		if routesErr != nil {
			utility.DefaultLogger.Error("failed to list plugin routes", routesErr)
			routes = nil
		}
		hooks, hooksErr := svc.Plugins.ListHooks(r.Context())
		if hooksErr != nil {
			utility.DefaultLogger.Error("failed to list plugin hooks", hooksErr)
			hooks = nil
		}

		// Filter to this plugin's routes and hooks.
		pluginRoutes := filterRoutesByPlugin(routes, name)
		pluginHooks := filterHooksByPlugin(hooks, name)

		csrfToken := CSRFTokenFromContext(r.Context())
		hasAdmin := HasPermission(r, "plugins:admin")

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Plugin: `+name+`"}`)
			Render(w, r, pages.PluginDetailContent(detail, pluginRoutes, pluginHooks, csrfToken, hasAdmin))
			return
		}

		layout := NewAdminData(r, "Plugin: "+name)
		Render(w, r, pages.PluginDetail(layout, detail, pluginRoutes, pluginHooks, csrfToken, hasAdmin))
	}
}

// PluginEnableHandler enables a plugin by name.
func PluginEnableHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			htmxError(w, r, "Plugin name is required", http.StatusBadRequest)
			return
		}

		if err := svc.Plugins.Enable(r.Context(), name); err != nil {
			htmxServiceError(w, r, "Failed to enable plugin", err)
			return
		}

		htmxRedirectToDetail(w, name, "Plugin enabled")
	}
}

// PluginDisableHandler disables a plugin by name.
func PluginDisableHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			htmxError(w, r, "Plugin name is required", http.StatusBadRequest)
			return
		}

		if err := svc.Plugins.Disable(r.Context(), name); err != nil {
			htmxServiceError(w, r, "Failed to disable plugin", err)
			return
		}

		htmxRedirectToDetail(w, name, "Plugin disabled")
	}
}

// PluginReloadHandler reloads a plugin by name.
func PluginReloadHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			htmxError(w, r, "Plugin name is required", http.StatusBadRequest)
			return
		}

		if err := svc.Plugins.Reload(r.Context(), name); err != nil {
			htmxServiceError(w, r, "Failed to reload plugin", err)
			return
		}

		htmxRedirectToDetail(w, name, "Plugin reloaded")
	}
}

// PluginApproveRouteHandler approves a single plugin route.
func PluginApproveRouteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			htmxError(w, r, "Invalid form data", http.StatusBadRequest)
			return
		}

		pluginName := r.FormValue("plugin")
		method := r.FormValue("method")
		path := r.FormValue("path")
		if pluginName == "" || method == "" || path == "" {
			htmxError(w, r, "Missing route fields", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		approvedBy := ""
		if user != nil {
			approvedBy = user.Username
		}

		err := svc.Plugins.ApproveRoutes(r.Context(), []service.RouteApprovalInput{
			{Plugin: pluginName, Method: method, Path: path},
		}, approvedBy)
		if err != nil {
			htmxServiceError(w, r, "Failed to approve route", err)
			return
		}

		renderPluginRoutes(w, r, svc, pluginName, "Route approved")
	}
}

// PluginRevokeRouteHandler revokes a single plugin route.
func PluginRevokeRouteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			htmxError(w, r, "Invalid form data", http.StatusBadRequest)
			return
		}

		pluginName := r.FormValue("plugin")
		method := r.FormValue("method")
		path := r.FormValue("path")
		if pluginName == "" || method == "" || path == "" {
			htmxError(w, r, "Missing route fields", http.StatusBadRequest)
			return
		}

		err := svc.Plugins.RevokeRoutes(r.Context(), []service.RouteApprovalInput{
			{Plugin: pluginName, Method: method, Path: path},
		})
		if err != nil {
			htmxServiceError(w, r, "Failed to revoke route", err)
			return
		}

		renderPluginRoutes(w, r, svc, pluginName, "Route revoked")
	}
}

// PluginApproveHookHandler approves a single plugin hook.
func PluginApproveHookHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			htmxError(w, r, "Invalid form data", http.StatusBadRequest)
			return
		}

		pluginName := r.FormValue("plugin")
		event := r.FormValue("event")
		table := r.FormValue("table")
		if pluginName == "" || event == "" {
			htmxError(w, r, "Missing hook fields", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		approvedBy := ""
		if user != nil {
			approvedBy = user.Username
		}

		err := svc.Plugins.ApproveHooks(r.Context(), []service.HookApprovalInput{
			{PluginName: pluginName, Event: event, Table: table},
		}, approvedBy)
		if err != nil {
			htmxServiceError(w, r, "Failed to approve hook", err)
			return
		}

		renderPluginHooks(w, r, svc, pluginName, "Hook approved")
	}
}

// PluginRevokeHookHandler revokes a single plugin hook.
func PluginRevokeHookHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			htmxError(w, r, "Invalid form data", http.StatusBadRequest)
			return
		}

		pluginName := r.FormValue("plugin")
		event := r.FormValue("event")
		table := r.FormValue("table")
		if pluginName == "" || event == "" {
			htmxError(w, r, "Missing hook fields", http.StatusBadRequest)
			return
		}

		err := svc.Plugins.RevokeHooks(r.Context(), []service.HookApprovalInput{
			{PluginName: pluginName, Event: event, Table: table},
		})
		if err != nil {
			htmxServiceError(w, r, "Failed to revoke hook", err)
			return
		}

		renderPluginHooks(w, r, svc, pluginName, "Hook revoked")
	}
}

// --- Helpers ---

// filterRoutesByPlugin returns only routes belonging to the named plugin.
func filterRoutesByPlugin(routes []service.PluginRoute, name string) []service.PluginRoute {
	var filtered []service.PluginRoute
	for _, r := range routes {
		if r.PluginName == name {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// filterHooksByPlugin returns only hooks belonging to the named plugin.
func filterHooksByPlugin(hooks []service.HookInfo, name string) []service.HookInfo {
	var filtered []service.HookInfo
	for _, h := range hooks {
		if h.PluginName == name {
			filtered = append(filtered, h)
		}
	}
	return filtered
}

// renderPluginRoutes re-fetches and renders plugin route rows after a mutation.
func renderPluginRoutes(w http.ResponseWriter, r *http.Request, svc *service.Registry, pluginName, toastMsg string) {
	routes, err := svc.Plugins.ListRoutes(r.Context())
	if err != nil {
		htmxServiceError(w, r, "Failed to reload routes", err)
		return
	}
	pluginRoutes := filterRoutesByPlugin(routes, pluginName)
	hasAdmin := HasPermission(r, "plugins:admin")
	csrfToken := CSRFTokenFromContext(r.Context())
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast": {"message": %q, "type": "success"}}`, toastMsg))
	Render(w, r, partials.PluginRoutesTableRows(pluginRoutes, pluginName, csrfToken, hasAdmin))
}

// renderPluginHooks re-fetches and renders plugin hook rows after a mutation.
func renderPluginHooks(w http.ResponseWriter, r *http.Request, svc *service.Registry, pluginName, toastMsg string) {
	hooks, err := svc.Plugins.ListHooks(r.Context())
	if err != nil {
		htmxServiceError(w, r, "Failed to reload hooks", err)
		return
	}
	pluginHooks := filterHooksByPlugin(hooks, pluginName)
	hasAdmin := HasPermission(r, "plugins:admin")
	csrfToken := CSRFTokenFromContext(r.Context())
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast": {"message": %q, "type": "success"}}`, toastMsg))
	Render(w, r, partials.PluginHooksTableRows(pluginHooks, pluginName, csrfToken, hasAdmin))
}

// htmxError sends an HTMX toast error or a plain HTTP error.
func htmxError(w http.ResponseWriter, r *http.Request, msg string, status int) {
	if IsHTMX(r) {
		w.Header().Set("HX-Retarget", "#none")
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast": {"message": %q, "type": "error"}}`, msg))
		w.WriteHeader(status)
		return
	}
	http.Error(w, msg, status)
}

// htmxServiceError logs the error and sends an HTMX toast.
func htmxServiceError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	utility.DefaultLogger.Error(msg, err)
	if service.IsConflict(err) {
		htmxError(w, r, err.Error(), http.StatusConflict)
		return
	}
	htmxError(w, r, msg, http.StatusInternalServerError)
}

// htmxRedirectToDetail triggers an HTMX redirect back to the plugin detail page
// with a success toast.
func htmxRedirectToDetail(w http.ResponseWriter, name, toastMsg string) {
	w.Header().Set("HX-Redirect", "/admin/plugins/"+name)
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast": {"message": %q, "type": "success"}}`, toastMsg))
	w.WriteHeader(http.StatusOK)
}

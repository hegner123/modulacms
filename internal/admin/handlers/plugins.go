package handlers

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/admin/pages"
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

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Plugin: `+name+`"}`)
			Render(w, r, pages.PluginDetailContent(detail))
			return
		}

		layout := NewAdminData(r, "Plugin: "+name)
		Render(w, r, pages.PluginDetail(layout, detail))
	}
}

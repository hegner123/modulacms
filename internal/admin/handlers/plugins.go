package handlers

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/db"
)

// PluginsListHandler lists installed plugins.
// Currently a placeholder since the plugin system is in development.
func PluginsListHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		layout := NewAdminData(r, "Plugins")
		Render(w, r, pages.PluginsList(layout))
	}
}

// PluginDetailHandler shows plugin detail.
// Currently a placeholder since the plugin system is in development.
func PluginDetailHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			http.NotFound(w, r)
			return
		}

		layout := NewAdminData(r, "Plugin: "+name)
		Render(w, r, pages.PluginDetail(layout, name))
	}
}

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
		RenderNav(w, r, "Plugins", pages.PluginsListContent(), pages.PluginsList(layout))
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
		RenderNav(w, r, "Plugin: "+name, pages.PluginDetailContent(name), pages.PluginDetail(layout, name))
	}
}

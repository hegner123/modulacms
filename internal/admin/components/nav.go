package components

import (
	"strings"

	"github.com/hegner123/modulacms/internal/middleware"
)

// NavItem represents a single navigation entry in the sidebar.
type NavItem struct {
	Label      string
	Href       string
	Icon       string
	Permission string
}

// NavItems is the ordered list of sidebar navigation entries.
var NavItems = []NavItem{
	{Label: "Dashboard", Href: "/admin/", Icon: "home"},
	{Label: "Content", Href: "/admin/content", Icon: "file-text", Permission: "content:read"},
	{Label: "Media", Href: "/admin/media", Icon: "image", Permission: "media:read"},
	{Label: "Datatypes", Href: "/admin/schema/datatypes", Icon: "blocks", Permission: "datatypes:read"},
	{Label: "Routes", Href: "/admin/routes", Icon: "globe", Permission: "routes:read"},
	{Label: "Users", Href: "/admin/users", Icon: "users", Permission: "users:read"},
	{Label: "Roles", Href: "/admin/users/roles", Icon: "shield", Permission: "roles:read"},
	{Label: "Tokens", Href: "/admin/users/tokens", Icon: "key", Permission: "tokens:read"},
	{Label: "Plugins", Href: "/admin/plugins", Icon: "puzzle", Permission: "plugins:read"},
	{Label: "Import", Href: "/admin/import", Icon: "upload", Permission: "import:create"},
	{Label: "Audit", Href: "/admin/audit", Icon: "history"},
	{Label: "Settings", Href: "/admin/settings", Icon: "settings", Permission: "config:read"},
	{Label: "Demo", Href: "/admin/demo", Icon: "flask-conical"},
}

// IsActive returns true if the current path matches or is a child of the nav item's href.
func IsActive(current, href string) bool {
	if href == "/admin/" {
		return current == "/admin/" || current == "/admin"
	}
	if !strings.HasPrefix(current, href) {
		return false
	}
	if len(current) == len(href) {
		return true
	}
	return current[len(href)] == '/'
}

// CanSee returns true if the nav item has no permission requirement or the user has the required permission.
func CanSee(item NavItem, perms middleware.PermissionSet) bool {
	return item.Permission == "" || perms.Has(item.Permission)
}

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
	Section    string // Section header displayed before this item when it changes.
}

// NavItems is the ordered list of sidebar navigation entries.
// Items are grouped by Section; a new section header renders when the
// Section value changes from the previous item.
var NavItems = []NavItem{
	// (no section header)
	{Label: "Dashboard", Href: "/admin/", Icon: "home"},

	// Content
	{Section: "Content", Label: "Content", Href: "/admin/content", Icon: "file-text", Permission: "content:read"},
	{Section: "Content", Label: "Datatypes", Href: "/admin/datatypes", Icon: "blocks", Permission: "datatypes:read"},
	{Section: "Content", Label: "Field Types", Href: "/admin/field-types", Icon: "tag", Permission: "field_types:read"},
	{Section: "Content", Label: "Routes", Href: "/admin/routes", Icon: "globe", Permission: "routes:read"},

	// Media
	{Section: "Media", Label: "Media", Href: "/admin/media", Icon: "image", Permission: "media:read"},
	{Section: "Media", Label: "Dimensions", Href: "/admin/media/dimensions", Icon: "ruler", Permission: "media:read"},

	// Admin Panel
	{Section: "Admin Panel", Label: "Admin Content", Href: "/admin/admin-content", Icon: "layout", Permission: "content:read"},
	{Section: "Admin Panel", Label: "Admin Datatypes", Href: "/admin/admin-datatypes", Icon: "blocks", Permission: "datatypes:read"},
	{Section: "Admin Panel", Label: "Admin Field Types", Href: "/admin/admin-field-types", Icon: "tag", Permission: "field_types:read"},
	{Section: "Admin Panel", Label: "Admin Routes", Href: "/admin/admin-routes", Icon: "globe", Permission: "routes:read"},

	// Users & Access
	{Section: "Users & Access", Label: "Users", Href: "/admin/users", Icon: "users", Permission: "users:read"},
	{Section: "Users & Access", Label: "Roles", Href: "/admin/users/roles", Icon: "shield", Permission: "roles:read"},
	{Section: "Users & Access", Label: "Tokens", Href: "/admin/users/tokens", Icon: "key", Permission: "tokens:read"},
	{Section: "Users & Access", Label: "Sessions", Href: "/admin/sessions", Icon: "monitor", Permission: "sessions:read"},

	// System
	{Section: "System", Label: "Plugins", Href: "/admin/plugins", Icon: "puzzle", Permission: "plugins:read"},
	{Section: "System", Label: "Pipelines", Href: "/admin/pipelines", Icon: "git-branch", Permission: "plugins:read"},
	{Section: "System", Label: "Tables", Href: "/admin/tables", Icon: "database", Permission: "tables:read"},
	{Section: "System", Label: "Deploy", Href: "/admin/deploy", Icon: "rocket", Permission: "deploy:read"},
	{Section: "System", Label: "Import", Href: "/admin/import", Icon: "upload", Permission: "import:create"},
	{Section: "System", Label: "Audit", Href: "/admin/audit", Icon: "history"},

	// Settings
	{Section: "Settings", Label: "Settings", Href: "/admin/settings", Icon: "settings", Permission: "config:read"},
	{Section: "Settings", Label: "Locales", Href: "/admin/settings/locales", Icon: "globe", Permission: "locale:read"},
	{Section: "Settings", Label: "Webhooks", Href: "/admin/settings/webhooks", Icon: "webhook", Permission: "webhook:read"},
	{Section: "Settings", Label: "Backups", Href: "/admin/settings/backups", Icon: "archive", Permission: "backup:create"},
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

// CanSeeSection returns true if at least one item in the section is visible.
func CanSeeSection(section string, perms middleware.PermissionSet) bool {
	for _, item := range NavItems {
		if item.Section == section && CanSee(item, perms) {
			return true
		}
	}
	return false
}

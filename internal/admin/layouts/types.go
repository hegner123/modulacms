package layouts

import (
	"github.com/a-h/templ"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/middleware"
)

type AdminData struct {
	Title       string
	CurrentPath string
	User        *db.Users
	Permissions middleware.PermissionSet
	IsAdmin     bool
	CSRFToken   string
	Version     string
	Dialogs     templ.Component
}

// WithDialogs returns a copy of AdminData with the Dialogs field set.
func (d AdminData) WithDialogs(c templ.Component) AdminData {
	d.Dialogs = c
	return d
}

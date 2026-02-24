package layouts

import (
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
}

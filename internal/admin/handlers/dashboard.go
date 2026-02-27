package handlers

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/db"
)

// DashboardHandler renders the admin dashboard.
func DashboardHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		layout := NewAdminData(r, "Dashboard")
		RenderNav(w, r, "Dashboard", pages.DashboardContent(), pages.Dashboard(layout))
	}
}

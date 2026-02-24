package partials

import (
	"strconv"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// PaginationPageData holds pagination state for use in templ partials.
// This type lives in partials to avoid import cycles between handlers and pages/partials.
type PaginationPageData struct {
	Current    int
	TotalPages int
	Limit      int64
	Target     string
	BaseURL    string
}

// IntToStr converts an int to its string representation.
func IntToStr(i int) string {
	return strconv.Itoa(i)
}

// pages returns a slice of page numbers from 1 to totalPages.
func pages(totalPages int) []int {
	p := make([]int, 0, totalPages)
	for i := 1; i <= totalPages; i++ {
		p = append(p, i)
	}
	return p
}

// urlForPage builds a paginated URL with limit and offset query parameters.
func urlForPage(baseURL string, page int, limit int64) string {
	offset := int64(page-1) * limit
	sep := "?"
	for _, c := range baseURL {
		if c == '?' {
			sep = "&"
			break
		}
	}
	return baseURL + sep + "limit=" + strconv.FormatInt(limit, 10) + "&offset=" + strconv.FormatInt(offset, 10)
}

// maskToken returns a masked version of a token, showing only the last 8 characters.
// Sensitive tokens must not be fully displayed in the admin UI.
func maskToken(token string) string {
	if len(token) <= 8 {
		return "********"
	}
	return "..." + token[len(token)-8:]
}

// truncateStr returns a shortened string, truncated with ellipsis if longer than maxLen.
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// nullStrDisplay extracts the string value from a db.NullString for display.
// Returns a dash for null/empty values.
func nullStrDisplay(ns db.NullString) string {
	if !ns.Valid || ns.String == "" {
		return "-"
	}
	return ns.String
}

// routeStatusLabel returns a human-readable label for a route status integer.
func routeStatusLabel(status int64) string {
	if status == 1 {
		return "Active"
	}
	return "Inactive"
}

// routeStatusBadgeClass returns the CSS class for a route status badge.
func routeStatusBadgeClass(status int64) string {
	if status == 1 {
		return "badge-published"
	}
	return "badge-draft"
}

// showEditDialog returns a JavaScript expression string to show a dialog by ID prefix + suffix.
func showEditDialog(id string) string {
	return "document.getElementById('edit-route-" + id + "').showModal()"
}

// countPermissions returns the number of permissions in a permission map.
func countPermissions(perms map[types.PermissionID]bool) int {
	if perms == nil {
		return 0
	}
	return len(perms)
}

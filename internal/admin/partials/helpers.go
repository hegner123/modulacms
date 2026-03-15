package partials

import (
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// PermissionCell represents a single cell in the permissions matrix.
type PermissionCell struct {
	PermissionID types.PermissionID
	Exists       bool
	Active       bool
}

// PermissionRow represents one resource row in the permissions matrix.
type PermissionRow struct {
	Resource    string
	DisplayName string
	Cells       map[string]PermissionCell
}

// PermissionMatrix holds the full grouped permissions matrix data.
type PermissionMatrix struct {
	AdminPerm PermissionCell
	Columns   []string
	Rows      []PermissionRow
}

// BuildPermissionMatrix groups permissions by resource:operation into a matrix.
func BuildPermissionMatrix(permissions []db.Permissions, activePerms map[types.PermissionID]bool) PermissionMatrix {
	m := PermissionMatrix{
		Columns: []string{"read", "create", "update", "delete", "admin"},
	}

	rowMap := make(map[string]*PermissionRow)

	for _, perm := range permissions {
		resource, operation := splitLabel(perm.Label)
		if operation == "" {
			// Standalone permission (e.g. "admin" with no colon)
			m.AdminPerm = PermissionCell{
				PermissionID: perm.PermissionID,
				Exists:       true,
				Active:       activePerms != nil && activePerms[perm.PermissionID],
			}
			continue
		}

		row, ok := rowMap[resource]
		if !ok {
			row = &PermissionRow{
				Resource:    resource,
				DisplayName: ResourceDisplayName(resource),
				Cells:       make(map[string]PermissionCell),
			}
			rowMap[resource] = row
		}

		row.Cells[operation] = PermissionCell{
			PermissionID: perm.PermissionID,
			Exists:       true,
			Active:       activePerms != nil && activePerms[perm.PermissionID],
		}
	}

	m.Rows = make([]PermissionRow, 0, len(rowMap))
	for _, row := range rowMap {
		m.Rows = append(m.Rows, *row)
	}
	sort.Slice(m.Rows, func(i, j int) bool {
		return m.Rows[i].Resource < m.Rows[j].Resource
	})

	return m
}

// splitLabel splits a "resource:operation" label by scanning for ':'.
func splitLabel(label string) (resource, operation string) {
	for i := 0; i < len(label); i++ {
		if label[i] == ':' {
			return label[:i], label[i+1:]
		}
	}
	return label, ""
}

// ResourceDisplayName converts a snake_case resource name to a display name.
func ResourceDisplayName(s string) string {
	switch s {
	case "admin_tree":
		return "Admin Tree"
	case "admin_field_types":
		return "Admin Field Types"
	case "ssh_keys":
		return "SSH Keys"
	case "field_types":
		return "Field Types"
	case "content_data":
		return "Content Data"
	}
	// Default: capitalize first letter of each word, split on '_'
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			runes := []rune(p)
			runes[0] = unicode.ToUpper(runes[0])
			parts[i] = string(runes)
		}
	}
	return strings.Join(parts, " ")
}

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

// buildMediaFolderTree assembles a flat list of folders into a nested tree structure.
func buildMediaFolderTree(folders []db.MediaFolder) []MediaFolderNode {
	if len(folders) == 0 {
		return nil
	}

	folderByID := make(map[types.MediaFolderID]db.MediaFolder, len(folders))
	for _, f := range folders {
		folderByID[f.FolderID] = f
	}

	childrenOf := make(map[types.MediaFolderID][]types.MediaFolderID)
	var rootIDs []types.MediaFolderID

	for _, f := range folders {
		if !f.ParentID.Valid {
			rootIDs = append(rootIDs, f.FolderID)
		} else {
			pid := types.MediaFolderID(f.ParentID.ID)
			childrenOf[pid] = append(childrenOf[pid], f.FolderID)
		}
	}

	var buildNode func(id types.MediaFolderID) MediaFolderNode
	buildNode = func(id types.MediaFolderID) MediaFolderNode {
		f := folderByID[id]
		node := MediaFolderNode{
			Folder:   f,
			Children: make([]MediaFolderNode, 0, len(childrenOf[id])),
		}
		for _, childID := range childrenOf[id] {
			node.Children = append(node.Children, buildNode(childID))
		}
		return node
	}

	roots := make([]MediaFolderNode, 0, len(rootIDs))
	for _, rid := range rootIDs {
		roots = append(roots, buildNode(rid))
	}
	return roots
}

// escapeJS escapes a string for safe inclusion in JavaScript string literals.
func escapeJS(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `'`, `\'`, `"`, `\"`, "\n", `\n`, "\r", `\r`)
	return r.Replace(s)
}

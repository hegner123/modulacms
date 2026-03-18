package partials

import (
	"encoding/json"
	"fmt"
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

// routeStatusBadgeClassTW returns Tailwind UI badge classes for a route status.
func routeStatusBadgeClassTW(status int64) string {
	if status == 1 {
		return "bg-green-400/10 text-green-400 ring-green-400/20"
	}
	return "bg-gray-400/10 text-gray-400 ring-gray-400/20"
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

// SelectOption represents a value/label pair for a select dropdown.
type SelectOption struct {
	Value string
	Label string
}

// DatatypeParentOptions builds select options for the parent datatype dropdown.
// excludeID is excluded from the list (to prevent self-reference in edit forms).
func DatatypeParentOptions(datatypes []db.Datatypes, excludeID string) []SelectOption {
	opts := make([]SelectOption, 0, len(datatypes)+1)
	opts = append(opts, SelectOption{Value: "", Label: "None (root level)"})
	for _, dt := range datatypes {
		if dt.DatatypeID.String() == excludeID {
			continue
		}
		opts = append(opts, SelectOption{
			Value: dt.DatatypeID.String(),
			Label: dt.Label,
		})
	}
	return opts
}

// displayTimestamp formats a Timestamp for human-readable display.
// Output: "Mar 16, 2026 at 6:17:16 PM UTC"
func displayTimestamp(ts types.Timestamp) string {
	if !ts.Valid {
		return "-"
	}
	return ts.Time.UTC().Format("Jan 2, 2006 at 3:04:05 PM MST")
}

// prettyJSON returns indented JSON from a JSONData value.
func prettyJSON(j types.JSONData) string {
	if !j.Valid || j.Data == nil {
		return "{}"
	}
	raw, err := json.Marshal(j.Data)
	if err != nil {
		return j.String()
	}
	var indented json.RawMessage
	if json.Unmarshal(raw, &indented) != nil {
		return string(raw)
	}
	pretty, err := json.MarshalIndent(indented, "", "  ")
	if err != nil {
		return string(raw)
	}
	return string(pretty)
}

// hlcStr formats an HLC timestamp for display.
func hlcStr(hlc types.HLC) string {
	return fmt.Sprintf("%d (physical: %s, logical: %d)", int64(hlc), hlc.Physical().Format("2006-01-02 15:04:05.000"), hlc.Logical())
}

// nullableStr extracts the string from a NullableString for display.
func nullableStr(ns types.NullableString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

// DiffEntry represents a single field change between old and new JSON values.
type DiffEntry struct {
	Key  string
	Kind string // "added", "removed", "changed"
	Old  string
	New  string
}

// computeDiff compares two JSONData values and returns field-level differences.
func computeDiff(oldValues, newValues types.JSONData) []DiffEntry {
	oldMap := jsonDataToMap(oldValues)
	newMap := jsonDataToMap(newValues)

	// Collect all keys
	keySet := make(map[string]bool)
	for k := range oldMap {
		keySet[k] = true
	}
	for k := range newMap {
		keySet[k] = true
	}

	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var diffs []DiffEntry
	for _, key := range keys {
		oldVal, inOld := oldMap[key]
		newVal, inNew := newMap[key]

		if inOld && !inNew {
			diffs = append(diffs, DiffEntry{Key: key, Kind: "removed", Old: formatValue(oldVal)})
		} else if !inOld && inNew {
			diffs = append(diffs, DiffEntry{Key: key, Kind: "added", New: formatValue(newVal)})
		} else {
			oldStr := formatValue(oldVal)
			newStr := formatValue(newVal)
			if oldStr != newStr {
				diffs = append(diffs, DiffEntry{Key: key, Kind: "changed", Old: oldStr, New: newStr})
			}
		}
	}
	return diffs
}

// jsonDataToMap attempts to convert JSONData to a string-keyed map.
func jsonDataToMap(j types.JSONData) map[string]any {
	if !j.Valid || j.Data == nil {
		return nil
	}
	if m, ok := j.Data.(map[string]any); ok {
		return m
	}
	// Try re-marshaling in case the underlying type is different
	raw, err := json.Marshal(j.Data)
	if err != nil {
		return nil
	}
	var m map[string]any
	if json.Unmarshal(raw, &m) != nil {
		return nil
	}
	return m
}

// formatValue converts an arbitrary value to a display string.
func formatValue(v any) string {
	if v == nil {
		return "null"
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}

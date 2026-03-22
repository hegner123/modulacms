package pages

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// truncateID returns a shortened version of an ID string for display.
// Shows first 8 characters followed by ellipsis.
func truncateID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[:8] + "..."
}

// ContentListItem wraps ContentDataTopLevel with human-readable display fields
// resolved from the associated route or title content field.
type ContentListItem struct {
	db.ContentDataTopLevel
	DisplayName    string
	Slug           string
	HasPublishPerm bool
}

// AdminContentListItem wraps AdminContentDataTopLevel with human-readable display fields
// resolved from the associated admin route or title content field.
type AdminContentListItem struct {
	db.AdminContentDataTopLevel
	DisplayName    string
	HasPublishPerm bool
}

// nullStr extracts the string value from a NullString, returning empty string if not valid.
func nullStr(ns db.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

// isImage returns true if the mimetype indicates an image file.
func isImage(mimetype string) bool {
	return strings.HasPrefix(mimetype, "image/")
}

// fileExtension extracts the file extension from a filename.
// Returns the extension including the dot, or an empty string.
func fileExtension(filename string) string {
	if dotIdx := strings.LastIndex(filename, "."); dotIdx >= 0 {
		return strings.ToUpper(filename[dotIdx+1:])
	}
	return "FILE"
}

// nullableIDValue returns the ID string for a valid NullableDatatypeID, or "" if null.
func nullableIDValue(n types.NullableDatatypeID) string {
	if !n.Valid {
		return ""
	}
	return n.ID.String()
}

// nullableAdminDatatypeIDValue returns the ID string for a valid NullableAdminDatatypeID, or "" if null.
func nullableAdminDatatypeIDValue(n types.NullableAdminDatatypeID) string {
	if !n.Valid {
		return ""
	}
	return n.ID.String()
}

// mediaGridItem is the JSON shape for media files in <mcms-media-grid>.
type mediaGridItem struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	URL         string `json:"url,omitempty"`
	Alt         string `json:"alt,omitempty"`
	DisplayName string `json:"displayName"`
	Mimetype    string `json:"mimetype,omitempty"`
	Name        string `json:"name"`
	DateCreated string `json:"dateCreated,omitempty"`
}

// mediaGridJSON converts folders and media into a single JSON array for the grid component.
// Folders appear first (type:"folder"), then media files (type:"file").
func mediaGridJSON(folders []db.MediaFolder, items []db.Media) string {
	out := make([]mediaGridItem, 0, len(folders)+len(items))
	for _, f := range folders {
		out = append(out, mediaGridItem{
			ID:          f.FolderID.String(),
			Type:        "folder",
			DisplayName: f.Name,
			Name:        f.Name,
		})
	}
	for _, m := range items {
		out = append(out, mediaGridItem{
			ID:          m.MediaID.String(),
			Type:        "file",
			URL:         m.URL.String(),
			Alt:         nullStr(m.Alt),
			DisplayName: nullStr(m.DisplayName),
			Mimetype:    nullStr(m.Mimetype),
			Name:        nullStr(m.Name),
			DateCreated: m.DateCreated.String(),
		})
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// focalStr converts a NullableFloat64 to its string representation for template attributes.
// Returns empty string if the value is not valid.
func focalStr(n types.NullableFloat64) string {
	if !n.Valid {
		return ""
	}
	return strconv.FormatFloat(n.Float64, 'f', -1, 64)
}

// isRoleSelected checks if a role ID appears in the field's roles JSON array.
func isRoleSelected(rolesJSON types.NullableString, roleID string) bool {
	if !rolesJSON.Valid {
		return false
	}
	var roles []string
	if err := json.Unmarshal([]byte(rolesJSON.String), &roles); err != nil {
		return false
	}
	for _, r := range roles {
		if r == roleID {
			return true
		}
	}
	return false
}

// ReservedTypeInfo holds a reserved datatype type and its description for display.
type ReservedTypeInfo struct {
	Type        string
	Description string
	Example     string // optional suffixed example (e.g. "_global_menu")
	UserSet     bool   // true if users can set this type; false if engine-assigned only
}

// reservedTypesList returns an ordered list of reserved datatype types for UI display.
func reservedTypesList() []ReservedTypeInfo {
	return []ReservedTypeInfo{
		{Type: "_root", Description: "Tree entry point. One per route — the top-level node for a content tree.", UserSet: true},
		{Type: "_reference", Description: "Triggers tree composition. Resolves _id field values and attaches referenced trees as children. Use a suffix to target specific content types.", Example: "_reference_menu", UserSet: true},
		{Type: "_collection", Description: "Queryable collection. Signals to clients that children support filtering, sorting, and pagination.", Example: "_collection_blog", UserSet: true},
		{Type: "_global", Description: "Singleton site-wide content (menus, footers, settings). No route association, delivered via /globals endpoint.", Example: "_global_menu", UserSet: true},
		{Type: "_nested_root", Description: "Root of a composed subtree. Assigned automatically by the engine during tree composition.", UserSet: false},
		{Type: "_system_log", Description: "Synthetic node injected when a reference cannot be resolved. Engine-assigned only.", UserSet: false},
		{Type: "_plugin", Description: "Plugin-provided content. Actual types use _plugin_{name} namespace, registered by plugins on init.", UserSet: false},
	}
}

// stepNumber converts an int step to its display string.
func stepNumber(step int) string {
	return fmt.Sprintf("%d", step)
}

// revisionStr converts a revision number to a display string.
func revisionStr(rev int64) string {
	return fmt.Sprintf("%d", rev)
}

// statusOptions returns the content status options for a select dropdown.
func statusOptions() []partials.SelectOption {
	return []partials.SelectOption{
		{Value: "draft", Label: "Draft"},
		{Value: "published", Label: "Published"},
	}
}

// adminMediaGridJSON converts admin media folders and media into a single JSON array for the grid component.
// Folders appear first (type:"folder"), then media files (type:"file").
func adminMediaGridJSON(folders []db.AdminMediaFolder, items []db.AdminMedia) string {
	out := make([]mediaGridItem, 0, len(folders)+len(items))
	for _, f := range folders {
		out = append(out, mediaGridItem{
			ID:          f.AdminFolderID.String(),
			Type:        "folder",
			DisplayName: f.Name,
			Name:        f.Name,
		})
	}
	for _, m := range items {
		out = append(out, mediaGridItem{
			ID:          m.AdminMediaID.String(),
			Type:        "file",
			URL:         m.URL.String(),
			Alt:         nullStr(m.Alt),
			DisplayName: nullStr(m.DisplayName),
			Mimetype:    nullStr(m.Mimetype),
			Name:        nullStr(m.Name),
			DateCreated: m.DateCreated.String(),
		})
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "[]"
	}
	return string(b)
}

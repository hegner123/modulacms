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


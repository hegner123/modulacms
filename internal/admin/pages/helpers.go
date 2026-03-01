package pages

import (
	"encoding/json"
	"strings"

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

package tui

import (
	"slices"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/tree"
)

// DecideNodeName determines the display name for a tree node based on its fields and datatype.
func DecideNodeName(node tree.Node) string {
	var out string
	if index := slices.IndexFunc(node.Fields, FieldMatchesLabel); index > -1 {
		id := node.Fields[index].FieldID
		contentIndex := slices.IndexFunc(node.InstanceFields, func(cf db.ContentFields) bool {
			return cf.FieldID.Valid && cf.FieldID.ID == id
		})
		if contentIndex > -1 {
			out += node.InstanceFields[contentIndex].FieldValue
			out += "  ["
			out += node.Datatype.Label
			out += "]"
		} else {
			out += node.Datatype.Label
		}
	} else {
		out += node.Datatype.Label
	}
	return out
}

// FieldMatchesLabel checks if a field's label matches a label field identifier.
func FieldMatchesLabel(field db.Fields) bool {
	ValidLabelFields := []string{"Label", "label", "Title", "title", "Name", "name"}
	return slices.Contains(ValidLabelFields, field.Label)
}

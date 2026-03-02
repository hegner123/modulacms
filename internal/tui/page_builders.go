package tui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree"
)

// CMSPage holds content tree rendering methods used by the panel view.
type CMSPage struct{}

// ProcessTreeDatatypes processes and renders the content tree from the model.
func (c CMSPage) ProcessTreeDatatypes(model Model) string {
	if model.PageRouteId.IsZero() {
		return "No route selected\n\nPlease select a route to view content."
	}
	if model.Root.Root == nil {
		return "No content loaded"
	}
	display := make([]string, 0)
	currentIndex := 0
	c.traverseTree(model.Root.Root, &display, model.Cursor, &currentIndex)
	if len(display) == 0 {
		return "Empty content tree"
	}
	return lipgloss.JoinVertical(lipgloss.Top, display...)
}

// traverseTree recursively renders the tree with proper indentation and cursor.
func (c CMSPage) traverseTree(node *tree.Node, display *[]string, cursor int, currentIndex *int) {
	c.traverseTreeWithDepth(node, display, cursor, currentIndex, 0)
}

// traverseTreeWithDepth recursively renders the tree tracking depth for indentation.
func (c CMSPage) traverseTreeWithDepth(node *tree.Node, display *[]string, cursor int, currentIndex *int, depth int) {
	if node == nil {
		return
	}

	// Render current node with cursor and depth-based indentation
	row := c.FormatTreeRow(node, *currentIndex == cursor, depth)
	*display = append(*display, row)
	*currentIndex++

	// Render children if expanded (increase depth)
	if node.Expand && node.FirstChild != nil {
		c.traverseTreeWithDepth(node.FirstChild, display, cursor, currentIndex, depth+1)
	}

	// Render siblings (same depth)
	if node.NextSibling != nil {
		c.traverseTreeWithDepth(node.NextSibling, display, cursor, currentIndex, depth)
	}
}

// FormatTreeRow formats a single tree node with cursor and indentation.
func (c CMSPage) FormatTreeRow(node *tree.Node, isSelected bool, depth int) string {
	indent := strings.Repeat("  ", depth)

	// Icon based on node type and state
	icon := "├─"
	if node.FirstChild != nil {
		if node.Expand {
			icon = "▼"
		} else {
			icon = "▶"
		}
	}

	// Get node name
	name := DecideNodeName(*node)

	// Selection indicator (cursor only, no highlighting)
	cursor := "  "
	if isSelected {
		cursor = "->"
	}

	// Status indicator for published/archived content
	statusMark := ""
	if node.Instance != nil {
		if node.Instance.Status == types.ContentStatusPublished {
			statusMark = lipgloss.NewStyle().Foreground(lipgloss.Color("#16a34a")).Render("● ")
		} else {
			statusMark = lipgloss.NewStyle().Foreground(lipgloss.Color("#ca8a04")).Render("○ ")
		}
	}

	// Build the row: cursor + indent + icon + statusMark + name
	return cursor + indent + icon + " " + statusMark + name
}

// FormatRow formats a tree node as a string row with indentation and wrapping.
func FormatRow(node *tree.Node) string {
	row := ""
	Indent := "  "
	Wrapped := ">>"
	row += strings.Repeat(Wrapped, node.Wrapped)
	row += strings.Repeat(Indent, node.Indent-(node.Wrapped-1))
	row += DecideNodeName(*node)

	return row
}

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

// resolveAuthorName looks up the author's display name from the users list.
// Returns the username if found, or the raw ID string as fallback.
func resolveAuthorName(authorID types.UserID, users []db.UserWithRoleLabelRow) string {
	for _, u := range users {
		if u.UserID == authorID {
			if u.Name != "" {
				return u.Name
			}
			return u.Username
		}
	}
	return string(authorID)
}

// ProcessContentPreview generates a preview of the selected content node,
// showing metadata at the top and content field values below.
func (c CMSPage) ProcessContentPreview(model Model) string {
	node := model.Root.NodeAtIndex(model.Cursor)
	if node == nil {
		return "No content selected"
	}

	preview := []string{}

	// Title/Name
	title := DecideNodeName(*node)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(config.DefaultStyle.Accent2)
	preview = append(preview, titleStyle.Render(title))
	preview = append(preview, "")

	// Metadata line: Type | Status | Author
	metaParts := []string{node.Datatype.Label}
	metaParts = append(metaParts, string(node.Instance.Status))
	if !node.Instance.AuthorID.IsZero() {
		metaParts = append(metaParts, resolveAuthorName(node.Instance.AuthorID, model.UsersList))
	}
	dimStyle := lipgloss.NewStyle().Faint(true)
	preview = append(preview, dimStyle.Render(strings.Join(metaParts, " | ")))
	preview = append(preview, "")

	// Content field values preview
	if len(model.SelectedContentFields) > 0 {
		labelStyle := lipgloss.NewStyle().Bold(true).Foreground(config.DefaultStyle.Accent)
		for _, cf := range model.SelectedContentFields {
			preview = append(preview, labelStyle.Render(cf.Label))
			if cf.Value == "" {
				preview = append(preview, dimStyle.Render("  (empty)"))
			} else {
				// Wrap long values to fit the panel
				lines := strings.Split(cf.Value, "\n")
				for _, line := range lines {
					preview = append(preview, fmt.Sprintf("  %s", line))
				}
			}
			preview = append(preview, "")
		}
	}

	// Children summary
	if node.FirstChild != nil {
		preview = append(preview, dimStyle.Render("Children:"))
		child := node.FirstChild
		for child != nil {
			childName := DecideNodeName(*child)
			preview = append(preview, dimStyle.Render(fmt.Sprintf("  %s (%s)", childName, child.Datatype.Label)))
			child = child.NextSibling
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, preview...)
}

// ProcessFields renders the fields of the selected content node.
func (c CMSPage) ProcessFields(model Model) string {
	if len(model.SelectedContentFields) == 0 {
		return "No fields"
	}

	fields := []string{}
	for i, cf := range model.SelectedContentFields {
		cursor := "   "
		if model.PanelFocus == RoutePanel && model.FieldCursor == i {
			cursor = " > "
		}

		value := cf.Value
		if value == "" {
			value = "(empty)"
		} else if len(value) > 40 {
			value = value[:37] + "..."
		}

		fields = append(fields, fmt.Sprintf("%s%s: %s", cursor, cf.Label, value))
	}

	return lipgloss.JoinVertical(lipgloss.Left, fields...)
}

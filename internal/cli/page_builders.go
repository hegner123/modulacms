package cli

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree"
	"github.com/hegner123/modulacms/internal/tui"
)

// PageUI defines the interface for page rendering components.
type PageUI interface {
	AddTitle(string)
	AddHeader(string)
	AddBody(string)
	AddControls(string)
	AddStatus(string)
	Render(int) string
}

// BasePage is the base structure for all page types with common UI elements.
type BasePage struct {
	Title    string
	Header   string
	Body     string
	Controls string
	Status   string
}

// NewBasePage creates a new BasePage with empty fields.
func NewBasePage() BasePage {
	return BasePage{
		Title:    "",
		Header:   "",
		Body:     "",
		Controls: "",
		Status:   "",
	}
}

// StaticPage is a page type for displaying static content.
type StaticPage struct {
	BasePage
}

// AddTitle adds a title to the static page.
func (s *StaticPage) AddTitle(t string) {
	s.Title += t
}

// AddHeader adds a header to the static page.
func (s *StaticPage) AddHeader(h string) {
	s.Header += h
}

// AddBody adds body content to the static page.
func (s *StaticPage) AddBody(b string) {
	s.Body += b
}

// AddControls adds control text to the static page.
func (s *StaticPage) AddControls(c string) {
	s.Controls += c
}

// AddStatus adds status text to the static page.
func (s *StaticPage) AddStatus(st string) {
	s.Status += st
}

// Render returns the rendered representation of the static page.
func (s StaticPage) Render(model Model) string {
	docStyle := lipgloss.NewStyle().Padding(1, 2, 1, 2)
	rows := []string{RenderTitle(s.Title), RenderHeading(s.Header), s.Body}
	if model.DialogActive {
		rows = append(rows, model.Dialog.Render(model.Width, model.Height))
	}
	ui := lipgloss.JoinVertical(
		lipgloss.Left,
		rows...,
	)

	h := model.RenderSpace(docStyle.Render(ui) + RenderFooter(s.Controls))
	f := RenderFooter(s.Controls)
	status := s.Status
	return lipgloss.JoinVertical(
		lipgloss.Left,
		docStyle.Render(ui),
		h,
		f,
		status,
	)
}

// NewStaticPage creates a new StaticPage.
func NewStaticPage() StaticPage {
	page := NewBasePage()
	return StaticPage{
		BasePage: page,
	}

}

// MenuPage is a page type for displaying a menu of options.
type MenuPage struct {
	BasePage
	Menu []string
}

// AddTitle adds a title to the menu page.
func (s *MenuPage) AddTitle(t string) {
	s.Title += t
}

// AddHeader adds a header to the menu page.
func (m *MenuPage) AddHeader(h string) {
	m.Header += h
}

// AddBody adds body content to the menu page.
func (m *MenuPage) AddBody(b string) {
	m.Body += b
}

// AddControls adds control text to the menu page.
func (m *MenuPage) AddControls(c string) {
	m.Controls += c
}

// AddStatus adds status text to the menu page.
func (m *MenuPage) AddStatus(st string) {
	m.Status += st
}

// AddMenu adds a menu to the menu page.
func (m *MenuPage) AddMenu(menu []string) {
	m.Menu = menu
}

// RenderBody renders the body of the menu page with formatted menu items.
func (m *MenuPage) RenderBody(model Model) string {
	r := make([]string, len(m.Body)+len(m.Menu))
	var row []string
	var column []string

	for i, choice := range m.Menu {

		cursor := "   "
		if model.Cursor == i {
			cursor = " ->"
		}

		fs := fmt.Sprintf("%s%s   ", cursor, choice)
		column = append(column, fs)
		if (i+1)%8 == 0 || i == len(m.Menu)-1 {
			c := NewVerticalGroup(lipgloss.Left, column)
			row = append(row, c)
			column = []string{}
		}
	}
	r = append(r, lipgloss.JoinHorizontal(lipgloss.Top, row...))
	r = append(r, m.Body)
	return RenderBorderFlex(lipgloss.JoinHorizontal(lipgloss.Center, r...))

}

// Render returns the rendered representation of the menu page.
func (m MenuPage) Render(model Model) string {
	docStyle := lipgloss.NewStyle().Padding(1, 2, 1, 2)
	s := lipgloss.JoinVertical(
		lipgloss.Left,
		RenderTitle(m.Title),
		RenderHeading(m.Header),
		m.RenderBody(model),
	)
	h := model.RenderSpace(docStyle.Render(s) + RenderFooter(m.Controls))
	f := RenderFooter(m.Controls)
	status := m.Status
	return lipgloss.JoinVertical(
		lipgloss.Left,
		docStyle.Render(s),
		h,
		f,
		status,
	)
}

// NewMenuPage creates a new MenuPage.
func NewMenuPage() MenuPage {
	basePage := NewBasePage()
	m := make([]string, 0)
	return MenuPage{
		BasePage: basePage,
		Menu:     m,
	}
}

// TablePage is a page type for displaying tabular data.
type TablePage struct {
	BasePage
	Table        string
	TableHeaders []string
	TableRows    [][]string
	TableUI      *table.Table
	PageMod      int
}

// AddTitle adds a title to the table page.
func (s *TablePage) AddTitle(t string) {
	s.Title += t
}

// AddHeader adds a header to the table page.
func (t *TablePage) AddHeader(h string) {
	t.Header += h
}

// AddHeaders adds table column headers to the table page.
func (t *TablePage) AddHeaders(h []string) {
	t.TableHeaders = h
}

// AddRows adds table rows to the table page.
func (t *TablePage) AddRows(r [][]string) {
	t.TableRows = r
}

// AddControls adds control text to the table page.
func (t *TablePage) AddControls(c string) {
	t.Controls += c
}

// AddStatus adds status text to the table page.
func (t *TablePage) AddStatus(st string) {
	t.Status += st
}

// AddBody adds body content to the table page.
func (t *TablePage) AddBody(b string) {
	t.Body += b
}

// RenderBody renders the body of the table page with paginated table data.
func (t *TablePage) RenderBody(m Model) string {
	if len(t.TableHeaders) == 0 {
		return t.Body
	}
	start, end := m.Paginator.GetSliceBounds(len(t.TableRows))
	currentView := t.TableRows[start:end]

	t.TableUI = TableRender(t.TableHeaders, currentView, m.Cursor)
	paginator := ""
	if len(t.TableRows) > m.MaxRows {
		paginator = "\n\n" + m.Paginator.View()
	}
	b := lipgloss.JoinVertical(
		lipgloss.Top,
		t.TableUI.Render(),
		paginator,
	)
	return b
}

// Render returns the rendered representation of the table page.
func (t TablePage) Render(model Model) string {
	docStyle := lipgloss.NewStyle().Padding(1, 2, 1, 2)
	s := lipgloss.JoinVertical(
		lipgloss.Left,
		RenderTitle(t.Title),
		RenderHeading(t.Header),
		t.RenderBody(model),
	)
	h := model.RenderSpace(docStyle.Render(s) + RenderFooter(t.Controls))
	f := RenderFooter(t.Controls)
	status := t.Status
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		docStyle.Render(s),
		h,
		f,
		status,
	)
	if model.DialogActive && model.Dialog != nil {
		d := model.Dialog
		old := content
		content = DialogOverlay(old, *d, model.Width, model.Height)
	}
	return content
}

// NewTablePage creates a new TablePage.
func NewTablePage() TablePage {
	basePage := NewBasePage()
	return TablePage{
		BasePage:     basePage,
		Table:        "",
		TableHeaders: make([]string, 0),
		TableRows:    make([][]string, 0),
	}
}

// FormPage is a page type for displaying forms.
type FormPage struct {
	BasePage
	Form *huh.Form
}

// AddTitle adds a title to the form page.
func (s *FormPage) AddTitle(t string) {
	s.Title += t
}

// AddHeader adds a header to the form page.
func (f *FormPage) AddHeader(h string) {
	f.Header += h
}

// AddControls adds control text to the form page.
func (f *FormPage) AddControls(c string) {
	f.Controls += c
}

// AddStatus adds status text to the form page.
func (f *FormPage) AddStatus(st string) {
	f.Status += st
}

// AddBody adds body content to the form page.
func (f *FormPage) AddBody(b string) {
	f.Body += b
}

// Render returns the rendered representation of the form page.
func (f FormPage) Render(model Model) string {
	docStyle := lipgloss.NewStyle().Padding(1, 2, 1, 2)
	form := ""
	if model.FormState.Form != nil {
		form = model.FormState.Form.View()

	}
	s := lipgloss.JoinVertical(
		lipgloss.Left,
		RenderTitle(f.Title),
		RenderHeading(f.Header),
		form,
	)
	h := model.RenderSpace(docStyle.Render(s) + RenderFooter(f.Controls))
	footer := RenderFooter(f.Controls)
	status := f.Status
	return lipgloss.JoinVertical(
		lipgloss.Left,
		docStyle.Render(s),
		h,
		footer,
		status,
	)
}

// NewFormPage creates a new FormPage.
func NewFormPage() FormPage {
	basePage := NewBasePage()
	return FormPage{
		BasePage: basePage,
	}
}

// DisplayMode represents the current display mode of the CMS page.
type DisplayMode int

const (
	Main         DisplayMode = iota
	NewDatatype              // Corresponds to a function that retrives the datatypes that can be a child of the current node.
	EditDatatype             // Corresponds to replacing the content preview with a form of fields for that type
	MoveDatatype             // Corresponds to a dialog where a cursor is used to select a node to move a node infront or behind.
	FindDatatype             // Corresponds to a dialog where you type and it finds fields that match, entering on one makes it active.
)

// CMSPage is a page type for displaying the CMS content tree interface.
type CMSPage struct {
	BasePage
	Tree    tree.Root
	Display DisplayMode
}

// AddTitle adds a title to the CMS page.
func (s *CMSPage) AddTitle(t string) {
	s.Title += t
}

// AddHeader adds a header to the CMS page.
func (c *CMSPage) AddHeader(h string) {
	c.Header += h
}

// AddControls adds control text to the CMS page.
func (c *CMSPage) AddControls(controls string) {
	c.Controls += controls
}

// AddStatus adds status text to the CMS page.
func (c *CMSPage) AddStatus(st string) {
	c.Status += st
}

// AddBody adds body content to the CMS page.
func (c *CMSPage) AddBody(b string) {
	c.Body += b
}

// RenderColumn renders a styled column with the specified width and content.
func (c *CMSPage) RenderColumn(width int, content string) string {
	colStyle := lipgloss.NewStyle().
		Background(config.DefaultStyle.PrimaryBG).
		Foreground(config.DefaultStyle.Primary).
		Width(width)
	return colStyle.Render(content)
}

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
		switch node.Instance.Status {
		case types.ContentStatusPublished:
			statusMark = "* "
		case types.ContentStatusArchived:
			statusMark = "~ "
		}
	}

	// Build the row: cursor + indent + icon + statusMark + name
	return cursor + indent + icon + " " + statusMark + name
}

// FormatRow formats a tree node as a string row with indentation and wrapping.
func FormatRow(node *tree.Node) string {
	row := ""
	//HasChildrenCollapsed := "+"
	//HasChildrenExpanded := "-"
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
		if model.PanelFocus == tui.RoutePanel && model.FieldCursor == i {
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

// CenterColumn processes the center column based on the current display mode.
func (c CMSPage) CenterColumn(content string) string {
	switch c.Display {
	case Main:
		return content
	default:
		return content
	}

}

// Render returns the rendered representation of the CMS page with three-panel layout.
func (c CMSPage) Render(model Model) string {
	docStyle := lipgloss.NewStyle().Padding(1, 2, 1, 2)
	col1 := c.ProcessTreeDatatypes(model)
	col2 := c.ProcessContentPreview(model)
	col3 := c.ProcessFields(model)
	editor := lipgloss.JoinHorizontal(
		lipgloss.Center,
		c.RenderColumn(model.Width/4, col1),
		c.RenderColumn(model.Width/2, col2),
		c.RenderColumn(model.Width/4, col3),
	)
	s := lipgloss.JoinVertical(
		lipgloss.Left,
		RenderTitle(c.Title),
		RenderHeading(c.Header),
		editor,
	)
	h := model.RenderSpace(docStyle.Render(s) + RenderFooter(c.Controls))
	footer := RenderFooter(c.Controls)
	status := c.Status
	return lipgloss.JoinVertical(
		lipgloss.Left,
		docStyle.Render(s),
		h,
		footer,
		status,
	)
}

// NewCMSPage creates a new CMSPage.
func NewCMSPage() CMSPage {
	b := NewBasePage()
	p := CMSPage{
		BasePage: b,
		Tree:     tree.Root{},
	}

	return p
}

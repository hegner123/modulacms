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
)

type PageUI interface {
	AddTitle(string)
	AddHeader(string)
	AddBody(string)
	AddControls(string)
	AddStatus(string)
	Render(int) string
}

type BasePage struct {
	Title    string
	Header   string
	Body     string
	Controls string
	Status   string
}

func NewBasePage() BasePage {
	return BasePage{
		Title:    "",
		Header:   "",
		Body:     "",
		Controls: "",
		Status:   "",
	}
}

type StaticPage struct {
	BasePage
}

func (s *StaticPage) AddTitle(t string) {
	s.Title += t
}

func (s *StaticPage) AddHeader(h string) {
	s.Header += h
}
func (s *StaticPage) AddBody(b string) {
	s.Body += b
}
func (s *StaticPage) AddControls(c string) {
	s.Controls += c
}
func (s *StaticPage) AddStatus(st string) {
	s.Status += st
}

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

func NewStaticPage() StaticPage {
	page := NewBasePage()
	return StaticPage{
		BasePage: page,
	}

}

type MenuPage struct {
	BasePage
	Menu []string
}

func (s *MenuPage) AddTitle(t string) {
	s.Title += t
}

func (m *MenuPage) AddHeader(h string) {
	m.Header += h
}
func (m *MenuPage) AddBody(b string) {
	m.Body += b
}
func (m *MenuPage) AddControls(c string) {
	m.Controls += c
}
func (m *MenuPage) AddStatus(st string) {
	m.Status += st
}

func (m *MenuPage) AddMenu(menu []string) {
	m.Menu = menu
}
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
		if (i+1)%6 == 0 || i == len(m.Menu)-1 {
			c := NewVerticalGroup(lipgloss.Left, column)
			row = append(row, c)
			column = []string{}
		}
	}
	list := make([]string, 0)
	for _, v := range model.History {
		list = append(list, v.Page.Label)
	}
	content := lipgloss.JoinVertical(lipgloss.Top, list...)
	row = append(row, content)
	r = append(r, lipgloss.JoinHorizontal(lipgloss.Top, row...))
	r = append(r, m.Body)
	return RenderBorderFlex(lipgloss.JoinHorizontal(lipgloss.Center, r...))

}

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

func NewMenuPage() MenuPage {
	basePage := NewBasePage()
	m := make([]string, 0)
	return MenuPage{
		BasePage: basePage,
		Menu:     m,
	}
}

type TablePage struct {
	BasePage
	Table        string
	TableHeaders []string
	TableRows    [][]string
	TableUI      *table.Table
	PageMod      int
}

func (s *TablePage) AddTitle(t string) {
	s.Title += t
}

func (t *TablePage) AddHeader(h string) {
	t.Header += h
}

func (t *TablePage) AddHeaders(h []string) {
	t.TableHeaders = h
}
func (t *TablePage) AddRows(r [][]string) {
	t.TableRows = r
}

func (t *TablePage) AddControls(c string) {
	t.Controls += c
}

func (t *TablePage) AddStatus(st string) {
	t.Status += st
}

func (t *TablePage) AddBody(b string) {
	t.Body += b
}

func (t *TablePage) RenderBody(m Model) string {
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

func NewTablePage() TablePage {
	basePage := NewBasePage()
	return TablePage{
		BasePage:     basePage,
		Table:        "",
		TableHeaders: make([]string, 0),
		TableRows:    make([][]string, 0),
	}
}

type FormPage struct {
	BasePage
	Form *huh.Form
}

func (s *FormPage) AddTitle(t string) {
	s.Title += t
}
func (f *FormPage) AddHeader(h string) {
	f.Header += h
}

func (f *FormPage) AddControls(c string) {
	f.Controls += c
}

func (f *FormPage) AddStatus(st string) {
	f.Status += st
}
func (f *FormPage) AddBody(b string) {
	f.Body += b
}

func (f FormPage) Render(model Model) string {
	docStyle := lipgloss.NewStyle().Padding(1, 2, 1, 2)
	form := ""
	if model.Form != nil {
		form = model.Form.View()

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

func NewFormPage() FormPage {
	basePage := NewBasePage()
	return FormPage{
		BasePage: basePage,
	}
}

type DisplayMode int

const (
	Main         DisplayMode = iota
	NewDatatype              // Corresponds to a function that retrives the datatypes that can be a child of the current node.
	EditDatatype             // Corresponds to replacing the content preview with a form of fields for that type
	MoveDatatype             // Corresponds to a dialog where a cursor is used to select a node to move a node infront or behind.
	FindDatatype             // Corresponds to a dialog where you type and it finds fields that match, entering on one makes it active.
)

type CMSPage struct {
	BasePage
	Tree    TreeRoot
	Display DisplayMode
}

func (s *CMSPage) AddTitle(t string) {
	s.Title += t
}

func (c *CMSPage) AddHeader(h string) {
	c.Header += h
}

func (c *CMSPage) AddControls(controls string) {
	c.Controls += controls
}

func (c *CMSPage) AddStatus(st string) {
	c.Status += st
}

func (c *CMSPage) AddBody(b string) {
	c.Body += b
}

func (c *CMSPage) RenderColumn(width int, content string) string {
	colStyle := lipgloss.NewStyle().
		Background(config.DefaultStyle.PrimaryBG).
		Foreground(config.DefaultStyle.Primary).
		Width(width)
	return colStyle.Render(content)
}

func (c CMSPage) ProcessTreeDatatypes(model Model) string {
	current := model.Root.Root
	display := make([]string, 0)
	index := 0
	for current != nil {
		row := FormatRow(current)
		display = append(display, row)
		index++
	}
	return lipgloss.JoinVertical(lipgloss.Top, display...)
}

func FormatRow(node *TreeNode) string {
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

func DecideNodeName(node TreeNode) string {
	var out string
	if index := slices.IndexFunc(node.Fields, FieldMatchesLabel); index > -1 {
		id := node.Fields[index].FieldID
		contentIndex := slices.IndexFunc(node.InstanceFields, func(cf db.ContentFields) bool {
			return cf.FieldID == id
		})
		out += node.InstanceFields[contentIndex].FieldValue
		out += "  ["
		out += node.Datatype.Label
		out += "]"

	} else {
		out += node.Datatype.Label
	}
	return out
}

func FieldMatchesLabel(field db.Fields) bool {
	ValidLabelFields := []string{"Label", "label", "Title", "title", "Name", "name"}
	return slices.Contains(ValidLabelFields, field.Label)
}

func (c CMSPage) ProcessContentPreview(tree TreeRoot) string {
	return "Content Preview"
}
func (c CMSPage) ProcessFields(tree TreeRoot) string {
	return "ProcessFields"
}

func (c CMSPage) CenterColumn(content string) string {
	switch c.Display {
	case Main:
		return content
	default:
		return content
	}

}

func (c CMSPage) Render(model Model) string {
	docStyle := lipgloss.NewStyle().Padding(1, 2, 1, 2)
	col1 := c.ProcessTreeDatatypes(model)
	col2 := c.ProcessContentPreview(model.Root)
	col3 := c.ProcessFields(model.Root)
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

func NewCMSPage() CMSPage {
	b := NewBasePage()
	p := CMSPage{
		BasePage: b,
		Tree:     TreeRoot{},
	}

	return p
}

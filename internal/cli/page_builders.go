package cli

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

/*
  2. Create a BasePage instance with a header and render function
  3. Set up any content needed for the BasePage (header, body, footer)
  4. Create specialized components needed for your page (MenuComponent, TableComponent, FormComponent, etc.)
  5. Define a concrete page struct that includes the BasePage and any specialized components
  6. Implement the View method for your concrete page
  7. Initialize your page with the necessary data
  8. Render the page by calling its View method
*/

type PageUI interface {
	AddHeader(string)
	AddBody(string)
	AddRow(string)
	AddColumn(string)
	AddControls(string)
	AddStatus(string)
	Render(int) string
}

type BasePage struct {
	Title    string
	Header   string
	Rows     []Row
	Controls string
	Status   string
}

func NewBasePage(title string, header string, rows []Row, controls string, status string) BasePage {
	return BasePage{
		Title:    title,
		Header:   header,
		Rows:     rows,
		Controls: controls,
		Status:   status,
	}
}

type StaticPage struct {
	BasePage
}

func (s *StaticPage) AddHeader(h string) {
	s.Header += h
}
func (s *StaticPage) AddRow(r Row) {
	s.Rows = append(s.Rows, r)
}
func (s *StaticPage) AddControls(c string) {
	s.Controls += c
}
func (s *StaticPage) AddStatus(st string) {
	s.Status += st
}
func (s *StaticPage) RenderBody() string {
	r := make([]string, len(s.Rows))
	for _, v := range s.Rows {
		r = append(r, v.Build())
	}
	return lipgloss.JoinVertical(lipgloss.Left, r...)
}

func (s StaticPage) Render(model Model) string {
	docStyle := lipgloss.NewStyle().Padding(1, 2, 1, 2)
	rows := []string{RenderTitle(s.Title), RenderHeading(s.Header), s.RenderBody()}
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

func NewStaticPage(title string, header string, rows []Row, controls string, status string) StaticPage {
	page := BasePage{
		Title:    title,
		Header:   header,
		Rows:     rows,
		Controls: controls,
		Status:   status,
	}
	return StaticPage{
		BasePage: page,
	}

}

type MenuPage struct {
	BasePage
	Menu []string
}

func (m *MenuPage) AddHeader(h string) {
	m.Header += h
}
func (m *MenuPage) AddRow(r Row) {
	m.Rows = append(m.Rows, r)
}
func (m *MenuPage) AddControls(c string) {
	m.Controls += c
}
func (m *MenuPage) AddStatus(st string) {
	m.Status += st
}
func (m *MenuPage) RenderBody(model Model) string {
	r := make([]string, len(m.Rows)+len(m.Menu))
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
	r = append(r, lipgloss.JoinHorizontal(lipgloss.Top, row...))
	for _, v := range m.Rows {
		r = append(r, v.Build())
	}
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

func NewMenuPage(m []string, title string, header string, body []Row, controls string, status string) MenuPage {
	basePage := NewBasePage(title, header, body, controls, status)
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

func (t *TablePage) AddHeader(h string) {
	t.Header += h
}

func (t *TablePage) AddControls(c string) {
	t.Controls += c
}

func (t *TablePage) AddStatus(st string) {
	t.Status += st
}

func (t *TablePage) RenderBody(m Model) string {
	start, end := m.Paginator.GetSliceBounds(len(t.TableRows))
	currentView := t.TableRows[start:end]

	t.TableUI = StyledTable(t.TableHeaders, currentView, m.Cursor)
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

func NewTablePage(headers []string, rows [][]string, table string, title string, header string, body []Row, controls string, status string) TablePage {
	basePage := NewBasePage(title, header, body, controls, status)
	return TablePage{
		BasePage:     basePage,
		Table:        table,
		TableHeaders: headers,
		TableRows:    rows,
	}
}

type FormPage struct {
	BasePage
	Form *huh.Form
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

func NewFormPage(title string, header string, body []Row, controls string, status string) FormPage {
	basePage := NewBasePage(title, header, body, controls, status)
	return FormPage{
		BasePage: basePage,
	}
}

type DisplayMode int

const (
	Main DisplayMode = iota
	SelectDatatype
	Options
)

type CMSPage struct {
	BasePage
	Tree    TreeRoot
	Display DisplayMode
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

func (c CMSPage) Render(model Model) string {
	docStyle := lipgloss.NewStyle().Padding(1, 2, 1, 2)
	s := lipgloss.JoinVertical(
		lipgloss.Left,
		RenderTitle(c.Title),
		RenderHeading(c.Header),
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

func NewCMSPage(title string, header string, body []Row, controls string, status string) CMSPage {
	b := NewBasePage(title, header, body, controls, status)
	p := CMSPage{
		BasePage: b,
		Tree:     TreeRoot{},
	}

	return p
}

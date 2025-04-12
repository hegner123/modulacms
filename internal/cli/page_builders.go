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

func NewBasePage(t string, h string, r []Row, c string, s string) BasePage {
	return BasePage{
		Title:    t,
		Header:   h,
		Rows:     r,
		Controls: c,
		Status:   s,
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

func (s StaticPage) Render(c int) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		s.Title,
		s.Header,
		s.RenderBody(),
		s.Controls,
		s.Status,
	)
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
		if model.cursor == i {
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

func (t *TablePage) RenderBody(m *Model) string {
	start, end := m.paginator.GetSliceBounds(len(t.TableRows))
	currentView := t.TableRows[start:end]

	t.TableUI = StyledTable(t.TableHeaders, currentView, m.cursor)
	paginator := ""
	if len(t.TableRows) > m.maxRows {
		paginator = "\n\n" + m.paginator.View()
	}
	b := lipgloss.JoinVertical(
		lipgloss.Top,
		t.TableUI.Render(),
		paginator,
	)
	return b
}

func (t TablePage) Render(model *Model) string {
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
	return lipgloss.JoinVertical(
		lipgloss.Left,
		docStyle.Render(s),
		h,
		f,
		status,
	)
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

type CMSPage struct {
	BasePage
	Datatype string
	Tree     any
}

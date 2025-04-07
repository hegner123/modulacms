package cli

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
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
	Menu []*Page
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
func (m *MenuPage) RenderBody(c int) string {
	r := make([]string, len(m.Rows)+len(m.Menu))
	for i, v := range m.Menu {
		if i == c {
			r = append(r, "->"+v.Label)
		} else {
			r = append(r, "  "+v.Label)
		}
	}
	for _, v := range m.Rows {
		r = append(r, v.Build())
	}
	return RenderBorderFixed(lipgloss.JoinVertical(lipgloss.Left, r...))
}

func (m MenuPage) Render(c int) string {
	docStyle := lipgloss.NewStyle().Padding(1, 2, 1, 2)
	return docStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		RenderTitle(m.Title),
		RenderHeading(m.Header),
		m.RenderBody(c),
		RenderFooter(m.Controls),
		m.Status,
	),
	)
}

func NewMenuPage(m []*Page, title string, header string, body []Row, controls string, status string) MenuPage {
	basePage := NewBasePage(title, header, body, controls, status)
	return MenuPage{
		BasePage: basePage,
		Menu:     m,
	}
}

type TablePage struct {
	BasePage
	Table   string
	Headers []string
	Rows    [][]string
	PageMod int
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

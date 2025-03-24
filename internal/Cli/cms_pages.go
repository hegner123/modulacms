package cli

import "fmt"

var (
	defineDatatype *Page = &Page{Index: DEFINEDATATYPE, Controller: pageInterface, Label: "Define Datatype", Parent: nil, Children: nil}
	datatypeMenu *Page = &Page{Index: DEFINEDATATYPE, Controller: pageInterface, Label: "Define Datatype", Parent: nil, Children: nil}
)

func (m model) PageDefineDatatype() string {
	m.header = "Define Datatype"
	for i, choice := range m.pageMenu {

		cursor := "  "
		if m.cursor == i {
			cursor = "->"
		}

		m.body += fmt.Sprintf("%s%s  \n", cursor, choice.Label)
	}
	m.body = RenderBorder(m.body)
	return m.RenderUI()
}

func (m model) PageDatatypeMenu() string {
	m.header = "Datatypes"
	for i, choice := range m.datatypeMenu {

		cursor := "  "
		if m.cursor == i {
			cursor = "->"
		}

		m.body += fmt.Sprintf("%s%s  \n", cursor, choice)
	}
	m.body = RenderBorder(m.body)
	return m.RenderUI()
}

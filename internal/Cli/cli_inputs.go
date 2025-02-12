package cli

import "fmt"

var (
	txtarea InputType = "textarea"
)

func (m model) PageEditor(page string) string {
	m.header += fmt.Sprintf("ModulaCMS\n\nEdit %s \n", page)
	for i := range m.Options {

		for _, option := range m.Options[i].List {
			m.body += m.InputUI(option)
		}
	}
	return m.RenderUI()

}

func (m model) InputUI(o Option) string {
	var view string
	switch o.InputType {
	case txtarea:
		view = m.textarea.View()

	}
	return view
}

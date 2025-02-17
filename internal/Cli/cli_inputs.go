package cli

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
)

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

func (m model) InputInit(numInputs int) model {
	for i := 0; i < numInputs; i++ {
		ti := textinput.New()
		ti.Placeholder = fmt.Sprintf("Input %d", i+1)
		ti.CharLimit = 32
		if i == 0 {
			ti.Focus()
		} else {
			ti.Blur()
		}
		m.textInputs[i] = ti
	}
	return m
}

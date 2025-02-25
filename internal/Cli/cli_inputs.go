package cli

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
)

var (
	textInput InputType = "textInput"
	textArea  InputType = "textArea"
	fileInput InputType = "fileInput"
)

type Input struct {
	Index        int
	Format       InputType
	Key          any
	Value        any
	CurrentIndex int
	History      []any
}

func (i Input) View() string {
	var r string
	switch i.Format {
	case textInput:
	case textArea:
	case fileInput:

	}
	return r
}

func (i *Input) Update(newValue any) {
	i.History = i.History[:i.CurrentIndex+1]
	i.History = append(i.History, newValue)
	i.CurrentIndex++
	i.Value = newValue
}

func (i *Input) Redo() {
	if i.CurrentIndex+1 >= len(i.History) {
		return
	}
	i.CurrentIndex++
	i.Value = i.History[i.CurrentIndex]
}

func (i *Input) Undo() {
	if i.CurrentIndex-1 < 0 {
		return
	}
	i.CurrentIndex--
	i.Value = i.History[i.CurrentIndex]
}

func (m model) PageEditor(page string) string {
	m.header += fmt.Sprintf("ModulaCMS\n\nEdit %s \n", page)
	for i := range m.Options {

		for _, option := range m.Options[i].List {
			m.body += m.TextAreaView(option)
		}
	}
	return m.RenderUI()

}

func (m model) TextAreaView(o Option) string {
	var view string
	switch o.InputType {
	case textArea:
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

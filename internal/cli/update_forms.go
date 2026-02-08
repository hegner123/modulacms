package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type UpdatedForm struct{}

func NewUpdatedForm() tea.Cmd {
	return func() tea.Msg {
		return UpdatedForm{}
	}
}

func (m Model) UpdateForm(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	case NewFormMsg:
		return m, tea.Batch(
			LoadingStartCmd(),
			SetFormDataCmd(*msg.Form, msg.FieldsCount, msg.Values, msg.FormMap),
			NavigateToPageCmd(m.PageMap[CREATEPAGE]),
		)
	case CmsBuildDefineDatatypeFormMsg:
		form, count, values := NewDefineDatatypeForm(m, false)
		return m, tea.Batch(
			SetFormDataCmd(*form, count, values, nil),
			NavigateToPageCmd(m.PageMap[DATATYPE]),
		)
	case CmsEditDatatypeFormMsg:
		form, count, values := NewEditDatatypeForm(m, msg.Datatype)
		return m, tea.Batch(
			SetFormDataCmd(*form, count, values, nil),
			NavigateToPageCmd(m.PageMap[DATATYPE]),
		)
	case FormSubmitMsg:
		newModel := m
		newModel.FormState.FormSubmit = true
		return newModel, nil
	case FormInitOptionsMsg:
		newModel := m
		if newModel.FormState.FormOptions == nil {
			newOptions := make(map[string][]huh.Option[string], 0)
			newModel.FormState.FormOptions = (*FormOptionsMap)(&newOptions)
		}
		fo := *newModel.FormState.FormOptions
		if fo[msg.Form] != nil {
			return m, nil
		}

		newOptionsSet := make([]huh.Option[string], 0)
		fo[msg.Form] = newOptionsSet
		return newModel, NewUpdatedForm()
	case DbResMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Database operation completed for table %s", msg.Table)),
		)
	}

	return m, nil
}

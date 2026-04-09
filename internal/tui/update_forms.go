package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// UpdatedForm signals that a form has been updated.
type UpdatedForm struct{}

// NewUpdatedForm creates a command returning an UpdatedForm message.
func NewUpdatedForm() tea.Cmd {
	return func() tea.Msg {
		return UpdatedForm{}
	}
}

// UpdateForm handles form-related messages and state updates.
func (m Model) UpdateForm(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	case FormSubmitMsg:
		newModel := m
		newModel.FormState.FormSubmit = true
		return newModel, NewStateUpdate()
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
			LogMessageCmd(fmt.Sprintf("database operation completed for table %s", msg.Table)),
		)
	}

	return m, nil
}

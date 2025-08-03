package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
)

type UpdatedForm struct{}

func NewUpdatedForm() tea.Cmd {
	return func() tea.Msg {
		return UpdatedForm{}
	}

}

func (m Model) UpdateForm(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	case FormCreate:
		switch msg.FormType {
		case DATABASECREATE:
			if m.Columns == nil {
				return m, tea.Batch(
					LogMessageCmd(fmt.Sprintf("Form creation failed: no columns available for table %s", m.Table)),
				)
			}
			var cmds []tea.Cmd
			c := m.NewInsertForm(db.DBTable(m.Table))
			cmds = append(cmds, c)
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, LogMessageCmd(fmt.Sprintf("Database create form initialized for table %s with %d fields", m.Table, len(*m.Columns)-1)))
			return m, tea.Batch(cmds...)
		}
	case FormLenSet:
		newModel := m
		newModel.FormLen = msg.FormLen
		return newModel, NewUpdatedForm()
	case FormSet:
		newModel := m
		newModel.Form = &msg.Form
		newModel.FormValues = msg.Values
		return newModel, NewUpdatedForm()
	case FormValuesSet:
		newModel := m
		newModel.FormValues = msg.Values
		return newModel, NewUpdatedForm()
	case NewFormMsg:
		return m, tea.Batch(
			LoadingStartCmd(),
			SetFormDataCmd(*msg.Form, msg.FieldsCount, msg.Values),
		)
	case FormSubmitMsg:
		newModel := m
		newModel.FormSubmit = true
		return newModel, tea.Batch()
	case FormActionMsg:
		switch msg.Action {
		case INSERT:
			return m, tea.Batch(
				LogMessageCmd(fmt.Sprintf("Processing %s action for table %s", msg.Action, msg.Table)),
				DatabaseInsertCmd(db.DBTable(msg.Table), msg.Columns, msg.Values),
			)

		}
	case DbResMsg:
		return m, tea.Batch(
			LogMessageCmd(fmt.Sprintf("Database operation completed for table %s", msg.Table)),
		)
	}

	return m, nil
}

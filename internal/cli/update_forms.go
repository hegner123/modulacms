package cli

import (
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
					LogMessage("Columns Empty"),
				)
			}
			var cmds []tea.Cmd
			c := m.BuildCreateDBForm(db.DBTable(m.Table))
			cmds = append(cmds, c)
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, LogMessage("NewForm Exit"))
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
		return m, SetFormDataCmd(*msg.Form, msg.FieldsCount, msg.Values)
	}

	return m, nil
}

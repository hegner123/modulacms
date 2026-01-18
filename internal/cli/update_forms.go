package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
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
			if m.TableState.Columns == nil {
				return m, LogMessageCmd(fmt.Sprintf("Form creation failed: no columns available for table %s", m.TableState.Table))
			}
			var cmds []tea.Cmd
			c := m.NewInsertForm(db.DBTable(m.TableState.Table))
			cmds = append(cmds, c)
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, LogMessageCmd(fmt.Sprintf("Database create form initialized for table %s with %d fields", m.TableState.Table, len(*m.TableState.Columns)-1)))
			return m, tea.Batch(cmds...)
		}
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
			NavigateToPageCmd(m.PageMap[DATATYPES]),
		)
	case FormSubmitMsg:
		newModel := m
		newModel.FormState.FormSubmit = true
		return newModel, nil
	case FormActionMsg:
		switch msg.Action {
		case INSERT:
			filteredColumns := make([]string, 0)
			filteredValues := make([]*string, 0)

			for i, value := range msg.Values {
				if value != nil && *value != "" {
					filteredColumns = append(filteredColumns, msg.Columns[i])
					filteredValues = append(filteredValues, value)
				} else {
					filteredColumns = append(filteredColumns, msg.Columns[i])
					filteredValues = append(filteredValues, nil)

				}
			}
			return m, tea.Batch(
				DatabaseInsertCmd(db.DBTable(msg.Table), filteredColumns, filteredValues),
				LogMessageCmd(fmt.Sprintln(filteredColumns)),
				LogMessageCmd(fmt.Sprintln(filteredValues)),
			)

		}
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
		return m, LogMessageCmd(fmt.Sprintf("Database operation completed for table %s", msg.Table))
	}

	return m, nil
}

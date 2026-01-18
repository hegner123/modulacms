package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/db"
)

type NewFormMsg struct {
	Form        *huh.Form
	FieldsCount int
	Values      []*string
}

func (m Model) NewInsertForm(table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		var fields []huh.Field
		var values []*string
		for i, c := range *m.TableState.Columns {
			blank := ""
			if i == 0 {
				values = append(values, &blank)
				continue
			} else {
				value := ""
				t := *m.TableState.ColumnTypes
				f, err := m.NewFieldFromType(m.Config, c, t[i], &value)
				if err != nil {
					return FetchErrMsg{Error: err}
				}
				if f == nil {
					continue
				}
				fields = append(fields, f)
				values = append(values, &value)
			}

		}
		group := huh.NewGroup(fields...)
		form := huh.NewForm(
			group,
		)
		form.Init() // Initialize immediately
		// Add submit handler with proper focus management
		form.SubmitCmd = tea.Batch(
			LogMessageCmd(fmt.Sprintf("Form SubmitCmd triggered for INSERT on table %s", string(table))),
			LogMessageCmd(fmt.Sprintf("Headers  %v", m.TableState.Columns)),
			FormActionCmd(INSERT, string(table), *m.TableState.Columns, values),
			FocusSetCmd(PAGEFOCUS),
			func() tea.Msg {
				return tea.ResumeMsg{}
			},
		)
		return NewFormMsg{Form: form, FieldsCount: len(*m.TableState.Columns), Values: values}
	}

}

func (m *Model) NewUpdateForm(table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		row := *m.TableState.Row
		var fields []huh.Field
		for i, c := range *m.TableState.Columns {
			if i == 0 {
				id := row[i]
				m.FormState.FormValues = append(m.FormState.FormValues, &id)
				continue
			} else {

				value := row[i]
				t := *m.TableState.ColumnTypes
				f, err := m.NewUpdateFieldFromType(m.Config, c, t[i], &value, row[i])
				if err != nil {
					return FetchErrMsg{Error: err}
				}
				if f == nil {
					continue
				}

				fields = append(fields, f)
				m.FormState.FormValues = append(m.FormState.FormValues, &value)
			}

		}

		m.FormState.FormFields = fields

		group := huh.NewGroup(fields...)
		m.FormState.FormGroups = []huh.Group{*group}

		form := huh.NewForm(
			group,
		)
		form.Init() // Initialize immediately

		// Add submit handler with proper focus management
		form.SubmitCmd = tea.Batch(
			LogMessageCmd(fmt.Sprintf("Form SubmitCmd triggered for UPDATE on table %s", string(table))),
			FormActionCmd(UPDATE, string(table), m.TableState.Headers, m.FormState.FormValues),
			FocusSetCmd(PAGEFOCUS),
			func() tea.Msg {
				return tea.ResumeMsg{}
			},
		)
		return NewFormMsg{Form: form, FieldsCount: len(*m.TableState.Columns)}
	}
}

func (m *Model) BuildCMSForm(table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		var fields []huh.Field
		for i, c := range *m.TableState.Columns {
			if i == 0 {
				continue
			}
			var value string
			t := *m.TableState.ColumnTypes
			f, err := m.NewFieldFromType(m.Config, c, t[i], &value)
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if f == nil {
				continue
			}
			fields = append(fields, f)
			m.FormState.FormValues = append(m.FormState.FormValues, &value)

		}
		group := huh.NewGroup(fields...)
		// Create form with both groups
		form := huh.NewForm(
			group,
		)
		form.Init() // Initialize immediately

		// Add submit handler with proper focus management
		form.SubmitCmd = func() tea.Msg {
			if m.FormState.FormSubmit {
				m.Focus = PAGEFOCUS
				return FormActionMsg{}
			}
			return FormCancelMsg{}
		}
		return NewFormMsg{Form: form, FieldsCount: len(*m.TableState.Columns)}
	}

}

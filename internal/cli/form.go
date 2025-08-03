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
		for i, c := range *m.Columns {
			blank := ""
			if i == 0 {
				values = append(values, &blank)
				continue
			} else {
				value := ""
				t := *m.ColumnTypes
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
			LogMessageCmd(fmt.Sprintf("Headers  %v", m.Columns)),
			FormActionCmd(INSERT, string(table), *m.Columns, values),
			FocusSetCmd(PAGEFOCUS),
			func() tea.Msg {
				return tea.ResumeMsg{}
			},
		)
		return NewFormMsg{Form: form, FieldsCount: len(*m.Columns), Values: values}
	}

}

func (m *Model) NewUpdateForm(table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		row := *m.Row
		var fields []huh.Field
		for i, c := range *m.Columns {
			if i == 0 {
				id := row[i]
				m.FormValues = append(m.FormValues, &id)
				continue
			} else {

				value := row[i]
				t := *m.ColumnTypes
				f, err := m.NewUpdateFieldFromType(m.Config, c, t[i], &value, row[i])
				if err != nil {
					return FetchErrMsg{Error: err}
				}
				if f == nil {
					continue
				}

				fields = append(fields, f)
				m.FormValues = append(m.FormValues, &value)
			}

		}

		m.FormFields = fields

		group := huh.NewGroup(fields...)
		m.FormGroups = []huh.Group{*group}

		form := huh.NewForm(
			group,
		)
		form.Init() // Initialize immediately

		// Add submit handler with proper focus management
		form.SubmitCmd = tea.Batch(
			LogMessageCmd(fmt.Sprintf("Form SubmitCmd triggered for UPDATE on table %s", string(table))),
			FormActionCmd(UPDATE, string(table), m.Headers, m.FormValues),
			FocusSetCmd(PAGEFOCUS),
			func() tea.Msg {
				return tea.ResumeMsg{}
			},
		)
		return NewFormMsg{Form: form, FieldsCount: len(*m.Columns)}
	}
}

func (m *Model) BuildCMSForm(table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		var fields []huh.Field
		for i, c := range *m.Columns {
			if i == 0 {
				continue
			}
			var value string
			t := *m.ColumnTypes
			f, err := m.NewFieldFromType(m.Config, c, t[i], &value)
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if f == nil {
				continue
			}
			fields = append(fields, f)
			m.FormValues = append(m.FormValues, &value)

		}
		group := huh.NewGroup(fields...)
		// Create form with both groups
		form := huh.NewForm(
			group,
		)
		form.Init() // Initialize immediately

		// Add submit handler with proper focus management
		form.SubmitCmd = func() tea.Msg {
			if m.FormSubmit {
				m.Focus = PAGEFOCUS
				return FormActionMsg{}
			}
			return FormCancelMsg{}
		}
		form.SubmitCmd = func() tea.Msg {
			m.Focus = PAGEFOCUS
			return tea.ResumeMsg{}
		}
		return NewFormMsg{Form: form, FieldsCount: len(*m.Columns)}
	}

}

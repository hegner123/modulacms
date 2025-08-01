package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/db"
)

type createFormMsg struct {
	Form        *huh.Form
	FieldsCount int
}

type updateFormMsg struct {
	Form        *huh.Form
	FieldsCount int
}

type cmsFormMsg struct {
	Form        *huh.Form
	FieldsCount int
}

func (m *Model) BuildCreateDBForm(table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		var fields []huh.Field
		for i, c := range *m.Columns {
			blank := ""
			if i == 0 {
				m.FormValues = append(m.FormValues, &blank)
				continue
			}
			value := ""
			t := *m.ColumnTypes
			f, err := m.NewFieldFromType(m.Config, c, t[i], &value)
			if err != nil {
				return ErrMsg{Error: err}
			}
			if f == nil {
				continue
			}
			fields = append(fields, f)
			m.FormValues = append(m.FormValues, &value)

		}
		group := huh.NewGroup(fields...)
		form := huh.NewForm(
			group,
		)

		// Add submit handler with proper focus management
		form.SubmitCmd = func() tea.Msg {
			if m.FormSubmit {
				return formCompletedMsg{}
			}
			return formCancelledMsg{}
		}
		form.SubmitCmd = func() tea.Msg {
			return tea.ResumeMsg{}
		}
		return createFormMsg{Form: form, FieldsCount: len(*m.Columns)}
	}

}

func (m *Model) BuildUpdateDBForm(table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		row := *m.Row
		var fields []huh.Field
		for i, c := range *m.Columns {
			if i == 0 {
				id := row[i]
				m.FormValues = append(m.FormValues, &id)
				continue
			}
			value := row[i]
			t := *m.ColumnTypes
			f, err := m.NewUpdateFieldFromType(m.Config, c, t[i], &value, row[i])
			if err != nil {
				return ErrMsg{Error: err}
			}
			if f == nil {
				continue
			}

			fields = append(fields, f)
			m.FormValues = append(m.FormValues, &value)

		}

		m.FormFields = fields

		group := huh.NewGroup(fields...)
		m.FormGroups = []huh.Group{*group}

		form := huh.NewForm(
			group,
		)

		// Add submit handler with proper focus management
		form.SubmitCmd = func() tea.Msg {
			if m.FormSubmit {
				m.Focus = PAGEFOCUS
				return formCompletedMsg{}
			}
			return formCancelledMsg{}
		}
		form.SubmitCmd = func() tea.Msg {
			m.Focus = PAGEFOCUS
			return tea.ResumeMsg{}
		}
		return updateFormMsg{Form: form, FieldsCount: len(*m.Columns)}
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
				return ErrMsg{Error: err}
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

		// Add submit handler with proper focus management
		form.SubmitCmd = func() tea.Msg {
			if m.FormSubmit {
				m.Focus = PAGEFOCUS
				return formCompletedMsg{}
			}
			return formCancelledMsg{}
		}
		form.SubmitCmd = func() tea.Msg {
			m.Focus = PAGEFOCUS
			return tea.ResumeMsg{}
		}
		return updateFormMsg{Form: form, FieldsCount: len(*m.Columns)}
	}

}

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
		for i, c := range *m.columns {
			blank := ""
			if i == 0 {
				m.formValues = append(m.formValues, &blank)
				continue
			}
			value := ""
			t := *m.columnTypes
			f, err := m.NewFieldFromType(m.config, c, t[i], &value)
			if err != nil {
				return ErrMsg{Error: err}
			}
			if f == nil {
				continue
			}
			fields = append(fields, f)
			m.formValues = append(m.formValues, &value)

		}
		group := huh.NewGroup(fields...)
		form := huh.NewForm(
			group,
		)

		// Add submit handler with proper focus management
		form.SubmitCmd = func() tea.Msg {
			if m.formSubmit {
				return formCompletedMsg{}
			}
			return formCancelledMsg{}
		}
		form.SubmitCmd = func() tea.Msg {
			return tea.ResumeMsg{}
		}
		return createFormMsg{Form: form, FieldsCount: len(*m.columns)}
	}

}

func (m *Model) BuildUpdateDBForm(table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		row := *m.row
		var fields []huh.Field
		for i, c := range *m.columns {
			if i == 0 {
				id := row[i]
				m.formValues = append(m.formValues, &id)
				continue
			}
			value := row[i]
			t := *m.columnTypes
			f, err := m.NewUpdateFieldFromType(m.config, c, t[i], &value, row[i])
			if err != nil {
				return ErrMsg{Error: err}
			}
			if f == nil {
				continue
			}

			fields = append(fields, f)
			m.formValues = append(m.formValues, &value)

		}

		m.formFields = fields

		group := huh.NewGroup(fields...)
		m.formGroups = []huh.Group{*group}

		form := huh.NewForm(
			group,
		)

		// Add submit handler with proper focus management
		form.SubmitCmd = func() tea.Msg {
			if m.formSubmit {
				m.focus = PAGEFOCUS
				return formCompletedMsg{}
			}
			return formCancelledMsg{}
		}
		form.SubmitCmd = func() tea.Msg {
			m.focus = PAGEFOCUS
			return tea.ResumeMsg{}
		}
		return updateFormMsg{Form: form, FieldsCount: len(*m.columns)}
	}
}

func (m *Model) BuildCMSForm(table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		var fields []huh.Field
		for i, c := range *m.columns {
			if i == 0 {
				continue
			}
			var value string
			t := *m.columnTypes
			f, err := m.NewFieldFromType(m.config, c, t[i], &value)
			if err != nil {
				return ErrMsg{Error: err}
			}
			if f == nil {
				continue
			}
			fields = append(fields, f)
			m.formValues = append(m.formValues, &value)

		}
		group := huh.NewGroup(fields...)
		// Create form with both groups
		form := huh.NewForm(
			group,
		)

		// Add submit handler with proper focus management
		form.SubmitCmd = func() tea.Msg {
			if m.formSubmit {
				m.focus = PAGEFOCUS
				return formCompletedMsg{}
			}
			return formCancelledMsg{}
		}
		form.SubmitCmd = func() tea.Msg {
			m.focus = PAGEFOCUS
			return tea.ResumeMsg{}
		}
		return updateFormMsg{Form: form, FieldsCount: len(*m.columns)}
	}

}

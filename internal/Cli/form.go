package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	db "github.com/hegner123/modulacms/internal/Db"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

var ErrLog utility.Logger = *utility.NewLogger(utility.ERROR)

func (m *model) BuildCreateDBForm(table db.DBTable) (*huh.Form, int) {
	columns, colType, err := GetColumns(string(table))
	if err != nil {
		return nil, 0
	}
	var fields []huh.Field
	for i, c := range *columns {
		blank := ""
		if i == 0 {
			m.formValues = append(m.formValues, &blank)
			continue
		}
		var value string
		t := *colType
		f, err := m.NewFieldFromType(c, t[i], &value)
		if err != nil {
			return nil, 0
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
	return form, len(*columns)

}

func (m *model) BuildUpdateDBForm(table db.DBTable) (*huh.Form, int) {
	row := *m.row
	columns, colType, err := GetColumns(string(table))
	if err != nil {
		return nil, 0
	}
	var fields []huh.Field
	for i, c := range *columns {
		value := row[i]
		t := *colType
		f, err := m.NewUpdateFieldFromType(c, t[i], &value, row[i])
		if err != nil {
			return nil, 0
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
	return form, len(*columns)
}

func (m *model) BuildCMSForm(table db.DBTable) (*huh.Form, int) {
	columns, colType, err := GetColumns(string(table))
	if err != nil {
		return nil, 0
	}
	var fields []huh.Field
	for i, c := range *columns {
		if i == 0 {
			continue
		}
		var value string
		t := *colType
		f, err := m.NewFieldFromType(c, t[i], &value)
		if err != nil {
			return nil, 0
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
	return form, len(*columns)

}

package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/db"
)


func CreateDatatypeForm(m Model) (*huh.Form, int) {
	var (
		parentID int
		label    string
		dType    string
	)
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Parent").
				OptionsFunc(
					func() []huh.Option[int] {
						options := make([]huh.Option[int], 0)
						d := db.ConfigDB(*m.GetConfig())
						r, err := d.ListDatatypes()
						if err != nil {
							e := fmt.Errorf("error listing datatypes %w", err)
							m.SetError(e)
						}
						for _, v := range *r {
							option := huh.NewOption(v.Label, int(v.DatatypeID))
							options = append(options, option)
						}
						return options
					},
					nil,
				).
				Value(&parentID), // st
			huh.NewInput().
				Title("Label").
				Value(&label),
			huh.NewInput().
				Title("Type").
				Value(&dType),
		))

	return form, 3 // 3 fields
}
func CreateFieldForm(m Model) (*huh.Form, int) {
	var (
		parentID int
		label    string
		data     *[]string
		dType    string
	)
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Parent").
				OptionsFunc(
					func() []huh.Option[int] {
						options := make([]huh.Option[int], 0)
						d := db.ConfigDB(*m.GetConfig())
						r, err := d.ListDatatypes()
						if err != nil {
							e := fmt.Errorf("error listing datatypes %w", err)
							m.SetError(e)
						}
						for _, v := range *r {
							option := huh.NewOption(v.Label, int(v.DatatypeID))
							options = append(options, option)
						}
						return options
					},
					nil,
				).
				Value(&parentID),
			huh.NewMultiSelect[string]().Title("Options").Options(
				huh.NewOption("Required", "required"),
				huh.NewOption("Validation", "Validation"),
			).Value(data),
			huh.NewInput().
				Title("Label").
				Value(&label),
			huh.NewInput().
				Title("Type").
				Value(&dType),
		))

	return form, 4 // 4 fields
}

// Using the existing cmsFormMsg from form.go

// BuildCMSDatatypeForm creates a form for CMS datatypes
func (m Model) BuildCMSDatatypeForm() tea.Cmd {
	return func() tea.Msg {
		form, count := CreateDatatypeForm(m)
		return cmsFormMsg{
			Form:        form,
			FieldsCount: count,
		}
	}
}

// BuildCMSFieldForm creates a form for CMS fields
func (m Model) BuildCMSFieldForm() tea.Cmd {
	return func() tea.Msg {
		form, count := CreateFieldForm(m)
		return cmsFormMsg{
			Form:        form,
			FieldsCount: count,
		}
	}
}

// CMSFormControls handles the UI controls for CMS forms
func (m Model) CMSFormControls(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	m.Focus = FORMFOCUS

	form, cmd := m.Form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.Form = f
		cmds = append(cmds, cmd)
	}

	if m.Form.State == huh.StateAborted {
		_ = tea.ClearScreen()
		m.Focus = PAGEFOCUS
		m.Page = m.Pages[CMSPAGE]
	}

	if m.Form.State == huh.StateCompleted {
		_ = tea.ClearScreen()
		m.Focus = PAGEFOCUS
		// Process completed form (implementation would depend on form type)
		m.Page = m.Pages[CMSPAGE]
	}
	
	return m, tea.Batch(cmds...)
}

package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/db"
)

type FormIndex int

const (
	DATABASECREATE FormIndex = iota
	DATABASEUPDATE
	CMSCREATE
	CMSUPDATE
)

// TODO add argument for admin / client specific action
func NewDefineDatatypeForm(m Model, admin bool) (*huh.Form, int, []*string) {
	values := make([]*string, 8)
	columns := []string{
		"datatype_id",
		"parent_id",
		"label",
		"type",
		"author_id",
		"date_created",
		"date_modified",
		"history",
	}
	var (
		parent   string
		label    string
		datatype string = "ROOT"
	)
	groupDescription := "Define datatype"
	typeDescription := "Optional - ROOT is reserved for root content types.\n"
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Parent").
				OptionsFunc(func() []huh.Option[string] {
					options := make([]huh.Option[string], 0)
					dbc := db.ConfigDB(*m.Config)
					rows, err := dbc.ListDatatypes()
					if err != nil {
						return options
					}
					blankOption := huh.Option[string]{
						Key:   "ROOT",
						Value: "",
					}
					options = append(options, blankOption)
					for _, v := range *rows {
						option := huh.Option[string]{
							Key:   v.Label,
							Value: fmt.Sprint(v.DatatypeID),
						}
						options = append(options, option)
					}
					return options
				}, nil).
				Value(&parent),
			huh.NewInput().
				Title("Label").
				Description("Display name for this content type").
				Value(&label),
			huh.NewInput().
				Title("Type").
				Description(typeDescription).
				Value(&datatype),
		).Description(groupDescription),
	)
	values[1] = &parent
	values[2] = &label
	values[3] = &datatype
	form.Init()
	table := "datatypes"
	if admin {
		table = "admin_datatypes"
	}
	form.SubmitCmd = tea.Batch(
		LogMessageCmd(fmt.Sprintf("Form SubmitCmd triggered for INSERT on table %s", string(table))),
		FormActionCmd(INSERT, table, columns, values),
		FocusSetCmd(PAGEFOCUS),
		func() tea.Msg {
			return tea.ResumeMsg{}
		},
	)

	return form, len(values), values
}

// Field Form for adding fields to a Datatype

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

// BuildCMSFieldForm creates a form for CMS fields
func (m Model) BuildCMSFieldForm() tea.Cmd {
	return func() tea.Msg {
		form, count := CreateFieldForm(m)
		return NewFormMsg{
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

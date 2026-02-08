package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
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
		datatype = "ROOT"
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

func NewEditDatatypeForm(m Model, dt db.Datatypes) (*huh.Form, int, []*string) {
	values := make([]*string, 8)

	var (
		parent   string
		label    = dt.Label
		datatype = dt.Type
	)
	if dt.ParentID.Valid {
		parent = dt.ParentID.ID.String()
	}

	groupDescription := "Edit datatype"
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

	datatypeID := dt.DatatypeID
	form.SubmitCmd = tea.Batch(
		LogMessageCmd(fmt.Sprintf("Edit datatype SubmitCmd triggered for %s", datatypeID)),
		func() tea.Msg {
			return DatatypeUpdateSaveMsg{
				DatatypeID: datatypeID,
				Parent:     *values[1],
				Label:      *values[2],
				Type:       *values[3],
			}
		},
		FocusSetCmd(PAGEFOCUS),
		func() tea.Msg {
			return tea.ResumeMsg{}
		},
	)

	return form, len(values), values
}

// Field Form for adding fields to a Datatype

func CreateDatatypeForm(m Model) (*huh.Form, int) {
	logger := m.Logger
	var (
		parentID string
		label    string
		dType    string
	)
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Parent").
				OptionsFunc(
					func() []huh.Option[string] {
						options := make([]huh.Option[string], 0)
						d := db.ConfigDB(*m.GetConfig())
						r, err := d.ListDatatypes()
						if err != nil {
							logger.Error("error listing datatypes %w", err)
						}
						for _, v := range *r {
							option := huh.NewOption(v.Label, string(v.DatatypeID))
							options = append(options, option)
						}
						return options
					},
					nil,
				).
				Value(&parentID),
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
	logger := m.Logger
	var (
		parentID string
		label    string
		data     *[]string
		dType    string
	)
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Parent").
				OptionsFunc(
					func() []huh.Option[string] {
						options := make([]huh.Option[string], 0)
						d := db.ConfigDB(*m.GetConfig())
						r, err := d.ListDatatypes()
						if err != nil {
							logger.Error("error listing datatypes %w", err)
						}
						for _, v := range *r {
							option := huh.NewOption(v.Label, string(v.DatatypeID))
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

	form, cmd := m.FormState.Form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.FormState.Form = f
		cmds = append(cmds, cmd)
	}

	if m.FormState.Form.State == huh.StateAborted {
		_ = tea.ClearScreen()
		m.Focus = PAGEFOCUS
		m.Page = m.Pages[CMSPAGE]
	}

	if m.FormState.Form.State == huh.StateCompleted {
		_ = tea.ClearScreen()
		m.Focus = PAGEFOCUS
		// Process completed form (implementation would depend on form type)
		m.Page = m.Pages[CMSPAGE]
	}

	return m, tea.Batch(cmds...)
}

// Database form builders

type NewFormMsg struct {
	Form        *huh.Form
	FieldsCount int
	Values      []*string
	FormMap     []string
}


// BuildContentFieldsForm creates a dynamic form for content creation based on datatype fields
func (m Model) BuildContentFieldsForm(datatypeID types.DatatypeID, routeID types.RouteID) tea.Cmd {
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}
	return func() tea.Msg {
		d := db.ConfigDB(*m.Config)

		logger.Finfo(fmt.Sprintf("Building content form for datatype %s, route %s", datatypeID, routeID))

		// Fetch field IDs from the datatypes_fields join table
		dtID := types.NullableDatatypeID{ID: datatypeID, Valid: true}
		dtFields, err := d.ListDatatypeFieldByDatatypeID(dtID)
		if err != nil {
			logger.Ferror("ListDatatypeFieldByDatatypeID error", err)
			return FetchErrMsg{Error: err}
		}

		if dtFields == nil || len(*dtFields) == 0 {
			logger.Finfo(fmt.Sprintf("No fields found for datatype %s", datatypeID))
			return FetchErrMsg{Error: fmt.Errorf("no fields found for datatype %s", datatypeID)}
		}

		// Fetch actual field details for each field ID
		var fields []db.Fields
		for _, dtf := range *dtFields {
			if dtf.FieldID.Valid {
				field, err := d.GetField(dtf.FieldID.ID)
				if err == nil && field != nil {
					fields = append(fields, *field)
				}
			}
		}

		if len(fields) == 0 {
			logger.Finfo(fmt.Sprintf("No valid fields found for datatype %s", datatypeID))
			return FetchErrMsg{Error: fmt.Errorf("no fields found for datatype %s", datatypeID)}
		}

		logger.Finfo(fmt.Sprintf("Found %d fields for datatype %s", len(fields), datatypeID))

		// Build form inputs for each field
		var formFields []huh.Field
		var formMap []string
		var values []*string

		for _, field := range fields {
			value := ""
			values = append(values, &value)

			// Create input field based on field type
			input := huh.NewInput().
				Title(field.Label).
				Value(&value)

			// Add description if field has data
			if field.Data != "" {
				input = input.Description(field.Data)
			}

			formFields = append(formFields, input)
			formMap = append(formMap, fmt.Sprint(field.FieldID))
		}

		// Create form with field group
		group := huh.NewGroup(formFields...)
		form := huh.NewForm(group)
		form.Init()

		// Set submit handler to dispatch content creation
		form.SubmitCmd = tea.Batch(
			LogMessageCmd(fmt.Sprintf("Form submitted for content creation (Datatype: %s, Route: %s)", datatypeID, routeID)),
			CmsAddNewContentDataCmd(datatypeID),
			FocusSetCmd(PAGEFOCUS),
			func() tea.Msg {
				return tea.ResumeMsg{}
			},
		)

		logger.Finfo(fmt.Sprintf("Returning NewFormMsg with %d fields", len(fields)))

		return NewFormMsg{
			Form:        form,
			FieldsCount: len(fields),
			Values:      values,
			FormMap:     formMap,
		}
	}
}

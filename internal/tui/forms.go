package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// FormIndex enumerates form type indices.
type FormIndex int

// FormIndex constants define form types.
const (
	DATABASECREATE FormIndex = iota
	DATABASEUPDATE
	CMSCREATE
	CMSUPDATE
)

// NewDefineDatatypeForm creates a form for defining a new datatype.
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
		datatype string
	)
	groupDescription := "Define datatype"
	typeDescription := "Types starting with '_' enable system features (e.g. _root, _nested_root).\n"
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
						Key:   "_root",
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
				Validate(func(s string) error {
					return types.ValidateUserDatatypeType(s)
				}).
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

// NewEditDatatypeForm creates a form for editing an existing datatype.
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
	typeDescription := "Types starting with '_' enable system features (e.g. _root, _nested_root).\n"
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
						Key:   "_root",
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
				Validate(func(s string) error {
					return types.ValidateUserDatatypeType(s)
				}).
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
				Name:       "", // Legacy form has no Name input; handler preserves existing value
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

// CreateDatatypeForm creates a form for creating a new datatype.
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

// CreateFieldForm creates a form for adding a field to a datatype.
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

// BuildCMSFieldForm creates a command to build a form for CMS fields.
func (m Model) BuildCMSFieldForm() tea.Cmd {
	return func() tea.Msg {
		form, count := CreateFieldForm(m)
		return NewFormMsg{
			Form:        form,
			FieldsCount: count,
		}
	}
}

// CMSFormControls handles the UI controls for CMS forms.
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
		m.Page = m.PageMap[CMSPAGE]
	}

	if m.FormState.Form.State == huh.StateCompleted {
		_ = tea.ClearScreen()
		m.Focus = PAGEFOCUS
		// Process completed form (implementation would depend on form type)
		m.Page = m.PageMap[CMSPAGE]
	}

	return m, tea.Batch(cmds...)
}

// NewFormMsg carries a newly constructed form and field metadata.
type NewFormMsg struct {
	Form        *huh.Form
	FieldsCount int
	Values      []*string
	FormMap     []string
}

// BuildContentFieldsForm creates a command to build a dynamic form for content creation.
func (m Model) BuildContentFieldsForm(datatypeID types.DatatypeID, routeID types.RouteID) tea.Cmd {
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}
	return func() tea.Msg {
		d := db.ConfigDB(*m.Config)

		logger.Finfo(fmt.Sprintf("Building content form for datatype %s, route %s", datatypeID, routeID))

		// Fetch fields by parent datatype ID
		fieldList, err := d.ListFieldsByDatatypeID(types.NullableDatatypeID{ID: datatypeID, Valid: true})
		if err != nil {
			logger.Ferror("ListFieldsByDatatypeID error", err)
			return FetchErrMsg{Error: err}
		}

		var fields []db.Fields
		if fieldList != nil {
			fields = *fieldList
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

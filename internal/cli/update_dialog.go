package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

type UpdatedDialog struct{}

func NewDialogUpdate() tea.Cmd {
	return func() tea.Msg {
		return UpdatedDialog{}
	}
}

func (m Model) UpdateDialog(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DialogReadyOKSet:
		newModel := m
		if newModel.Dialog != nil {
			newModel.Dialog.ReadyOK = msg.Ready
		}
		return newModel, NewDialogUpdate()
	case ShowDialogMsg:
		// Handle showing a dialog
		dialog := NewDialog(msg.Title, msg.Message, msg.ShowCancel, DIALOGDELETE)
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ActionConfirmMsg:
		// Show confirmation dialog for destructive actions
		actions := ActionsMenu()
		label := "this action"
		if msg.ActionIndex < len(actions) {
			label = actions[msg.ActionIndex].Label
		}
		dialog := NewDialog(
			"Confirm: "+label,
			"WARNING: This is a destructive operation that cannot be undone. Continue?",
			true,
			DIALOGACTIONCONFIRM,
		)
		dialog.ActionIndex = msg.ActionIndex
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ActionResultMsg:
		// Show result dialog after an action completes
		dialog := NewDialog(msg.Title, msg.Message, false, DIALOGGENERIC)
		return m, tea.Batch(
			LoadingStopCmd(),
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case DialogAcceptMsg:
		// Handle dialog accept action
		switch msg.Action {
		case DIALOGDELETE:
			id := m.GetCurrentRowId()
			return m, tea.Batch(
				DatabaseDeleteEntryCmd(int(id), m.TableState.Table),
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case DIALOGACTIONCONFIRM:
			actionIndex := 0
			if m.Dialog != nil {
				actionIndex = m.Dialog.ActionIndex
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				RunDestructiveActionCmd(ActionParams{
					Config:         m.Config,
					UserID:         m.UserID,
					SSHFingerprint: m.SSHFingerprint,
					SSHKeyType:     m.SSHKeyType,
					SSHPublicKey:   m.SSHPublicKey,
				}, actionIndex),
			)
		default:
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		}
	case DialogCancelMsg:
		// Handle dialog cancel action
		return m, tea.Batch(
			DialogActiveSetCmd(false),
			FocusSetCmd(PAGEFOCUS),
		)

	// Form dialog handling
	case ShowFormDialogMsg:
		dialog := NewFormDialog(msg.Title, msg.Action, msg.Parents)
		return m, tea.Batch(
			FormDialogSetCmd(&dialog),
			FormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowFieldFormDialogMsg:
		// Field form dialog has no parent selector
		dialog := NewFieldFormDialog(msg.Title, msg.Action)
		return m, tea.Batch(
			FormDialogSetCmd(&dialog),
			FormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case FormDialogAcceptMsg:
		// Handle form dialog accept based on action type
		switch msg.Action {
		case FORMDIALOGCREATEDATATYPE:
			// Create the datatype
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				CreateDatatypeFromDialogCmd(msg.Label, msg.Type, msg.ParentID),
			)
		case FORMDIALOGCREATEFIELD:
			// Create a field and link it to the selected datatype
			if len(m.AllDatatypes) > 0 && m.Cursor < len(m.AllDatatypes) {
				dt := m.AllDatatypes[m.Cursor]
				return m, tea.Batch(
					FormDialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					CreateFieldFromDialogCmd(msg.Label, msg.Type, dt.DatatypeID),
				)
			}
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		default:
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		}
	case FormDialogCancelMsg:
		return m, tea.Batch(
			FormDialogActiveSetCmd(false),
			FocusSetCmd(PAGEFOCUS),
		)
	case DatatypeCreatedFromDialogMsg:
		// Refresh datatypes list after creation
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Datatype created: %s", msg.Label)),
			AllDatatypesFetchCmd(),
		)
	case FieldCreatedFromDialogMsg:
		// Refresh fields list after creation
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Field created: %s", msg.Label)),
			DatatypeFieldsFetchCmd(msg.DatatypeID),
		)
	default:
		return m, nil
	}
}

// DatatypeCreatedFromDialogMsg is sent after a datatype is created from the form dialog
type DatatypeCreatedFromDialogMsg struct {
	DatatypeID types.DatatypeID
	Label      string
}

// CreateDatatypeFromDialogCmd creates a datatype from form dialog input
func CreateDatatypeFromDialogCmd(label, dtype, parentID string) tea.Cmd {
	return func() tea.Msg {
		return CreateDatatypeFromDialogRequestMsg{
			Label:    label,
			Type:     dtype,
			ParentID: parentID,
		}
	}
}

// CreateDatatypeFromDialogRequestMsg triggers datatype creation
type CreateDatatypeFromDialogRequestMsg struct {
	Label    string
	Type     string
	ParentID string
}

// HandleCreateDatatypeFromDialog processes the creation request
func (m Model) HandleCreateDatatypeFromDialog(msg CreateDatatypeFromDialogRequestMsg) tea.Cmd {
	// Capture values from model for use in closure
	authorID := m.UserID
	cfg := m.Config

	// Validate that we have a user ID (required by database constraint)
	if authorID.IsZero() {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create datatype: no user is logged in",
			}
		}
	}

	// Validate config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create datatype: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

		// Prepare the type - default to ROOT if empty
		dtype := msg.Type
		if dtype == "" {
			dtype = "ROOT"
		}

		// Prepare parent ID (uses NullableContentID per db package definition)
		var parentID types.NullableContentID
		if msg.ParentID != "" {
			parentID = types.NullableContentID{
				ID:    types.ContentID(msg.ParentID),
				Valid: true,
			}
		}

		// Set author ID (required by database NOT NULL constraint)
		nullableAuthorID := types.NullableUserID{
			ID:    authorID,
			Valid: true,
		}

		// Create the datatype
		params := db.CreateDatatypeParams{
			DatatypeID:   types.NewDatatypeID(),
			ParentID:     parentID,
			Label:        msg.Label,
			Type:         dtype,
			AuthorID:     nullableAuthorID,
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		dt := d.CreateDatatype(params)
		return DatatypeCreatedFromDialogMsg{
			DatatypeID: dt.DatatypeID,
			Label:      dt.Label,
		}
	}
}

// FieldCreatedFromDialogMsg is sent after a field is created from the form dialog
type FieldCreatedFromDialogMsg struct {
	FieldID    types.FieldID
	DatatypeID types.DatatypeID
	Label      string
}

// CreateFieldFromDialogCmd creates a field and links it to a datatype
func CreateFieldFromDialogCmd(label, fieldType string, datatypeID types.DatatypeID) tea.Cmd {
	return func() tea.Msg {
		return CreateFieldFromDialogRequestMsg{
			Label:      label,
			Type:       fieldType,
			DatatypeID: datatypeID,
		}
	}
}

// CreateFieldFromDialogRequestMsg triggers field creation
type CreateFieldFromDialogRequestMsg struct {
	Label      string
	Type       string
	DatatypeID types.DatatypeID
}

// HandleCreateFieldFromDialog processes the field creation request
func (m Model) HandleCreateFieldFromDialog(msg CreateFieldFromDialogRequestMsg) tea.Cmd {
	// Capture values from model for use in closure
	authorID := m.UserID
	cfg := m.Config

	// Validate that we have a user ID
	if authorID.IsZero() {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create field: no user is logged in",
			}
		}
	}

	// Validate config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create field: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

		// Prepare the field type - default to "text" if empty
		fieldTypeStr := msg.Type
		if fieldTypeStr == "" {
			fieldTypeStr = "text"
		}
		fieldType := types.FieldType(fieldTypeStr)

		// Set author ID
		nullableAuthorID := types.NullableUserID{
			ID:    authorID,
			Valid: true,
		}

		// Create the field
		fieldID := types.NewFieldID()
		fieldParams := db.CreateFieldParams{
			FieldID:      fieldID,
			Label:        msg.Label,
			Data:         "", // Empty data for now, can be extended later
			Type:         fieldType,
			AuthorID:     nullableAuthorID,
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		field := d.CreateField(fieldParams)
		if field.FieldID.IsZero() {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Failed to create field in database",
			}
		}

		// Link field to datatype via datatypes_fields join table
		dtFieldID := string(types.NewDatatypeFieldID())
		dtFieldParams := db.CreateDatatypeFieldParams{
			ID: dtFieldID,
			DatatypeID: types.NullableDatatypeID{
				ID:    msg.DatatypeID,
				Valid: true,
			},
			FieldID: types.NullableFieldID{
				ID:    field.FieldID,
				Valid: true,
			},
		}

		d.CreateDatatypeField(dtFieldParams)

		return FieldCreatedFromDialogMsg{
			FieldID:    field.FieldID,
			DatatypeID: msg.DatatypeID,
			Label:      field.Label,
		}
	}
}

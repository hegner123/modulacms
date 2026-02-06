package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
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
	case ShowQuitConfirmDialogMsg:
		// Show quit confirmation dialog
		dialog := NewDialog("Quit", "Are you sure you want to quit?", true, DIALOGQUITCONFIRM)
		dialog.SetButtons("Quit", "Cancel")
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteContentDialogMsg:
		// Show delete content confirmation dialog
		var dialogMsg string
		if msg.HasChildren {
			dialogMsg = fmt.Sprintf("Cannot delete '%s' because it has children.\nDelete child nodes first.", msg.ContentName)
			dialog := NewDialog("Cannot Delete", dialogMsg, false, DIALOGGENERIC)
			return m, tea.Batch(
				DialogSetCmd(&dialog),
				DialogActiveSetCmd(true),
				FocusSetCmd(DIALOGFOCUS),
			)
		}
		dialogMsg = fmt.Sprintf("Delete '%s'?\nThis will also delete all field values.", msg.ContentName)
		dialog := NewDialog("Delete Content", dialogMsg, true, DIALOGDELETECONTENT)
		dialog.SetButtons("Delete", "Cancel")
		// Store the content ID for deletion
		deleteContentContext = &DeleteContentContext{
			ContentID: msg.ContentID,
			RouteID:   string(m.PageRouteId),
		}
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
		case DIALOGQUITCONFIRM:
			// User confirmed quit
			return m, tea.Quit
		case DIALOGDELETECONTENT:
			// User confirmed content deletion
			if deleteContentContext != nil {
				contentID := deleteContentContext.ContentID
				routeID := deleteContentContext.RouteID
				deleteContentContext = nil // Clear the context
				return m, tea.Batch(
					DialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					DeleteContentCmd(contentID, routeID),
				)
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
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
		case DIALOGINITCONTENT:
			// Initialize content for route using stored context
			if initializeRouteContentContext != nil {
				routeID := initializeRouteContentContext.Route.RouteID
				datatypeID := initializeRouteContentContext.DatatypeID
				initializeRouteContentContext = nil // Clear the context
				return m, tea.Batch(
					DialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					InitializeRouteContentCmd(routeID, datatypeID),
				)
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
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
	case ShowRouteFormDialogMsg:
		// Route form dialog has Title and Slug inputs
		dialog := NewRouteFormDialog(msg.Title, msg.Action)
		return m, tea.Batch(
			FormDialogSetCmd(&dialog),
			FormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditDatatypeDialogMsg:
		// Edit datatype dialog with pre-populated values
		dialog := NewEditDatatypeDialog("Edit Datatype", FORMDIALOGEDITDATATYPE, msg.Parents, msg.Datatype)
		return m, tea.Batch(
			FormDialogSetCmd(&dialog),
			FormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditFieldDialogMsg:
		// Edit field dialog with pre-populated values
		dialog := NewEditFieldDialog("Edit Field", FORMDIALOGEDITFIELD, msg.Field)
		return m, tea.Batch(
			FormDialogSetCmd(&dialog),
			FormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditRouteDialogMsg:
		// Edit route dialog with pre-populated values
		dialog := NewEditRouteDialog("Edit Route", FORMDIALOGEDITROUTE, msg.Route)
		return m, tea.Batch(
			FormDialogSetCmd(&dialog),
			FormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowCreateRouteWithContentDialogMsg:
		// Create route with initial content dialog
		dialog := NewRouteWithContentDialog("New Route with Content", FORMDIALOGCREATEROUTEWITHCONTENT, msg.DatatypeID)
		return m, tea.Batch(
			FormDialogSetCmd(&dialog),
			FormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowInitializeRouteContentDialogMsg:
		// Show confirmation dialog to initialize content for an existing route
		dialog := NewDialog(
			"Initialize Content",
			fmt.Sprintf("Create root content for route '%s'?", msg.Route.Title),
			true,
			DIALOGINITCONTENT,
		)
		// Store the route and datatype info for when the dialog is accepted
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
			// Store context for the initialization
			InitializeRouteContentContextCmd(msg.Route, msg.DatatypeID),
		)
	case ShowChildDatatypeDialogMsg:
		// Show dialog to select a child datatype for creating new content
		dialog := NewChildDatatypeDialog("Select Child Type", msg.ChildDatatypes, string(msg.RouteID))
		return m, tea.Batch(
			FormDialogSetCmd(&dialog),
			FormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case FormDialogAcceptMsg:
		// Handle form dialog accept based on action type
		switch msg.Action {
		case FORMDIALOGCREATEDATATYPE:
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
		case FORMDIALOGCREATEROUTE:
			// Create a new route (Label=Title, Type=Slug)
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				CreateRouteFromDialogCmd(msg.Label, msg.Type),
			)
		case FORMDIALOGEDITDATATYPE:
			// Update an existing datatype
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				UpdateDatatypeFromDialogCmd(msg.EntityID, msg.Label, msg.Type, msg.ParentID),
			)
		case FORMDIALOGEDITFIELD:
			// Update an existing field
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				UpdateFieldFromDialogCmd(msg.EntityID, msg.Label, msg.Type),
			)
		case FORMDIALOGEDITROUTE:
			// Update an existing route (Label=Title, Type=Slug)
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				UpdateRouteFromDialogCmd(msg.EntityID, msg.Label, msg.Type),
			)
		case FORMDIALOGCREATEROUTEWITHCONTENT:
			// Create a new route with initial content (EntityID=DatatypeID, Label=Title, Type=Slug)
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				CreateRouteWithContentCmd(msg.Label, msg.Type, msg.EntityID),
			)
		case FORMDIALOGCHILDDATATYPE:
			// User selected a child datatype from the dialog
			// ParentID contains the selected datatype ID, EntityID contains the route ID
			if m.Logger != nil {
				m.Logger.Finfo(fmt.Sprintf("FORMDIALOGCHILDDATATYPE accepted: ParentID=%s, EntityID=%s", msg.ParentID, msg.EntityID))
			}
			if msg.ParentID != "" {
				if m.Logger != nil {
					m.Logger.Finfo("Dispatching ChildDatatypeSelectedCmd")
				}
				return m, tea.Batch(
					FormDialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					ChildDatatypeSelectedCmd(types.DatatypeID(msg.ParentID), types.RouteID(msg.EntityID)),
				)
			}
			if m.Logger != nil {
				m.Logger.Finfo("ParentID was empty, just closing dialog")
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
	case ShowContentFormDialogMsg:
		// Create content form dialog with dynamic fields
		logger := m.Logger
		if logger == nil {
			logger = utility.DefaultLogger
		}
		logger.Finfo(fmt.Sprintf("ShowContentFormDialogMsg received: %d fields, ParentID.Valid=%v", len(msg.Fields), msg.ParentID.Valid))
		var dialog ContentFormDialogModel
		if msg.ParentID.Valid {
			dialog = NewContentFormDialogWithParent(msg.Title, msg.Action, msg.DatatypeID, msg.RouteID, msg.ParentID.ID, msg.Fields)
		} else {
			dialog = NewContentFormDialog(msg.Title, msg.Action, msg.DatatypeID, msg.RouteID, msg.Fields)
		}
		logger.Finfo(fmt.Sprintf("ContentFormDialogModel created with %d fields", len(dialog.Fields)))
		return m, tea.Batch(
			ContentFormDialogSetCmd(&dialog),
			ContentFormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ContentFormDialogAcceptMsg:
		// Handle content form submission based on action
		switch msg.Action {
		case FORMDIALOGEDITCONTENT:
			// Update existing content
			return m, tea.Batch(
				ContentFormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				UpdateContentFromDialogCmd(msg.ContentID, msg.DatatypeID, msg.RouteID, msg.FieldValues),
			)
		default:
			// Create new content (FORMDIALOGCREATECONTENT or default)
			return m, tea.Batch(
				ContentFormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				CreateContentFromDialogCmd(msg.DatatypeID, msg.RouteID, msg.ParentID, msg.FieldValues),
			)
		}
	case ContentFormDialogCancelMsg:
		return m, tea.Batch(
			ContentFormDialogActiveSetCmd(false),
			FocusSetCmd(PAGEFOCUS),
		)
	case ShowEditContentFormDialogMsg:
		// Create edit content form dialog with pre-populated values
		logger := m.Logger
		if logger == nil {
			logger = utility.DefaultLogger
		}
		logger.Finfo(fmt.Sprintf("ShowEditContentFormDialogMsg received: ContentID=%s, %d fields", msg.ContentID, len(msg.ExistingFields)))
		dialog := NewEditContentFormDialog(msg.Title, msg.ContentID, msg.DatatypeID, msg.RouteID, msg.ExistingFields)
		logger.Finfo(fmt.Sprintf("EditContentFormDialogModel created with %d fields", len(dialog.Fields)))
		return m, tea.Batch(
			ContentFormDialogSetCmd(&dialog),
			ContentFormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ContentCreatedFromDialogMsg:
		// Content created successfully from dialog - reload tree and show success
		return m, tea.Batch(
			LoadingStopCmd(),
			ShowDialog(
				"Success",
				fmt.Sprintf("✓ Content created with %d fields", msg.FieldCount),
				false,
			),
			LogMessageCmd(fmt.Sprintf("Content created: ID=%s, DatatypeID=%s", msg.ContentID, msg.DatatypeID)),
			ReloadContentTreeCmd(m.Config, msg.RouteID),
		)
	case ContentUpdatedFromDialogMsg:
		// Content updated successfully from dialog - reload tree and show success
		return m, tea.Batch(
			LoadingStopCmd(),
			ShowDialog(
				"Success",
				fmt.Sprintf("✓ Content updated (%d fields)", msg.UpdatedCount),
				false,
			),
			LogMessageCmd(fmt.Sprintf("Content updated: ID=%s, DatatypeID=%s", msg.ContentID, msg.DatatypeID)),
			ReloadContentTreeCmd(m.Config, msg.RouteID),
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
	case RouteCreatedFromDialogMsg:
		// Refresh routes list after creation
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Route created: %s", msg.Title)),
			RoutesFetchCmd(),
		)
	case DatatypeUpdatedFromDialogMsg:
		// Refresh datatypes list after update
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Datatype updated: %s", msg.Label)),
			AllDatatypesFetchCmd(),
		)
	case FieldUpdatedFromDialogMsg:
		// Refresh fields list after update
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Field updated: %s", msg.Label)),
			DatatypeFieldsFetchCmd(msg.DatatypeID),
		)
	case RouteUpdatedFromDialogMsg:
		// Refresh routes list after update
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Route updated: %s", msg.Title)),
			RoutesFetchCmd(),
		)
	case RouteWithContentCreatedMsg:
		// Refresh routes list and show success after route+content creation
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Route created with content: %s (ContentID: %s)", msg.Title, msg.ContentDataID)),
			RoutesByDatatypeFetchCmd(msg.DatatypeID),
		)
	case RouteContentInitializedMsg:
		// Refresh and load content tree after initialization
		newModel := m
		newModel.PageRouteId = msg.RouteID
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Content initialized for route: %s", msg.Title)),
			ReloadContentTreeCmd(m.Config, msg.RouteID),
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
			Data:         "",
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

// RouteCreatedFromDialogMsg is sent after a route is created from the form dialog
type RouteCreatedFromDialogMsg struct {
	RouteID types.RouteID
	Title   string
	Slug    string
}

// CreateRouteFromDialogCmd creates a route from form dialog input
func CreateRouteFromDialogCmd(title, slug string) tea.Cmd {
	return func() tea.Msg {
		return CreateRouteFromDialogRequestMsg{
			Title: title,
			Slug:  slug,
		}
	}
}

// CreateRouteFromDialogRequestMsg triggers route creation
type CreateRouteFromDialogRequestMsg struct {
	Title string
	Slug  string
}

// HandleCreateRouteFromDialog processes the route creation request
func (m Model) HandleCreateRouteFromDialog(msg CreateRouteFromDialogRequestMsg) tea.Cmd {
	// Capture values from model for use in closure
	authorID := m.UserID
	cfg := m.Config

	// Validate that we have a user ID (required by database constraint)
	if authorID.IsZero() {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create route: no user is logged in",
			}
		}
	}

	// Validate config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create route: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

		// Prepare the slug - use Slugify to ensure valid format
		slug := msg.Slug
		if slug == "" {
			slug = msg.Title
		}
		validSlug := types.Slugify(slug)

		// Validate the slug
		if err := validSlug.Validate(); err != nil {
			return ActionResultMsg{
				Title:   "Invalid Slug",
				Message: fmt.Sprintf("Could not create route: %v", err),
			}
		}

		// Check if slug already exists
		existingID, _ := d.GetRouteID(string(validSlug))
		if existingID != nil {
			return ActionResultMsg{
				Title:   "Duplicate Slug",
				Message: fmt.Sprintf("A route with slug %q already exists", validSlug),
			}
		}

		// Create the route
		params := db.CreateRouteParams{
			RouteID: types.NewRouteID(),
			Slug:    validSlug,
			Title:   msg.Title,
			Status:  1, // Active by default
			AuthorID: types.NullableUserID{
				ID:    authorID,
				Valid: true,
			},
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		route := d.CreateRoute(params)
		if route.RouteID.IsZero() {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Failed to create route in database",
			}
		}

		return RouteCreatedFromDialogMsg{
			RouteID: route.RouteID,
			Title:   route.Title,
			Slug:    string(route.Slug),
		}
	}
}

// =============================================================================
// UPDATE DATATYPE FROM DIALOG
// =============================================================================

// DatatypeUpdatedFromDialogMsg is sent after a datatype is updated from the form dialog
type DatatypeUpdatedFromDialogMsg struct {
	DatatypeID types.DatatypeID
	Label      string
}

// UpdateDatatypeFromDialogRequestMsg triggers datatype update
type UpdateDatatypeFromDialogRequestMsg struct {
	DatatypeID string
	Label      string
	Type       string
	ParentID   string
}

// UpdateDatatypeFromDialogCmd creates a command to update a datatype from form dialog input
func UpdateDatatypeFromDialogCmd(datatypeID, label, dtype, parentID string) tea.Cmd {
	return func() tea.Msg {
		return UpdateDatatypeFromDialogRequestMsg{
			DatatypeID: datatypeID,
			Label:      label,
			Type:       dtype,
			ParentID:   parentID,
		}
	}
}

// HandleUpdateDatatypeFromDialog processes the datatype update request
func (m Model) HandleUpdateDatatypeFromDialog(msg UpdateDatatypeFromDialogRequestMsg) tea.Cmd {
	// Capture values from model for use in closure
	authorID := m.UserID
	cfg := m.Config

	// Validate config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot update datatype: configuration not loaded",
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

		// Set author ID
		nullableAuthorID := types.NullableUserID{
			ID:    authorID,
			Valid: !authorID.IsZero(),
		}

		// Update the datatype
		params := db.UpdateDatatypeParams{
			DatatypeID:   types.DatatypeID(msg.DatatypeID),
			ParentID:     parentID,
			Label:        msg.Label,
			Type:         dtype,
			AuthorID:     nullableAuthorID,
			DateModified: types.TimestampNow(),
		}

		_, err := d.UpdateDatatype(params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update datatype: %v", err),
			}
		}

		return DatatypeUpdatedFromDialogMsg{
			DatatypeID: types.DatatypeID(msg.DatatypeID),
			Label:      msg.Label,
		}
	}
}

// =============================================================================
// UPDATE FIELD FROM DIALOG
// =============================================================================

// FieldUpdatedFromDialogMsg is sent after a field is updated from the form dialog
type FieldUpdatedFromDialogMsg struct {
	FieldID    types.FieldID
	DatatypeID types.DatatypeID
	Label      string
}

// UpdateFieldFromDialogRequestMsg triggers field update
type UpdateFieldFromDialogRequestMsg struct {
	FieldID string
	Label   string
	Type    string
}

// UpdateFieldFromDialogCmd creates a command to update a field from form dialog input
func UpdateFieldFromDialogCmd(fieldID, label, fieldType string) tea.Cmd {
	return func() tea.Msg {
		return UpdateFieldFromDialogRequestMsg{
			FieldID: fieldID,
			Label:   label,
			Type:    fieldType,
		}
	}
}

// HandleUpdateFieldFromDialog processes the field update request
func (m Model) HandleUpdateFieldFromDialog(msg UpdateFieldFromDialogRequestMsg) tea.Cmd {
	// Capture values from model for use in closure
	authorID := m.UserID
	cfg := m.Config
	// Capture the current datatype ID to refresh fields after update
	var datatypeID types.DatatypeID
	if len(m.AllDatatypes) > 0 && m.Cursor < len(m.AllDatatypes) {
		datatypeID = m.AllDatatypes[m.Cursor].DatatypeID
	}

	// Validate config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot update field: configuration not loaded",
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
			Valid: !authorID.IsZero(),
		}

		// Update the field
		params := db.UpdateFieldParams{
			FieldID:      types.FieldID(msg.FieldID),
			Label:        msg.Label,
			Type:         fieldType,
			AuthorID:     nullableAuthorID,
			DateModified: types.TimestampNow(),
		}

		_, err := d.UpdateField(params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update field: %v", err),
			}
		}

		return FieldUpdatedFromDialogMsg{
			FieldID:    types.FieldID(msg.FieldID),
			DatatypeID: datatypeID,
			Label:      msg.Label,
		}
	}
}

// =============================================================================
// UPDATE ROUTE FROM DIALOG
// =============================================================================

// RouteUpdatedFromDialogMsg is sent after a route is updated from the form dialog
type RouteUpdatedFromDialogMsg struct {
	RouteID types.RouteID
	Title   string
	Slug    string
}

// UpdateRouteFromDialogRequestMsg triggers route update
type UpdateRouteFromDialogRequestMsg struct {
	RouteID string
	Title   string
	Slug    string
}

// UpdateRouteFromDialogCmd creates a command to update a route from form dialog input
func UpdateRouteFromDialogCmd(routeID, title, slug string) tea.Cmd {
	return func() tea.Msg {
		return UpdateRouteFromDialogRequestMsg{
			RouteID: routeID,
			Title:   title,
			Slug:    slug,
		}
	}
}

// HandleUpdateRouteFromDialog processes the route update request
func (m Model) HandleUpdateRouteFromDialog(msg UpdateRouteFromDialogRequestMsg) tea.Cmd {
	// Capture values from model for use in closure
	authorID := m.UserID
	cfg := m.Config

	// Validate config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot update route: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

		// Get the existing route to preserve its original slug for the WHERE clause
		existingRoute, err := d.GetRoute(types.RouteID(msg.RouteID))
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Route not found: %v", err),
			}
		}

		// Prepare the slug - use Slugify to ensure valid format
		slug := msg.Slug
		if slug == "" {
			slug = msg.Title
		}
		validSlug := types.Slugify(slug)

		// Validate the slug
		if err := validSlug.Validate(); err != nil {
			return ActionResultMsg{
				Title:   "Invalid Slug",
				Message: fmt.Sprintf("Could not update route: %v", err),
			}
		}

		// Check if new slug already exists (unless it's the same route)
		if validSlug != existingRoute.Slug {
			existingID, _ := d.GetRouteID(string(validSlug))
			if existingID != nil {
				return ActionResultMsg{
					Title:   "Duplicate Slug",
					Message: fmt.Sprintf("A route with slug %q already exists", validSlug),
				}
			}
		}

		// Set author ID
		nullableAuthorID := types.NullableUserID{
			ID:    authorID,
			Valid: !authorID.IsZero(),
		}

		// Update the route
		// Note: UpdateRouteParams uses Slug_2 for the WHERE clause (original slug)
		params := db.UpdateRouteParams{
			Slug:         validSlug,
			Title:        msg.Title,
			Status:       existingRoute.Status,
			AuthorID:     nullableAuthorID,
			DateCreated:  existingRoute.DateCreated,
			DateModified: types.TimestampNow(),
			Slug_2:       existingRoute.Slug, // Original slug for WHERE clause
		}

		_, err = d.UpdateRoute(params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update route: %v", err),
			}
		}

		return RouteUpdatedFromDialogMsg{
			RouteID: types.RouteID(msg.RouteID),
			Title:   msg.Title,
			Slug:    string(validSlug),
		}
	}
}

// =============================================================================
// CREATE ROUTE WITH CONTENT
// =============================================================================

// RouteWithContentCreatedMsg is sent after a route and initial content are created
type RouteWithContentCreatedMsg struct {
	RouteID       types.RouteID
	ContentDataID types.ContentID
	DatatypeID    types.DatatypeID
	Title         string
	Slug          string
}

// CreateRouteWithContentRequestMsg triggers route and content creation
type CreateRouteWithContentRequestMsg struct {
	Title      string
	Slug       string
	DatatypeID string
}

// CreateRouteWithContentCmd creates a command to create a route with initial content
func CreateRouteWithContentCmd(title, slug, datatypeID string) tea.Cmd {
	return func() tea.Msg {
		return CreateRouteWithContentRequestMsg{
			Title:      title,
			Slug:       slug,
			DatatypeID: datatypeID,
		}
	}
}

// HandleCreateRouteWithContent processes the route with content creation request
func (m Model) HandleCreateRouteWithContent(msg CreateRouteWithContentRequestMsg) tea.Cmd {
	// Capture values from model for use in closure
	authorID := m.UserID
	cfg := m.Config

	// Validate that we have a user ID (required by database constraint)
	if authorID.IsZero() {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create route: no user is logged in",
			}
		}
	}

	// Validate config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create route: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

		// Prepare the slug - use Slugify to ensure valid format
		slug := msg.Slug
		if slug == "" {
			slug = msg.Title
		}
		validSlug := types.Slugify(slug)

		// Validate the slug
		if err := validSlug.Validate(); err != nil {
			return ActionResultMsg{
				Title:   "Invalid Slug",
				Message: fmt.Sprintf("Could not create route: %v", err),
			}
		}

		// Check if slug already exists
		existingID, _ := d.GetRouteID(string(validSlug))
		if existingID != nil {
			return ActionResultMsg{
				Title:   "Duplicate Slug",
				Message: fmt.Sprintf("A route with slug %q already exists", validSlug),
			}
		}

		// Create the route
		routeParams := db.CreateRouteParams{
			RouteID: types.NewRouteID(),
			Slug:    validSlug,
			Title:   msg.Title,
			Status:  1, // Active by default
			AuthorID: types.NullableUserID{
				ID:    authorID,
				Valid: true,
			},
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		route := d.CreateRoute(routeParams)
		if route.RouteID.IsZero() {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Failed to create route in database",
			}
		}

		// Create initial content data for this route
		datatypeID := types.DatatypeID(msg.DatatypeID)
		contentParams := db.CreateContentDataParams{
			RouteID: types.NullableRouteID{
				ID:    route.RouteID,
				Valid: true,
			},
			DatatypeID: types.NullableDatatypeID{
				ID:    datatypeID,
				Valid: true,
			},
			AuthorID: types.NullableUserID{
				ID:    authorID,
				Valid: true,
			},
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		contentData := d.CreateContentData(contentParams)
		if contentData.ContentDataID.IsZero() {
			return ActionResultMsg{
				Title:   "Warning",
				Message: fmt.Sprintf("Route created but failed to create initial content. Route: %s", route.Title),
			}
		}

		return RouteWithContentCreatedMsg{
			RouteID:       route.RouteID,
			ContentDataID: contentData.ContentDataID,
			DatatypeID:    datatypeID,
			Title:         route.Title,
			Slug:          string(route.Slug),
		}
	}
}

// =============================================================================
// INITIALIZE ROUTE CONTENT
// =============================================================================

// InitializeRouteContentContext stores context for initializing route content
type InitializeRouteContentContext struct {
	Route      db.Routes
	DatatypeID string
}

// Global variable to store the context (will be set before dialog is shown)
var initializeRouteContentContext *InitializeRouteContentContext

// =============================================================================
// DELETE CONTENT
// =============================================================================

// DeleteContentContext stores context for deleting content
type DeleteContentContext struct {
	ContentID string
	RouteID   string
}

// Global variable to store the delete context
var deleteContentContext *DeleteContentContext

// InitializeRouteContentContextCmd stores the context for route content initialization
func InitializeRouteContentContextCmd(route db.Routes, datatypeID string) tea.Cmd {
	return func() tea.Msg {
		initializeRouteContentContext = &InitializeRouteContentContext{
			Route:      route,
			DatatypeID: datatypeID,
		}
		return nil
	}
}

// RouteContentInitializedMsg is sent after content is initialized for a route
type RouteContentInitializedMsg struct {
	RouteID       types.RouteID
	ContentDataID types.ContentID
	DatatypeID    types.DatatypeID
	Title         string
}

// InitializeRouteContentRequestMsg triggers content initialization for a route
type InitializeRouteContentRequestMsg struct {
	RouteID    types.RouteID
	DatatypeID string
}

// InitializeRouteContentCmd creates a command to initialize content for a route
func InitializeRouteContentCmd(routeID types.RouteID, datatypeID string) tea.Cmd {
	return func() tea.Msg {
		return InitializeRouteContentRequestMsg{
			RouteID:    routeID,
			DatatypeID: datatypeID,
		}
	}
}

// HandleInitializeRouteContent processes the route content initialization request
func (m Model) HandleInitializeRouteContent(msg InitializeRouteContentRequestMsg) tea.Cmd {
	// Capture values from model for use in closure
	authorID := m.UserID
	cfg := m.Config

	// Validate config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot initialize content: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

		// Get the route to include its title in the response
		route, err := d.GetRoute(msg.RouteID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Route not found: %v", err),
			}
		}

		// Create initial content data for this route
		datatypeID := types.DatatypeID(msg.DatatypeID)
		contentParams := db.CreateContentDataParams{
			RouteID: types.NullableRouteID{
				ID:    msg.RouteID,
				Valid: true,
			},
			DatatypeID: types.NullableDatatypeID{
				ID:    datatypeID,
				Valid: true,
			},
			AuthorID: types.NullableUserID{
				ID:    authorID,
				Valid: !authorID.IsZero(),
			},
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		contentData := d.CreateContentData(contentParams)
		if contentData.ContentDataID.IsZero() {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Failed to create content in database",
			}
		}

		return RouteContentInitializedMsg{
			RouteID:       msg.RouteID,
			ContentDataID: contentData.ContentDataID,
			DatatypeID:    datatypeID,
			Title:         route.Title,
		}
	}
}

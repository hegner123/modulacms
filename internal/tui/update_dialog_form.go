package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// =============================================================================
// CREATE/UPDATE USER FROM DIALOG
// =============================================================================

// CreateUserFromDialogRequestMsg triggers user creation from dialog.
type CreateUserFromDialogRequestMsg struct {
	Username string
	Name     string
	Email    string
	Password string
	Role     string
}

// UserCreatedFromDialogMsg is sent after a user is successfully created from dialog.
type UserCreatedFromDialogMsg struct {
	UserID   types.UserID
	Username string
}

// UpdateUserFromDialogRequestMsg triggers user update from dialog.
type UpdateUserFromDialogRequestMsg struct {
	UserID   string
	Username string
	Name     string
	Email    string
	Role     string
}

// UserUpdatedFromDialogMsg is sent after a user is successfully updated from dialog.
type UserUpdatedFromDialogMsg struct {
	UserID   types.UserID
	Username string
}

// ShowCreateUserDialogCmd creates a command to show a user creation dialog.
func ShowCreateUserDialogCmd(roles []db.Roles) tea.Cmd {
	return func() tea.Msg {
		return ShowUserFormDialogMsg{Title: "New User", Roles: roles}
	}
}

// ShowEditUserDialogCmd creates a command to show a user edit dialog.
func ShowEditUserDialogCmd(user db.UserWithRoleLabelRow, roles []db.Roles) tea.Cmd {
	return func() tea.Msg {
		return ShowEditUserDialogMsg{User: user, Roles: roles}
	}
}

// CreateUserFromDialogCmd creates a command to trigger user creation from dialog.
func CreateUserFromDialogCmd(username, name, email, password, role string) tea.Cmd {
	return func() tea.Msg {
		return CreateUserFromDialogRequestMsg{
			Username: username,
			Name:     name,
			Email:    email,
			Password: password,
			Role:     role,
		}
	}
}

// UpdateUserFromDialogCmd creates a command to trigger user update from dialog.
func UpdateUserFromDialogCmd(userID, username, name, email, role string) tea.Cmd {
	return func() tea.Msg {
		return UpdateUserFromDialogRequestMsg{
			UserID:   userID,
			Username: username,
			Name:     name,
			Email:    email,
			Role:     role,
		}
	}
}

// =============================================================================
// CREATE/UPDATE WEBHOOK FROM DIALOG
// =============================================================================

// CreateWebhookFromDialogRequestMsg triggers webhook creation from dialog.
type CreateWebhookFromDialogRequestMsg struct {
	Name     string
	URL      string
	Secret   string
	Events   string // comma-separated
	IsActive bool
}

// WebhookCreatedMsg is sent after a webhook is successfully created.
type WebhookCreatedMsg struct {
	WebhookID types.WebhookID
	Name      string
}

// UpdateWebhookFromDialogRequestMsg triggers webhook update from dialog.
type UpdateWebhookFromDialogRequestMsg struct {
	WebhookID string
	Name      string
	URL       string
	Secret    string
	Events    string // comma-separated
	IsActive  bool
}

// WebhookUpdatedMsg is sent after a webhook is successfully updated.
type WebhookUpdatedMsg struct {
	WebhookID types.WebhookID
	Name      string
}

// ShowCreateWebhookDialogCmd creates a command to show a webhook creation dialog.
func ShowCreateWebhookDialogCmd() tea.Cmd {
	return func() tea.Msg {
		return ShowWebhookFormDialogMsg{Title: "New Webhook"}
	}
}

// ShowEditWebhookDialogCmd creates a command to show a webhook edit dialog.
func ShowEditWebhookDialogCmd(webhook db.Webhook) tea.Cmd {
	return func() tea.Msg {
		return ShowEditWebhookDialogMsg{Webhook: webhook}
	}
}

// CreateWebhookFromDialogCmd creates a command to trigger webhook creation from dialog.
func CreateWebhookFromDialogCmd(name, url, secret, events string, isActive bool) tea.Cmd {
	return func() tea.Msg {
		return CreateWebhookFromDialogRequestMsg{
			Name:     name,
			URL:      url,
			Secret:   secret,
			Events:   events,
			IsActive: isActive,
		}
	}
}

// UpdateWebhookFromDialogCmd creates a command to trigger webhook update from dialog.
func UpdateWebhookFromDialogCmd(webhookID, name, url, secret, events string, isActive bool) tea.Cmd {
	return func() tea.Msg {
		return UpdateWebhookFromDialogRequestMsg{
			WebhookID: webhookID,
			Name:      name,
			URL:       url,
			Secret:    secret,
			Events:    events,
			IsActive:  isActive,
		}
	}
}

// =============================================================================
// CREATE DATATYPE FROM DIALOG
// =============================================================================

// DatatypeCreatedFromDialogMsg is sent after a datatype is successfully created from dialog.
type DatatypeCreatedFromDialogMsg struct {
	DatatypeID types.DatatypeID
	Label      string
}

// CreateDatatypeFromDialogCmd creates a command to create a datatype from dialog input.
func CreateDatatypeFromDialogCmd(name, label, dtype, parentID string) tea.Cmd {
	return func() tea.Msg {
		return CreateDatatypeFromDialogRequestMsg{
			Name:     name,
			Label:    label,
			Type:     dtype,
			ParentID: parentID,
		}
	}
}

// CreateDatatypeFromDialogRequestMsg triggers datatype creation from dialog.
type CreateDatatypeFromDialogRequestMsg struct {
	Name     string
	Label    string
	Type     string
	ParentID string
}

// =============================================================================
// CREATE FIELD FROM DIALOG
// =============================================================================

// FieldCreatedFromDialogMsg is sent after a field is successfully created from dialog.
type FieldCreatedFromDialogMsg struct {
	FieldID    types.FieldID
	DatatypeID types.DatatypeID
	Label      string
}

// CreateFieldFromDialogCmd creates a command to create a field and link it to a datatype.
func CreateFieldFromDialogCmd(name, label, fieldType string, datatypeID types.DatatypeID) tea.Cmd {
	return func() tea.Msg {
		return CreateFieldFromDialogRequestMsg{
			Name:       name,
			Label:      label,
			Type:       fieldType,
			DatatypeID: datatypeID,
		}
	}
}

// CreateFieldFromDialogRequestMsg triggers field creation from dialog.
type CreateFieldFromDialogRequestMsg struct {
	Name       string
	Label      string
	Type       string
	DatatypeID types.DatatypeID
}

// =============================================================================
// CREATE ROUTE FROM DIALOG
// =============================================================================

// RouteCreatedFromDialogMsg is sent after a route is successfully created from dialog.
type RouteCreatedFromDialogMsg struct {
	RouteID types.RouteID
	Title   string
	Slug    string
}

// CreateRouteFromDialogCmd creates a command to create a route from dialog input.
func CreateRouteFromDialogCmd(title, slug string) tea.Cmd {
	return func() tea.Msg {
		return CreateRouteFromDialogRequestMsg{
			Title: title,
			Slug:  slug,
		}
	}
}

// CreateRouteFromDialogRequestMsg triggers route creation from dialog.
type CreateRouteFromDialogRequestMsg struct {
	Title string
	Slug  string
}

// =============================================================================
// CREATE ROUTE WITH CONTENT
// =============================================================================

// RouteWithContentCreatedMsg is sent after a route and initial content are successfully created.
type RouteWithContentCreatedMsg struct {
	RouteID       types.RouteID
	ContentDataID types.ContentID
	DatatypeID    types.DatatypeID
	Title         string
	Slug          string
}

// CreateRouteWithContentRequestMsg triggers route and content creation from dialog.
type CreateRouteWithContentRequestMsg struct {
	Title      string
	Slug       string
	DatatypeID string
}

// CreateRouteWithContentCmd creates a command to create a route with initial content.
func CreateRouteWithContentCmd(title, slug, datatypeID string) tea.Cmd {
	return func() tea.Msg {
		return CreateRouteWithContentRequestMsg{
			Title:      title,
			Slug:       slug,
			DatatypeID: datatypeID,
		}
	}
}

// =============================================================================
// UPDATE DATATYPE FROM DIALOG
// =============================================================================

// DatatypeUpdatedFromDialogMsg is sent after a datatype is successfully updated from dialog.
type DatatypeUpdatedFromDialogMsg struct {
	DatatypeID types.DatatypeID
	Label      string
}

// UpdateDatatypeFromDialogRequestMsg triggers datatype update from dialog.
type UpdateDatatypeFromDialogRequestMsg struct {
	DatatypeID string
	Name       string
	Label      string
	Type       string
	ParentID   string
}

// UpdateDatatypeFromDialogCmd creates a command to update a datatype from dialog input.
func UpdateDatatypeFromDialogCmd(datatypeID, name, label, dtype, parentID string) tea.Cmd {
	return func() tea.Msg {
		return UpdateDatatypeFromDialogRequestMsg{
			DatatypeID: datatypeID,
			Name:       name,
			Label:      label,
			Type:       dtype,
			ParentID:   parentID,
		}
	}
}

// =============================================================================
// UPDATE FIELD FROM DIALOG
// =============================================================================

// FieldUpdatedFromDialogMsg is sent after a field is successfully updated from dialog.
type FieldUpdatedFromDialogMsg struct {
	FieldID    types.FieldID
	DatatypeID types.DatatypeID
	Label      string
}

// UpdateFieldFromDialogRequestMsg triggers field update from dialog.
type UpdateFieldFromDialogRequestMsg struct {
	FieldID string
	Name    string
	Label   string
	Type    string
}

// UpdateFieldFromDialogCmd creates a command to update a field from dialog input.
func UpdateFieldFromDialogCmd(fieldID, name, label, fieldType string) tea.Cmd {
	return func() tea.Msg {
		return UpdateFieldFromDialogRequestMsg{
			FieldID: fieldID,
			Name:    name,
			Label:   label,
			Type:    fieldType,
		}
	}
}

// =============================================================================
// UPDATE ROUTE FROM DIALOG
// =============================================================================

// RouteUpdatedFromDialogMsg is sent after a route is successfully updated from dialog.
type RouteUpdatedFromDialogMsg struct {
	RouteID types.RouteID
	Title   string
	Slug    string
}

// UpdateRouteFromDialogRequestMsg triggers route update from dialog.
type UpdateRouteFromDialogRequestMsg struct {
	RouteID string
	Title   string
	Slug    string
}

// UpdateRouteFromDialogCmd creates a command to update a route from dialog input.
func UpdateRouteFromDialogCmd(routeID, title, slug string) tea.Cmd {
	return func() tea.Msg {
		return UpdateRouteFromDialogRequestMsg{
			RouteID: routeID,
			Title:   title,
			Slug:    slug,
		}
	}
}

// =============================================================================
// INITIALIZE ROUTE CONTENT
// =============================================================================

// RouteContentInitializedMsg is sent after content is successfully initialized for a route.
type RouteContentInitializedMsg struct {
	RouteID       types.RouteID
	ContentDataID types.ContentID
	DatatypeID    types.DatatypeID
	Title         string
}

// InitializeRouteContentRequestMsg triggers content initialization for a route.
type InitializeRouteContentRequestMsg struct {
	RouteID    types.RouteID
	DatatypeID string
}

// InitializeRouteContentCmd creates a command to initialize content for a route.
func InitializeRouteContentCmd(routeID types.RouteID, datatypeID string) tea.Cmd {
	return func() tea.Msg {
		return InitializeRouteContentRequestMsg{
			RouteID:    routeID,
			DatatypeID: datatypeID,
		}
	}
}

// =============================================================================
// FORM DIALOG ACCEPT DISPATCH
// =============================================================================

// handleFormDialogAccept processes FormDialogAcceptMsg by dispatching on the action type.
func (m Model) handleFormDialogAccept(msg FormDialogAcceptMsg) (Model, tea.Cmd) {
	switch msg.Action {
	case FORMDIALOGCREATEDATATYPE:
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			CreateDatatypeFromDialogCmd(msg.Name, msg.Label, msg.Type, msg.ParentID),
		)
	case FORMDIALOGCREATEFIELD:
		// Create a field and link it to the datatype passed via EntityID
		datatypeID := types.DatatypeID(msg.EntityID)
		if datatypeID.IsZero() {
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			CreateFieldFromDialogCmd(msg.Name, msg.Label, msg.Type, datatypeID),
		)
	case FORMDIALOGCREATEROUTE:
		// Create a new route (Label=Title, Type=Slug)
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			CreateRouteFromDialogCmd(msg.Label, msg.Type),
		)
	case FORMDIALOGEDITDATATYPE:
		// Update an existing datatype
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			UpdateDatatypeFromDialogCmd(msg.EntityID, msg.Name, msg.Label, msg.Type, msg.ParentID),
		)
	case FORMDIALOGEDITFIELD:
		// Update an existing field
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			UpdateFieldFromDialogCmd(msg.EntityID, msg.Name, msg.Label, msg.Type),
		)
	case FORMDIALOGEDITROUTE:
		// Update an existing route (Label=Title, Type=Slug)
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			UpdateRouteFromDialogCmd(msg.EntityID, msg.Label, msg.Type),
		)
	case FORMDIALOGCREATEROUTEWITHCONTENT:
		// Create a new route with initial content (ParentID=DatatypeID from carousel, Label=Title, Type=Slug)
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			CreateRouteWithContentCmd(msg.Label, msg.Type, msg.ParentID),
		)
	case FORMDIALOGCREATEADMINROUTEWITHCONTENT:
		// Create a new admin route with initial content (ParentID=AdminDatatypeID from carousel, Label=Title, Type=Slug)
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			CreateAdminRouteWithContentCmd(msg.Label, msg.Type, msg.ParentID),
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
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				ChildDatatypeSelectedCmd(types.DatatypeID(msg.ParentID), types.RouteID(msg.EntityID)),
			)
		}
		if m.Logger != nil {
			m.Logger.Finfo("ParentID was empty, just closing dialog")
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case FORMDIALOGMOVECONTENT:
		// ParentID = selected target content ID, EntityID = "sourceContentID|routeID"
		parts := strings.SplitN(msg.EntityID, "|", 2)
		if len(parts) == 2 && msg.ParentID != "" {
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				MoveContentCmd(types.ContentID(parts[0]), types.ContentID(msg.ParentID), types.RouteID(parts[1])),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case FORMDIALOGADDCONTENTFIELD:
		// ParentID = selected field ID from the picker
		if ctx, ok := m.DCtx.Active.(*addContentFieldCtx); ok && msg.ParentID != "" {
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				m.HandleAddContentField(ctx.ContentID, types.FieldID(msg.ParentID), ctx.RouteID, ctx.DatatypeID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case FORMDIALOGCREATEADMINROUTE:
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			CreateAdminRouteFromDialogCmd(msg.Label, msg.Type),
		)
	case FORMDIALOGEDITADMINROUTE:
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			UpdateAdminRouteFromDialogCmd(msg.EntityID, msg.Label, msg.Type, msg.ParentID),
		)
	case FORMDIALOGCREATEADMINDATATYPE:
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			CreateAdminDatatypeFromDialogCmd(msg.Name, msg.Label, msg.Type, msg.ParentID),
		)
	case FORMDIALOGEDITADMINDATATYPE:
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			UpdateAdminDatatypeFromDialogCmd(msg.EntityID, msg.Name, msg.Label, msg.Type, msg.ParentID),
		)
	case FORMDIALOGCREATEADMINFIELD:
		// Create a field and link it to the admin datatype passed via EntityID
		adminDatatypeID := types.AdminDatatypeID(msg.EntityID)
		if adminDatatypeID.IsZero() {
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			CreateAdminFieldFromDialogCmd(msg.Name, msg.Label, msg.Type, adminDatatypeID),
		)
	case FORMDIALOGEDITADMINFIELD:
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			UpdateAdminFieldFromDialogCmd(msg.EntityID, msg.Name, msg.Label, msg.Type),
		)
	case FORMDIALOGCREATEFIELDTYPE:
		// Create a new field type (Label=Type value, Type=Label value)
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			CreateFieldTypeFromDialogCmd(msg.Label, msg.Type),
		)
	case FORMDIALOGEDITFIELDTYPE:
		// Update an existing field type (Label=Type value, Type=Label value)
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			UpdateFieldTypeFromDialogCmd(msg.EntityID, msg.Label, msg.Type),
		)
	case FORMDIALOGCREATEADMINFIELDTYPE:
		// Create a new admin field type (Label=Type value, Type=Label value)
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			CreateAdminFieldTypeFromDialogCmd(msg.Label, msg.Type),
		)
	case FORMDIALOGEDITADMINFIELDTYPE:
		// Update an existing admin field type (Label=Type value, Type=Label value)
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			UpdateAdminFieldTypeFromDialogCmd(msg.EntityID, msg.Label, msg.Type),
		)
	case FORMDIALOGCREATEVALIDATION:
		// Create a new validation (Label=Name, Type=Description)
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			CreateValidationFromDialogCmd(msg.Label, msg.Type),
		)
	case FORMDIALOGEDITVALIDATION:
		// Update an existing validation (Label=Name, Type=Description)
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			UpdateValidationFromDialogCmd(msg.EntityID, msg.Label, msg.Type),
		)
	case FORMDIALOGCREATEADMINVALIDATION:
		// Create a new admin validation (Label=Name, Type=Description)
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			CreateAdminValidationFromDialogCmd(msg.Label, msg.Type),
		)
	case FORMDIALOGEDITADMINVALIDATION:
		// Update an existing admin validation (Label=Name, Type=Description)
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			UpdateAdminValidationFromDialogCmd(msg.EntityID, msg.Label, msg.Type),
		)
	case FORMDIALOGCONFIGEDIT:
		// EntityID holds the JSON key, Label holds the new value
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			func() tea.Msg {
				return ConfigFieldUpdateMsg{
					Key:   msg.EntityID,
					Value: msg.Label,
				}
			},
		)
	case FORMDIALOGMOVEADMINCONTENT:
		// ParentID = selected target content ID, EntityID = "sourceContentID|routeID"
		parts := strings.SplitN(msg.EntityID, "|", 2)
		if len(parts) == 2 && msg.ParentID != "" {
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				func() tea.Msg {
					return AdminMoveContentRequestMsg{
						SourceID:     types.AdminContentID(parts[0]),
						TargetID:     types.AdminContentID(msg.ParentID),
						AdminRouteID: types.AdminRouteID(parts[1]),
					}
				},
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case FORMDIALOGCHILDADMINDATATYPE:
		// ParentID = selected child datatype ID, EntityID = routeID
		if msg.ParentID != "" {
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				func() tea.Msg {
					return AdminBuildContentFormMsg{
						AdminDatatypeID: types.AdminDatatypeID(msg.ParentID),
						AdminRouteID:    types.AdminRouteID(msg.EntityID),
					}
				},
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case FORMDIALOGADDADMINCONTENTFIELD:
		// ParentID = selected field ID from the picker
		if ctx, ok := m.DCtx.Active.(*addAdminContentFieldCtx); ok && msg.ParentID != "" {
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				m.HandleAddAdminContentField(ctx.AdminContentID, types.AdminFieldID(msg.ParentID), ctx.AdminRouteID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	default:
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	}
}

// handleContentFormDialogAccept processes ContentFormDialogAcceptMsg by dispatching on the action type.
func (m Model) handleContentFormDialogAccept(msg ContentFormDialogAcceptMsg) (Model, tea.Cmd) {
	switch msg.Action {
	case FORMDIALOGEDITCONTENT:
		// Update existing content
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			UpdateContentFromDialogCmd(msg.ContentID, msg.DatatypeID, msg.RouteID, msg.FieldValues),
		)
	case FORMDIALOGEDIITSINGLEFIELD:
		// Single-field edit: use stored context for ContentFieldID
		if ctx, ok := m.DCtx.Active.(*editSingleFieldCtx); ok {
			m.DCtx.Active = nil
			var newValue string
			for _, val := range msg.FieldValues {
				newValue = val
				break
			}
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				m.HandleEditSingleField(
					ctx.ContentFieldID,
					ctx.ContentID, ctx.FieldID, newValue, ctx.RouteID,
					ctx.DatatypeID,
				),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case FORMDIALOGCREATEADMINCONTENT:
		// Convert regular types to admin types and forward
		adminFields := make(map[types.AdminFieldID]string)
		for fid, val := range msg.FieldValues {
			adminFields[types.AdminFieldID(fid)] = val
		}
		adminMsg := AdminContentFormDialogAcceptMsg{
			Action:      FORMDIALOGCREATEADMINCONTENT,
			DatatypeID:  types.AdminDatatypeID(msg.DatatypeID),
			RouteID:     types.AdminRouteID(msg.RouteID),
			FieldValues: adminFields,
		}
		return m.HandleAdminContentFormDialogAccept(adminMsg)
	case FORMDIALOGEDITADMINCONTENT:
		adminFields := make(map[types.AdminFieldID]string)
		for fid, val := range msg.FieldValues {
			adminFields[types.AdminFieldID(fid)] = val
		}
		adminMsg := AdminContentFormDialogAcceptMsg{
			Action:      FORMDIALOGEDITADMINCONTENT,
			DatatypeID:  types.AdminDatatypeID(msg.DatatypeID),
			RouteID:     types.AdminRouteID(msg.RouteID),
			ContentID:   types.AdminContentID(msg.ContentID),
			FieldValues: adminFields,
		}
		return m.HandleAdminContentFormDialogAccept(adminMsg)
	case FORMDIALOGEDITADMINSINGLEFIELD:
		// Admin single-field edit: use stored admin context
		if ctx, ok := m.DCtx.Active.(*editAdminSingleFieldCtx); ok {
			m.DCtx.Active = nil
			var newValue string
			for _, val := range msg.FieldValues {
				newValue = val
				break
			}
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				m.HandleEditAdminSingleField(
					ctx.AdminContentFieldID,
					ctx.AdminContentID,
					ctx.AdminFieldID,
					newValue,
					ctx.AdminRouteID,
					ctx.AdminDatatypeID,
				),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	default:
		// Create new content (FORMDIALOGCREATECONTENT or default)
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			CreateContentFromDialogCmd(msg.DatatypeID, msg.RouteID, msg.ParentID, msg.FieldValues),
		)
	}
}

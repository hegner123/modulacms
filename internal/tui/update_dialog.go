package tui

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// UpdatedDialog signals that a dialog has been updated.
type UpdatedDialog struct{}

// NewDialogUpdate creates a command returning an UpdatedDialog message.
func NewDialogUpdate() tea.Cmd {
	return func() tea.Msg {
		return UpdatedDialog{}
	}
}

// UpdateDialog handles dialog-related messages and state updates.
func (m Model) UpdateDialog(msg tea.Msg) (Model, tea.Cmd) {
	// Dispatch admin-specific dialog messages first.
	if newM, cmd, handled := m.handleAdminDialogMsg(msg); handled {
		return newM, cmd
	}

	// Dispatch CRUD result messages (create/update/delete completions).
	if newM, cmd, handled := m.handleCrudResultMsg(msg); handled {
		return newM, cmd
	}

	switch msg := msg.(type) {
	case DialogReadyOKSet:
		newModel := m
		if d, ok := newModel.ActiveOverlay.(*DialogModel); ok && d != nil {
			d.ReadyOK = msg.Ready
		}
		return newModel, NewDialogUpdate()
	case ShowDialogMsg:
		// Handle showing a dialog
		dialog := NewDialog(msg.Title, msg.Message, msg.ShowCancel, DIALOGDELETE)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowQuitConfirmDialogMsg:
		// Show quit confirmation dialog
		dialog := NewDialog("Quit", "Are you sure you want to quit?", true, DIALOGQUITCONFIRM)
		dialog.SetButtons("Quit", "Cancel")
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowPluginConfirmDialogMsg:
		dialog := NewDialog(msg.Title, msg.Message, true, DIALOGPLUGINCONFIRM)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowPublishDialogMsg:
		// Show publish or unpublish confirmation dialog
		var dialogMsg string
		var dialogAction DialogAction
		var title string
		if msg.IsPublished {
			title = "Unpublish Content"
			dialogMsg = fmt.Sprintf("Unpublish '%s'?\nIt will no longer be publicly accessible.", msg.ContentName)
			dialogAction = DIALOGUNPUBLISHCONTENT
		} else {
			title = "Publish Content"
			dialogMsg = fmt.Sprintf("Publish '%s'?\nThis creates a public snapshot.", msg.ContentName)
			dialogAction = DIALOGPUBLISHCONTENT
		}
		dialog := NewDialog(title, dialogMsg, true, dialogAction)
		if msg.IsPublished {
			dialog.SetButtons("Unpublish", "Cancel")
		} else {
			dialog.SetButtons("Publish", "Cancel")
		}
		m.DCtx.Active = &PublishContentContext{
			ContentID: msg.ContentID,
			RouteID:   msg.RouteID,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case LocaleListMsg:
		if msg.Err != nil {
			return m, ShowDialog("Error", fmt.Sprintf("Failed to load locales: %v", msg.Err), false)
		}
		if len(msg.Locales) == 0 {
			return m, ShowDialog("Info", "No enabled locales configured", false)
		}
		// Build the locale code list for the selection dialog
		locales := make([]string, len(msg.Locales))
		for i, loc := range msg.Locales {
			locales[i] = loc.Code + " - " + loc.Label
		}
		// Find current locale index
		startIdx := 0
		for i, loc := range msg.Locales {
			if loc.Code == m.ActiveLocale {
				startIdx = i
				break
			}
		}
		dialogMessage := buildLocaleDialogMessage(locales, startIdx)
		dialog := NewDialog("Select Locale", dialogMessage, false, DIALOGLOCALESELECT)
		dialog.Locales = locales
		dialog.ActionIndex = startIdx
		dialog.SetButtons("Select", "Cancel")
		dialog.ShowCancel = true
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)

	case LocaleSwitchMsg:
		m.ActiveLocale = msg.Locale
		// NOTE: Content field reload for the new locale is handled by the
		// ContentScreen via its own Update(). This Model-level handler only
		// persists the locale choice.
		return m, nil

	case ShowDeleteContentDialogMsg:
		// Show delete content confirmation dialog
		var dialogMsg string
		if msg.HasChildren {
			dialogMsg = fmt.Sprintf("Cannot delete '%s' because it has children.\nDelete child nodes first.", msg.ContentName)
			dialog := NewDialog("Cannot Delete", dialogMsg, false, DIALOGGENERIC)
			return m, tea.Batch(
				OverlaySetCmd(&dialog),
				FocusSetCmd(DIALOGFOCUS),
			)
		}
		dialogMsg = fmt.Sprintf("Delete '%s'?\nThis will also delete all field values.", msg.ContentName)
		dialog := NewDialog("Delete Content", dialogMsg, true, DIALOGDELETECONTENT)
		dialog.SetButtons("Delete", "Cancel")
		// Store the content ID for deletion
		m.DCtx.Active = &DeleteContentContext{
			ContentID: msg.ContentID,
			RouteID:   string(m.PageRouteId),
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
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
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case QuickstartConfirmMsg:
		// Show confirmation dialog for quickstart schema install
		labels := QuickstartMenuLabels()
		label := "this schema"
		if msg.SchemaIndex < len(labels) {
			label = labels[msg.SchemaIndex]
		}
		dialog := NewDialog(
			"Install Schema",
			fmt.Sprintf("Install %s?\nThis will create datatypes and fields in the database.", label),
			true,
			DIALOGQUICKSTART,
		)
		dialog.SetButtons("Install", "Cancel")
		dialog.ActionIndex = msg.SchemaIndex
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ActionResultMsg:
		// Signal serve to reload permission cache and start HTTP servers
		if msg.ReloadPermissions && !msg.IsError && m.DBReadyCh != nil {
			select {
			case m.DBReadyCh <- struct{}{}:
			default:
			}
		}
		// Show result dialog after an action completes
		dialog := NewDialog(msg.Title, msg.Message, false, DIALOGGENERIC)
		if msg.Width > 0 {
			dialog.Width = msg.Width
		}
		return m, tea.Batch(
			LoadingStopCmd(),
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case OpenFilePickerForRestoreMsg:
		// Open file picker for selecting a backup archive
		fp := filepicker.New()
		fp.AllowedTypes = []string{".zip"}
		fp.CurrentDirectory, _ = os.UserHomeDir()
		fp.AutoHeight = false
		fp.SetHeight(filePickerHeight(m.Height))
		m.FilePicker = fp
		m.FilePickerActive = true
		m.FilePickerPurpose = FILEPICKER_RESTORE
		return m, m.FilePicker.Init()
	case RestoreBackupFromPathMsg:
		// Show confirmation before restoring
		dialog := NewDialog(
			"Restore Backup",
			fmt.Sprintf("Restore from:\n%s\n\nThis will replace the current database. Continue?", msg.Path),
			true,
			DIALOGBACKUPRESTORE,
		)
		dialog.SetButtons("Restore", "Cancel")
		m.DCtx.Active = &RestoreBackupContext{Path: msg.Path}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case BackupRestoreCompleteMsg:
		// Show result dialog and quit on dismiss
		dialog := NewDialog(
			"Restore Complete",
			fmt.Sprintf("Backup restored from:\n%s\n\nThe application will exit. Please restart.", msg.Path),
			false,
			DIALOGGENERIC,
		)
		m.DCtx.RestoreRequiresQuit = true
		return m, tea.Batch(
			LoadingStopCmd(),
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteDatatypeDialogMsg:
		if msg.HasChildren {
			dialog := NewDialog("Cannot Delete", fmt.Sprintf("Cannot delete '%s' because it has child datatypes.\nDelete child datatypes first.", msg.Label), false, DIALOGGENERIC)
			return m, tea.Batch(
				OverlaySetCmd(&dialog),
				FocusSetCmd(DIALOGFOCUS),
			)
		}
		dialog := NewDialog("Delete Datatype", fmt.Sprintf("Delete datatype '%s'?\nThis will remove all field associations.", msg.Label), true, DIALOGDELETEDATATYPE)
		dialog.SetButtons("Delete", "Cancel")
		m.DCtx.Active = &DeleteDatatypeContext{
			DatatypeID: msg.DatatypeID,
			Label:      msg.Label,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteFieldDialogMsg:
		dialog := NewDialog("Delete Field", fmt.Sprintf("Delete field '%s'?\nThis will unlink it from the datatype and remove the field.", msg.Label), true, DIALOGDELETEFIELD)
		dialog.SetButtons("Delete", "Cancel")
		m.DCtx.Active = &DeleteFieldContext{
			FieldID:    msg.FieldID,
			DatatypeID: msg.DatatypeID,
			Label:      msg.Label,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteRouteDialogMsg:
		dialog := NewDialog("Delete Route", fmt.Sprintf("Delete route '%s'?\nAssociated content will also be removed.", msg.Title), true, DIALOGDELETEROUTE)
		dialog.SetButtons("Delete", "Cancel")
		m.DCtx.Active = &DeleteRouteContext{
			RouteID: msg.RouteID,
			Title:   msg.Title,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteMediaDialogMsg:
		dialog := NewDialog("Delete Media", fmt.Sprintf("Delete media '%s'?\nThis cannot be undone.", msg.Label), true, DIALOGDELETEMEDIA)
		dialog.SetButtons("Delete", "Cancel")
		m.DCtx.Active = &DeleteMediaContext{
			MediaID: msg.MediaID,
			Label:   msg.Label,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteUserDialogMsg:
		dialog := NewDialog("Delete User", fmt.Sprintf("Delete user '%s'?\nThis cannot be undone.", msg.Username), true, DIALOGDELETEUSER)
		dialog.SetButtons("Delete", "Cancel")
		m.DCtx.Active = &DeleteUserContext{
			UserID:   msg.UserID,
			Username: msg.Username,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowUserFormDialogMsg:
		dialog := NewUserFormDialog(msg.Title, msg.Roles)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditUserDialogMsg:
		dialog := NewEditUserFormDialog("Edit User", msg.User, msg.Roles)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditSingleFieldDialogMsg:
		// Store context for when the dialog is accepted
		m.DCtx.Active = &editSingleFieldCtx{
			ContentFieldID: msg.Field.ContentFieldID,
			ContentID:      msg.ContentID,
			FieldID:        msg.Field.FieldID,
			RouteID:        msg.RouteID,
			DatatypeID:     msg.DatatypeID,
		}
		// Create a content form dialog with a single field, pre-populated with current value
		existingFields := []ExistingContentField{{
			ContentFieldID: msg.Field.ContentFieldID,
			FieldID:        msg.Field.FieldID,
			Label:          msg.Field.Label,
			Type:           msg.Field.Type,
			Value:          msg.Field.Value,
			ValidationJSON: msg.Field.ValidationJSON,
			DataJSON:       msg.Field.DataJSON,
		}}
		dialog := NewEditContentFormDialog(
			fmt.Sprintf("Edit: %s", msg.Field.Label),
			msg.ContentID,
			types.DatatypeID(msg.DatatypeID.ID),
			msg.RouteID,
			existingFields,
		)
		dialog.Action = FORMDIALOGEDIITSINGLEFIELD
		dialog.Logger = m.Logger
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowAddContentFieldDialogMsg:
		// Store context for when the dialog is accepted
		m.DCtx.Active = &addContentFieldCtx{
			ContentID:  msg.ContentID,
			RouteID:    msg.RouteID,
			DatatypeID: msg.DatatypeID,
		}
		// Create a picker dialog to select which missing field to add
		parents := make([]ParentOption, 0, len(msg.Options))
		for _, opt := range msg.Options {
			parents = append(parents, ParentOption{
				Label: opt.Key,
				Value: opt.Value,
			})
		}
		dialog := FormDialogModel{
			dialogStyles:  newDialogStyles(),
			Title:         "Add Field",
			Width:         50,
			Action:        FORMDIALOGADDCONTENTFIELD,
			LabelInput:    textinput.New(),
			TypeInput:     textinput.New(),
			ParentOptions: parents,
			ParentIndex:   0,
			focusIndex:    FormDialogFieldParent,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteContentFieldDialogMsg:
		// Show delete confirmation for a content field
		dialog := NewDialog(
			"Delete Field Value",
			fmt.Sprintf("Delete value for field '%s'?\nThis removes the stored value.", msg.Field.Label),
			true,
			DIALOGDELETECONTENTFIELD,
		)
		dialog.SetButtons("Delete", "Cancel")
		m.DCtx.Active = &DeleteContentFieldContext{
			ContentFieldID: msg.Field.ContentFieldID,
			ContentID:      msg.ContentID,
			RouteID:        msg.RouteID,
			DatatypeID:     msg.DatatypeID,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowConfigFieldEditMsg:
		// Create a simple form dialog for editing a config field value
		labelInput := textinput.New()
		labelInput.Placeholder = "Value"
		labelInput.CharLimit = 512
		labelInput.SetWidth(50)
		labelInput.SetValue(msg.CurrentValue)
		labelInput.Focus()

		dialog := FormDialogModel{
			dialogStyles: newDialogStyles(),
			Title:        fmt.Sprintf("Edit: %s", msg.Field.Label),
			Width:        60,
			Action:       FORMDIALOGCONFIGEDIT,
			EntityID:     msg.Field.JSONKey,
			LabelInput:   labelInput,
			TypeInput:    textinput.New(),
			focusIndex:   FormDialogFieldLabel,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowApproveAllRoutesDialogMsg:
		// Show confirmation dialog listing pending routes before approving
		routeList := strings.Join(msg.PendingRoutes, "\n  ")
		dialogMsg := fmt.Sprintf("Approve %d routes for plugin '%s'?\n\n  %s", len(msg.PendingRoutes), msg.PluginName, routeList)
		dialog := NewDialog("Approve Plugin Routes", dialogMsg, true, DIALOGAPPROVEPLUGINROUTES)
		dialog.SetButtons("Approve", "Cancel")
		m.DCtx.Active = &ApprovePluginContext{PluginName: msg.PluginName}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowApproveAllHooksDialogMsg:
		// Show confirmation dialog listing pending hooks before approving
		hookList := strings.Join(msg.PendingHooks, "\n  ")
		dialogMsg := fmt.Sprintf("Approve %d hooks for plugin '%s'?\n\n  %s", len(msg.PendingHooks), msg.PluginName, hookList)
		dialog := NewDialog("Approve Plugin Hooks", dialogMsg, true, DIALOGAPPROVEPLUGINSHOOKS)
		dialog.SetButtons("Approve", "Cancel")
		m.DCtx.Active = &ApprovePluginContext{PluginName: msg.PluginName}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteAdminRouteDialogMsg:
		dialog := NewDialog("Delete Admin Route", fmt.Sprintf("Delete admin route '%s'?\nThis cannot be undone.", msg.Title), true, DIALOGDELETEADMINROUTE)
		dialog.SetButtons("Delete", "Cancel")
		m.DCtx.Active = &DeleteAdminRouteContext{
			AdminRouteID: msg.AdminRouteID,
			Title:        msg.Title,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteAdminDatatypeDialogMsg:
		if msg.HasChildren {
			dialog := NewDialog("Cannot Delete", fmt.Sprintf("Cannot delete '%s' because it has child datatypes.\nDelete child datatypes first.", msg.Label), false, DIALOGGENERIC)
			return m, tea.Batch(
				OverlaySetCmd(&dialog),
				FocusSetCmd(DIALOGFOCUS),
			)
		}
		dialog := NewDialog("Delete Admin Datatype", fmt.Sprintf("Delete admin datatype '%s'?\nThis will remove all field associations.", msg.Label), true, DIALOGDELETEADMINDATATYPE)
		dialog.SetButtons("Delete", "Cancel")
		m.DCtx.Active = &DeleteAdminDatatypeContext{
			AdminDatatypeID: msg.AdminDatatypeID,
			Label:           msg.Label,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteAdminFieldDialogMsg:
		dialog := NewDialog("Delete Admin Field", fmt.Sprintf("Delete admin field '%s'?\nThis will unlink it from the datatype and remove the field.", msg.Label), true, DIALOGDELETEADMINFIELD)
		dialog.SetButtons("Delete", "Cancel")
		m.DCtx.Active = &DeleteAdminFieldContext{
			AdminFieldID:    msg.AdminFieldID,
			AdminDatatypeID: msg.AdminDatatypeID,
			Label:           msg.Label,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteFieldTypeDialogMsg:
		dialog := NewDialog("Delete Field Type", fmt.Sprintf("Delete field type '%s'?\nThis cannot be undone.", msg.Label), true, DIALOGDELETEFIELDTYPE)
		dialog.SetButtons("Delete", "Cancel")
		m.DCtx.Active = &DeleteFieldTypeContext{
			FieldTypeID: msg.FieldTypeID,
			Label:       msg.Label,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditFieldTypeDialogMsg:
		// Edit field type dialog with pre-populated values
		dialog := NewRouteFormDialog("Edit Field Type", FORMDIALOGEDITFIELDTYPE)
		dialog.LabelInput.SetValue(msg.FieldType.Type)
		dialog.TypeInput.SetValue(msg.FieldType.Label)
		dialog.EntityID = string(msg.FieldType.FieldTypeID)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteAdminFieldTypeDialogMsg:
		dialog := NewDialog("Delete Admin Field Type", fmt.Sprintf("Delete admin field type '%s'?\nThis cannot be undone.", msg.Label), true, DIALOGDELETEADMINFIELDTYPE)
		dialog.SetButtons("Delete", "Cancel")
		m.DCtx.Active = &DeleteAdminFieldTypeContext{
			AdminFieldTypeID: msg.AdminFieldTypeID,
			Label:            msg.Label,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditAdminFieldTypeDialogMsg:
		// Edit admin field type dialog with pre-populated values
		dialog := NewRouteFormDialog("Edit Admin Field Type", FORMDIALOGEDITADMINFIELDTYPE)
		dialog.LabelInput.SetValue(msg.AdminFieldType.Type)
		dialog.TypeInput.SetValue(msg.AdminFieldType.Label)
		dialog.EntityID = string(msg.AdminFieldType.AdminFieldTypeID)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteValidationDialogMsg:
		dialog := NewDialog("Delete Validation", fmt.Sprintf("Delete validation '%s'?\nFields using this validation will lose their reference.\nThis cannot be undone.", msg.Name), true, DIALOGDELETEVALIDATION)
		dialog.SetButtons("Delete", "Cancel")
		m.DCtx.Active = &DeleteValidationContext{
			ValidationID: msg.ValidationID,
			Name:         msg.Name,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditValidationDialogMsg:
		// Edit validation dialog with pre-populated values
		// Label field = Name, Type field = Description
		dialog := NewRouteFormDialog("Edit Validation", FORMDIALOGEDITVALIDATION)
		dialog.LabelInput.SetValue(msg.Validation.Name)
		dialog.TypeInput.SetValue(msg.Validation.Description)
		dialog.EntityID = string(msg.Validation.ValidationID)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteAdminValidationDialogMsg:
		dialog := NewDialog("Delete Admin Validation", fmt.Sprintf("Delete admin validation '%s'?\nFields using this validation will lose their reference.\nThis cannot be undone.", msg.Name), true, DIALOGDELETEADMINVALIDATION)
		dialog.SetButtons("Delete", "Cancel")
		m.DCtx.Active = &DeleteAdminValidationContext{
			AdminValidationID: msg.AdminValidationID,
			Name:              msg.Name,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditAdminValidationDialogMsg:
		// Edit admin validation dialog with pre-populated values
		// Label field = Name, Type field = Description
		dialog := NewRouteFormDialog("Edit Admin Validation", FORMDIALOGEDITADMINVALIDATION)
		dialog.LabelInput.SetValue(msg.AdminValidation.Name)
		dialog.TypeInput.SetValue(msg.AdminValidation.Description)
		dialog.EntityID = string(msg.AdminValidation.AdminValidationID)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	// =========================================================================
	// DEPLOY CONFIRMATION DIALOGS
	// =========================================================================
	case DeployConfirmPullMsg:
		dialog := NewDialog("Pull from "+msg.EnvName, "This will overwrite local data with data\nfrom the remote environment.\n\nProceed?", true, DIALOGDEPLOYPULL)
		dialog.SetButtons("Pull", "Cancel")
		m.DCtx.Active = &DeployPullContext{EnvName: msg.EnvName}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case DeployConfirmPushMsg:
		dialog := NewDialog("Push to "+msg.EnvName, "This will overwrite remote data with\nlocal data.\n\nProceed?", true, DIALOGDEPLOYPUSH)
		dialog.SetButtons("Push", "Cancel")
		m.DCtx.Active = &DeployPushContext{EnvName: msg.EnvName}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)

	case UserFormDialogAcceptMsg:
		switch msg.Action {
		case FORMDIALOGCREATEUSER:
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				CreateUserFromDialogCmd(msg.Username, msg.Name, msg.Email, msg.Password, msg.Role),
			)
		case FORMDIALOGEDITUSER:
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				UpdateUserFromDialogCmd(msg.EntityID, msg.Username, msg.Name, msg.Email, msg.Role),
			)
		default:
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
			)
		}
	case UserFormDialogCancelMsg:
		m.DCtx.Active = nil
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case ShowWebhookFormDialogMsg:
		dialog := NewWebhookFormDialog(msg.Title)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditWebhookDialogMsg:
		dialog := NewEditWebhookFormDialog("Edit Webhook", msg.Webhook)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteWebhookDialogMsg:
		dialog := NewDialog("Delete Webhook", fmt.Sprintf("Delete webhook '%s'?\nThis cannot be undone.", msg.Name), true, DIALOGDELETEWEBHOOK)
		dialog.SetButtons("Delete", "Cancel")
		m.DCtx.Active = &DeleteWebhookContext{
			WebhookID: msg.WebhookID,
			Name:      msg.Name,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case WebhookFormDialogAcceptMsg:
		switch msg.Action {
		case FORMDIALOGCREATEWEBHOOK:
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				CreateWebhookFromDialogCmd(msg.Name, msg.URL, msg.Secret, msg.Events, msg.IsActive),
			)
		case FORMDIALOGEDITWEBHOOK:
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				UpdateWebhookFromDialogCmd(msg.EntityID, msg.Name, msg.URL, msg.Secret, msg.Events, msg.IsActive),
			)
		default:
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
			)
		}
	case WebhookFormDialogCancelMsg:
		m.DCtx.Active = nil
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)

	// --- Media folder dialog messages ---
	case ShowCreateMediaFolderDialogMsg:
		dialog := NewCreateFolderDialog(msg.ParentID)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowRenameMediaFolderDialogMsg:
		dialog := NewRenameFolderDialog(msg.FolderID, msg.CurrentName)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteMediaFolderDialogMsg:
		dialog := NewDialog("Delete Folder",
			fmt.Sprintf("Delete folder '%s'?\nFolder must be empty (no files or subfolders).", msg.Name),
			true, DIALOGDELETEMEDIAFOLDER)
		dialog.SetButtons("Delete", "Cancel")
		m.DCtx.Active = &DeleteMediaFolderContext{
			FolderID: msg.FolderID,
			Name:     msg.Name,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowMoveMediaToFolderDialogMsg:
		// Need to fetch current folders to build the picker list.
		// Use DB from model to get folder list synchronously in a command.
		d := m.DB
		if d == nil {
			dialog := NewDialog("Error", "Database not connected.", false, DIALOGGENERIC)
			return m, tea.Batch(
				OverlaySetCmd(&dialog),
				FocusSetCmd(DIALOGFOCUS),
			)
		}
		mediaID := msg.MediaID
		label := msg.Label
		return m, func() tea.Msg {
			folders, err := d.ListMediaFolders()
			if err != nil {
				return ActionResultMsg{
					Title:   "Error",
					Message: fmt.Sprintf("Failed to load folders: %v", err),
				}
			}
			var folderData []db.MediaFolder
			if folders != nil {
				folderData = *folders
			}
			return ShowMoveMediaToFolderPickerMsg{
				MediaID: mediaID,
				Label:   label,
				Folders: folderData,
			}
		}
	case ShowMoveMediaToFolderPickerMsg:
		dialog := NewMoveMediaFolderDialog(msg.MediaID, msg.Label, msg.Folders)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case MediaFolderNameDialogCancelMsg:
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case MoveMediaFolderDialogCancelMsg:
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case CreateMediaFolderRequestMsg:
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			m.HandleCreateMediaFolder(msg),
		)
	case RenameMediaFolderRequestMsg:
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			m.HandleRenameMediaFolder(msg),
		)
	case MoveMediaToFolderRequestMsg:
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			m.HandleMoveMediaToFolder(msg),
		)
	case MediaFolderCreatedMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Folder created: %s", msg.Name)),
			MediaFetchCmd(),
		)
	case MediaFolderRenamedMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Folder renamed to: %s", msg.NewName)),
			MediaFetchCmd(),
		)
	case MediaFolderDeletedMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Folder deleted: %s", msg.FolderID)),
			MediaFetchCmd(),
		)
	case MediaMovedToFolderMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Media moved: %s", msg.MediaID)),
			MediaFetchCmd(),
		)

	case DialogAcceptMsg:
		return m.handleDialogAccept(msg)
	case DialogCancelMsg:
		// If a restore just completed, quit the application
		if m.DCtx.RestoreRequiresQuit {
			m.DCtx.RestoreRequiresQuit = false
			return m, tea.Quit
		}
		// Plugin confirm dialogs send a cancel response to the coroutine.
		if d, ok := m.ActiveOverlay.(*DialogModel); ok && d != nil && d.Action == DIALOGPLUGINCONFIRM {
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				func() tea.Msg { return PluginDialogResponseMsg{Accepted: false} },
			)
		}
		// Handle dialog cancel action
		m.DCtx.Active = nil
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)

	// Form dialog handling
	case ShowFormDialogMsg:
		dialog := NewFormDialog(msg.Title, msg.Action, msg.Parents)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowFieldFormDialogMsg:
		// Field form dialog has no parent selector.
		// ContextID carries the parent datatype ID for field creation.
		dialog := NewFieldFormDialog(msg.Title, msg.Action)
		dialog.EntityID = msg.ContextID
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowRouteFormDialogMsg:
		// Route form dialog has Title and Slug inputs
		dialog := NewRouteFormDialog(msg.Title, msg.Action)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditDatatypeDialogMsg:
		// Edit datatype dialog with pre-populated values
		dialog := NewEditDatatypeDialog("Edit Datatype", FORMDIALOGEDITDATATYPE, msg.Parents, msg.Datatype)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditFieldDialogMsg:
		// Edit field dialog with pre-populated values
		dialog := NewEditFieldDialog("Edit Field", FORMDIALOGEDITFIELD, msg.Field)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditRouteDialogMsg:
		// Edit route dialog with pre-populated values
		dialog := NewEditRouteDialog("Edit Route", FORMDIALOGEDITROUTE, msg.Route)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowAdminFormDialogMsg:
		// Admin form dialog with parent options from admin datatypes
		parentOpts := []ParentOption{
			{Label: "_root (no parent)", Value: ""},
		}
		for _, p := range msg.Parents {
			parentOpts = append(parentOpts, ParentOption{
				Label: p.Label,
				Value: string(p.AdminDatatypeID),
			})
		}
		dialog := NewFormDialog(msg.Title, msg.Action, nil)
		dialog.ParentOptions = parentOpts
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditAdminRouteDialogMsg:
		// Edit admin route dialog with pre-populated values
		dialog := NewRouteFormDialog("Edit Admin Route", FORMDIALOGEDITADMINROUTE)
		dialog.LabelInput.SetValue(msg.Route.Title)
		dialog.TypeInput.SetValue(string(msg.Route.Slug))
		dialog.EntityID = string(msg.Route.AdminRouteID)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditAdminDatatypeDialogMsg:
		// Edit admin datatype dialog with pre-populated values
		parentOpts := []ParentOption{
			{Label: "_root (no parent)", Value: ""},
		}
		for _, p := range msg.Parents {
			if p.AdminDatatypeID == msg.Datatype.AdminDatatypeID {
				continue // skip self to prevent circular reference
			}
			parentOpts = append(parentOpts, ParentOption{
				Label: p.Label,
				Value: string(p.AdminDatatypeID),
			})
		}
		dialog := NewFormDialog("Edit Admin Datatype", FORMDIALOGEDITADMINDATATYPE, nil)
		dialog.ParentOptions = parentOpts
		dialog.NameInput.SetValue(msg.Datatype.Name)
		dialog.LabelInput.SetValue(msg.Datatype.Label)
		dialog.TypeInput.SetValue(msg.Datatype.Type)
		dialog.EntityID = string(msg.Datatype.AdminDatatypeID)
		if msg.Datatype.ParentID.Valid {
			for i, opt := range dialog.ParentOptions {
				if opt.Value == string(msg.Datatype.ParentID.ID) {
					dialog.ParentIndex = i
					break
				}
			}
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditAdminFieldDialogMsg:
		// Edit admin field dialog with pre-populated values
		dialog := NewFieldFormDialog("Edit Admin Field", FORMDIALOGEDITADMINFIELD)
		dialog.NameInput.SetValue(msg.Field.Name)
		dialog.LabelInput.SetValue(msg.Field.Label)
		dialog.TypeIndex = FieldInputTypeIndex(string(msg.Field.Type))
		dialog.EntityID = string(msg.Field.AdminFieldID)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowCreateRouteWithContentDialogMsg:
		// Create route with initial content dialog
		dialog := NewRouteWithContentDialog("New Content", FORMDIALOGCREATEROUTEWITHCONTENT, msg.RootDatatypes)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowCreateAdminRouteWithContentDialogMsg:
		// Create admin route with initial content dialog
		dialog := NewAdminRouteWithContentDialog("New Admin Content", FORMDIALOGCREATEADMINROUTEWITHCONTENT, msg.AdminRootDatatypes)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
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
		m.DCtx.Active = &InitializeRouteContentContext{
			Route:      msg.Route,
			DatatypeID: msg.DatatypeID,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowChildDatatypeDialogMsg:
		// Show dialog to select a child datatype for creating new content
		dialog := NewChildDatatypeDialog("Select Child Type", msg.ChildDatatypes, string(msg.RouteID))
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowMoveContentDialogMsg:
		sourceID := string(msg.SourceNode.Instance.ContentDataID)
		dialog := NewMoveContentDialog("Move Content", sourceID, string(msg.RouteID), msg.ValidTargets)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case FormDialogAcceptMsg:
		return m.handleFormDialogAccept(msg)
	case FormDialogCancelMsg:
		m.DCtx.Active = nil
		return m, tea.Batch(
			OverlayClearCmd(),
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
		dialog.Logger = logger
		logger.Finfo(fmt.Sprintf("ContentFormDialogModel created with %d fields", len(dialog.Fields)))
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ContentFormDialogAcceptMsg:
		return m.handleContentFormDialogAccept(msg)
	case ContentFormDialogCancelMsg:
		m.DCtx.Active = nil
		return m, tea.Batch(
			OverlayClearCmd(),
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
		dialog.Logger = logger
		logger.Finfo(fmt.Sprintf("EditContentFormDialogModel created with %d fields", len(dialog.Fields)))
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	// =========================================================================
	// DATABASE FORM DIALOG
	// =========================================================================
	// =========================================================================
	// UICONFIG FORM DIALOG
	// =========================================================================
	case ShowUIConfigFormDialogMsg:
		dialog := NewUIConfigFormDialog(msg.Title, msg.FieldID)
		m.ActiveOverlay = &dialog
		return m, FocusSetCmd(DIALOGFOCUS)
	case ShowEditUIConfigFormDialogMsg:
		dialog := NewEditUIConfigFormDialog(msg.Title, msg.FieldID, msg.Existing)
		m.ActiveOverlay = &dialog
		return m, FocusSetCmd(DIALOGFOCUS)
	case UIConfigFormDialogCancelMsg:
		m.ActiveOverlay = nil
		m.DCtx.Active = nil
		return m, FocusSetCmd(PAGEFOCUS)
	case UIConfigFormDialogAcceptMsg:
		m.ActiveOverlay = nil
		uiJSON := marshalUIConfig(msg.Widget, msg.Placeholder, msg.HelpText, msg.Hidden)
		return m, tea.Batch(
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			UpdateFieldUIConfigCmd(msg.FieldID, uiJSON),
		)
	case FieldUIConfigUpdatedMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("UIConfig updated for field %s", msg.FieldID)),
			DatatypeFieldsFetchCmd(msg.DatatypeID),
		)

	// =========================================================================
	// DATABASE FORM DIALOG
	// =========================================================================
	case ShowDatabaseFormDialogMsg:
		// Determine columns: prefer Columns, fall back to Headers
		var columns []string
		if m.TableState.Columns != nil {
			columns = *m.TableState.Columns
		} else if len(m.TableState.Headers) > 0 {
			columns = m.TableState.Headers
		} else {
			// Last resort: derive from GenericHeaders
			dbt := db.StringDBTable(m.TableState.Table)
			columns = db.GenericHeaders(dbt)
		}
		if len(columns) == 0 {
			return m, LogMessageCmd("No column metadata available")
		}

		switch msg.Action {
		case FORMDIALOGDBINSERT:
			dialog := NewDatabaseInsertDialog(msg.Title, msg.Table, columns, nil)
			m.ActiveOverlay = &dialog
			return m, tea.Batch(
				FocusSetCmd(DIALOGFOCUS),
				StatusSetCmd(EDITING),
			)
		case FORMDIALOGDBUPDATE:
			// Get current row data. m.Cursor is set to the absolute row index
			// by the screen's CursorSetCmd before this message arrives.
			if len(m.TableState.Rows) == 0 {
				return m, LogMessageCmd("No rows available for update")
			}
			if m.Cursor >= len(m.TableState.Rows) {
				return m, LogMessageCmd("Row index out of range")
			}
			currentRow := m.TableState.Rows[m.Cursor]
			dialog := NewDatabaseUpdateDialog(msg.Title, msg.Table, columns, nil, currentRow)
			m.ActiveOverlay = &dialog
			return m, tea.Batch(
				FocusSetCmd(DIALOGFOCUS),
				StatusSetCmd(EDITING),
			)
		default:
			return m, LogMessageCmd(fmt.Sprintf("Unknown database form action: %s", msg.Action))
		}
	case DatabaseFormDialogAcceptMsg:
		m.ActiveOverlay = nil

		switch msg.Action {
		case FORMDIALOGDBINSERT:
			// Build columns and values for insert
			var insertColumns []string
			var insertValues []*string
			for i, col := range msg.Columns {
				if isAutoFillColumn(col) {
					insertColumns = append(insertColumns, col)
					insertValues = append(insertValues, nil)
					continue
				}
				insertColumns = append(insertColumns, col)
				val := msg.Values[i]
				insertValues = append(insertValues, &val)
			}
			return m, tea.Batch(
				FocusSetCmd(PAGEFOCUS),
				StatusSetCmd(OK),
				LoadingStartCmd(),
				DatabaseInsertCmd(msg.Table, insertColumns, insertValues),
			)
		case FORMDIALOGDBUPDATE:
			// Build values map for update (non-auto columns only)
			valuesMap := make(map[string]string)
			for i, col := range msg.Columns {
				if isAutoFillColumn(col) {
					continue
				}
				valuesMap[col] = msg.Values[i]
			}
			return m, tea.Batch(
				FocusSetCmd(PAGEFOCUS),
				StatusSetCmd(OK),
				LoadingStartCmd(),
				DatabaseUpdateEntryCmd(msg.Table, msg.RowID, valuesMap),
			)
		default:
			return m, tea.Batch(
				FocusSetCmd(PAGEFOCUS),
				StatusSetCmd(OK),
			)
		}
	case DatabaseFormDialogCancelMsg:
		m.ActiveOverlay = nil
		m.DCtx.Active = nil
		return m, tea.Batch(
			FocusSetCmd(PAGEFOCUS),
			StatusSetCmd(OK),
		)

	case ConfigFieldUpdateMsg:
		// Apply config field update via the config manager
		if m.ConfigManager == nil {
			dialog := NewDialog("Error", "Config manager is not available.", false, DIALOGGENERIC)
			return m, tea.Batch(
				OverlaySetCmd(&dialog),
				FocusSetCmd(DIALOGFOCUS),
			)
		}
		key := msg.Key
		value := msg.Value
		return m, func() tea.Msg {
			updates := map[string]any{key: value}
			result, err := m.ConfigManager.Update(updates)
			if err != nil {
				return ConfigUpdateResultMsg{Err: err}
			}
			if !result.Valid {
				return ConfigUpdateResultMsg{
					Err: fmt.Errorf("validation failed: %s", strings.Join(result.Errors, "; ")),
				}
			}
			return ConfigUpdateResultMsg{RestartRequired: result.RestartRequired}
		}
	case ConfigUpdateResultMsg:
		// Refresh the config snapshot on the model
		if m.ConfigManager != nil {
			snapshot, snapErr := m.ConfigManager.Snapshot()
			if snapErr == nil {
				m.Config = &snapshot
			}
		}
		if msg.Err != nil {
			dialog := NewDialog("Config Update Failed", msg.Err.Error(), false, DIALOGGENERIC)
			return m, tea.Batch(
				OverlaySetCmd(&dialog),
				FocusSetCmd(DIALOGFOCUS),
			)
		}
		if len(msg.RestartRequired) > 0 {
			fields := strings.Join(msg.RestartRequired, ", ")
			dialog := NewDialog("Config Updated", fmt.Sprintf("Config saved. The following fields require a restart to take effect:\n\n%s", fields), false, DIALOGGENERIC)
			return m, tea.Batch(
				OverlaySetCmd(&dialog),
				FocusSetCmd(DIALOGFOCUS),
			)
		}
		dialog := NewDialog("Config Updated", "Config saved successfully.", false, DIALOGGENERIC)
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		)
	default:
		utility.DefaultLogger.Fdebug(fmt.Sprintf("UpdateDialog: unhandled message type %T", msg))
		return m, nil
	}
}

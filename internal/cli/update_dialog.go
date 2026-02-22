package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/hegner123/modulacms/internal/backup"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
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
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
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
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case OpenFilePickerForRestoreMsg:
		// Open file picker for selecting a backup archive
		fp := filepicker.New()
		fp.AllowedTypes = []string{".zip"}
		fp.CurrentDirectory, _ = os.UserHomeDir()
		fp.Height = m.Height - 4
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
		restoreBackupContext = &RestoreBackupContext{Path: msg.Path}
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
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
		restoreRequiresQuit = true
		return m, tea.Batch(
			LoadingStopCmd(),
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteDatatypeDialogMsg:
		if msg.HasChildren {
			dialog := NewDialog("Cannot Delete", fmt.Sprintf("Cannot delete '%s' because it has child datatypes.\nDelete child datatypes first.", msg.Label), false, DIALOGGENERIC)
			return m, tea.Batch(
				DialogSetCmd(&dialog),
				DialogActiveSetCmd(true),
				FocusSetCmd(DIALOGFOCUS),
			)
		}
		dialog := NewDialog("Delete Datatype", fmt.Sprintf("Delete datatype '%s'?\nThis will remove all field associations.", msg.Label), true, DIALOGDELETEDATATYPE)
		dialog.SetButtons("Delete", "Cancel")
		deleteDatatypeContext = &DeleteDatatypeContext{
			DatatypeID: msg.DatatypeID,
			Label:      msg.Label,
		}
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteFieldDialogMsg:
		dialog := NewDialog("Delete Field", fmt.Sprintf("Delete field '%s'?\nThis will unlink it from the datatype and remove the field.", msg.Label), true, DIALOGDELETEFIELD)
		dialog.SetButtons("Delete", "Cancel")
		deleteFieldContext = &DeleteFieldContext{
			FieldID:    msg.FieldID,
			DatatypeID: msg.DatatypeID,
			Label:      msg.Label,
		}
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteRouteDialogMsg:
		dialog := NewDialog("Delete Route", fmt.Sprintf("Delete route '%s'?\nAssociated content will also be removed.", msg.Title), true, DIALOGDELETEROUTE)
		dialog.SetButtons("Delete", "Cancel")
		deleteRouteContext = &DeleteRouteContext{
			RouteID: msg.RouteID,
			Title:   msg.Title,
		}
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteMediaDialogMsg:
		dialog := NewDialog("Delete Media", fmt.Sprintf("Delete media '%s'?\nThis cannot be undone.", msg.Label), true, DIALOGDELETEMEDIA)
		dialog.SetButtons("Delete", "Cancel")
		deleteMediaContext = &DeleteMediaContext{
			MediaID: msg.MediaID,
			Label:   msg.Label,
		}
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteUserDialogMsg:
		dialog := NewDialog("Delete User", fmt.Sprintf("Delete user '%s'?\nThis cannot be undone.", msg.Username), true, DIALOGDELETEUSER)
		dialog.SetButtons("Delete", "Cancel")
		deleteUserContext = &DeleteUserContext{
			UserID:   msg.UserID,
			Username: msg.Username,
		}
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowUserFormDialogMsg:
		dialog := NewUserFormDialog(msg.Title, msg.Roles)
		return m, tea.Batch(
			UserFormDialogSetCmd(&dialog),
			UserFormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditUserDialogMsg:
		dialog := NewEditUserFormDialog("Edit User", msg.User, msg.Roles)
		return m, tea.Batch(
			UserFormDialogSetCmd(&dialog),
			UserFormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditSingleFieldDialogMsg:
		// Store context for when the dialog is accepted
		editSingleFieldContext = &editSingleFieldCtx{
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
			ContentFormDialogSetCmd(&dialog),
			ContentFormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowAddContentFieldDialogMsg:
		// Store context for when the dialog is accepted
		addContentFieldContext = &addContentFieldCtx{
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
			FormDialogSetCmd(&dialog),
			FormDialogActiveSetCmd(true),
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
		deleteContentFieldContext = &DeleteContentFieldContext{
			ContentFieldID: msg.Field.ContentFieldID,
			ContentID:      msg.ContentID,
			RouteID:        msg.RouteID,
			DatatypeID:     msg.DatatypeID,
		}
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowConfigFieldEditMsg:
		// Create a simple form dialog for editing a config field value
		labelInput := textinput.New()
		labelInput.Placeholder = "Value"
		labelInput.CharLimit = 512
		labelInput.Width = 50
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
			FormDialogSetCmd(&dialog),
			FormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowApproveAllRoutesDialogMsg:
		// Show confirmation dialog listing pending routes before approving
		routeList := strings.Join(msg.PendingRoutes, "\n  ")
		dialogMsg := fmt.Sprintf("Approve %d routes for plugin '%s'?\n\n  %s", len(msg.PendingRoutes), msg.PluginName, routeList)
		dialog := NewDialog("Approve Plugin Routes", dialogMsg, true, DIALOGAPPROVEPLUGINROUTES)
		dialog.SetButtons("Approve", "Cancel")
		approvePluginRoutesContext = &ApprovePluginContext{PluginName: msg.PluginName}
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowApproveAllHooksDialogMsg:
		// Show confirmation dialog listing pending hooks before approving
		hookList := strings.Join(msg.PendingHooks, "\n  ")
		dialogMsg := fmt.Sprintf("Approve %d hooks for plugin '%s'?\n\n  %s", len(msg.PendingHooks), msg.PluginName, hookList)
		dialog := NewDialog("Approve Plugin Hooks", dialogMsg, true, DIALOGAPPROVEPLUGINSHOOKS)
		dialog.SetButtons("Approve", "Cancel")
		approvePluginHooksContext = &ApprovePluginContext{PluginName: msg.PluginName}
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteAdminRouteDialogMsg:
		dialog := NewDialog("Delete Admin Route", fmt.Sprintf("Delete admin route '%s'?\nThis cannot be undone.", msg.Title), true, DIALOGDELETEADMINROUTE)
		dialog.SetButtons("Delete", "Cancel")
		deleteAdminRouteContext = &DeleteAdminRouteContext{
			AdminRouteID: msg.AdminRouteID,
			Title:        msg.Title,
		}
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteAdminDatatypeDialogMsg:
		if msg.HasChildren {
			dialog := NewDialog("Cannot Delete", fmt.Sprintf("Cannot delete '%s' because it has child datatypes.\nDelete child datatypes first.", msg.Label), false, DIALOGGENERIC)
			return m, tea.Batch(
				DialogSetCmd(&dialog),
				DialogActiveSetCmd(true),
				FocusSetCmd(DIALOGFOCUS),
			)
		}
		dialog := NewDialog("Delete Admin Datatype", fmt.Sprintf("Delete admin datatype '%s'?\nThis will remove all field associations.", msg.Label), true, DIALOGDELETEADMINDATATYPE)
		dialog.SetButtons("Delete", "Cancel")
		deleteAdminDatatypeContext = &DeleteAdminDatatypeContext{
			AdminDatatypeID: msg.AdminDatatypeID,
			Label:           msg.Label,
		}
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteAdminFieldDialogMsg:
		dialog := NewDialog("Delete Admin Field", fmt.Sprintf("Delete admin field '%s'?\nThis will unlink it from the datatype and remove the field.", msg.Label), true, DIALOGDELETEADMINFIELD)
		dialog.SetButtons("Delete", "Cancel")
		deleteAdminFieldContext = &DeleteAdminFieldContext{
			AdminFieldID:    msg.AdminFieldID,
			AdminDatatypeID: msg.AdminDatatypeID,
			Label:           msg.Label,
		}
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteFieldTypeDialogMsg:
		dialog := NewDialog("Delete Field Type", fmt.Sprintf("Delete field type '%s'?\nThis cannot be undone.", msg.Label), true, DIALOGDELETEFIELDTYPE)
		dialog.SetButtons("Delete", "Cancel")
		deleteFieldTypeContext = &DeleteFieldTypeContext{
			FieldTypeID: msg.FieldTypeID,
			Label:       msg.Label,
		}
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditFieldTypeDialogMsg:
		// Edit field type dialog with pre-populated values
		dialog := NewRouteFormDialog("Edit Field Type", FORMDIALOGEDITFIELDTYPE)
		dialog.LabelInput.SetValue(msg.FieldType.Type)
		dialog.TypeInput.SetValue(msg.FieldType.Label)
		dialog.EntityID = string(msg.FieldType.FieldTypeID)
		return m, tea.Batch(
			FormDialogSetCmd(&dialog),
			FormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowDeleteAdminFieldTypeDialogMsg:
		dialog := NewDialog("Delete Admin Field Type", fmt.Sprintf("Delete admin field type '%s'?\nThis cannot be undone.", msg.Label), true, DIALOGDELETEADMINFIELDTYPE)
		dialog.SetButtons("Delete", "Cancel")
		deleteAdminFieldTypeContext = &DeleteAdminFieldTypeContext{
			AdminFieldTypeID: msg.AdminFieldTypeID,
			Label:            msg.Label,
		}
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditAdminFieldTypeDialogMsg:
		// Edit admin field type dialog with pre-populated values
		dialog := NewRouteFormDialog("Edit Admin Field Type", FORMDIALOGEDITADMINFIELDTYPE)
		dialog.LabelInput.SetValue(msg.AdminFieldType.Type)
		dialog.TypeInput.SetValue(msg.AdminFieldType.Label)
		dialog.EntityID = string(msg.AdminFieldType.AdminFieldTypeID)
		return m, tea.Batch(
			FormDialogSetCmd(&dialog),
			FormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case UserFormDialogAcceptMsg:
		switch msg.Action {
		case FORMDIALOGCREATEUSER:
			return m, tea.Batch(
				UserFormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				CreateUserFromDialogCmd(msg.Username, msg.Name, msg.Email, msg.Password, msg.Role),
			)
		case FORMDIALOGEDITUSER:
			return m, tea.Batch(
				UserFormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				UpdateUserFromDialogCmd(msg.EntityID, msg.Username, msg.Name, msg.Email, msg.Role),
			)
		default:
			return m, tea.Batch(
				UserFormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		}
	case UserFormDialogCancelMsg:
		return m, tea.Batch(
			UserFormDialogActiveSetCmd(false),
			FocusSetCmd(PAGEFOCUS),
		)
	case UserCreatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("User created: %s", msg.Username)),
			UsersFetchCmd(),
		)
	case UserUpdatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("User updated: %s", msg.Username)),
			UsersFetchCmd(),
		)
	case UserDeletedMsg:
		newModel := m
		newModel.Cursor = 0
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("User deleted: %s", msg.UserID)),
			UsersFetchCmd(),
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
		case DIALOGDELETEDATATYPE:
			if deleteDatatypeContext != nil {
				datatypeID := deleteDatatypeContext.DatatypeID
				deleteDatatypeContext = nil
				return m, tea.Batch(
					DialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					DeleteDatatypeCmd(datatypeID),
				)
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case DIALOGDELETEFIELD:
			if deleteFieldContext != nil {
				fieldID := deleteFieldContext.FieldID
				datatypeID := deleteFieldContext.DatatypeID
				deleteFieldContext = nil
				return m, tea.Batch(
					DialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					DeleteFieldCmd(fieldID, datatypeID),
				)
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case DIALOGDELETEROUTE:
			if deleteRouteContext != nil {
				routeID := deleteRouteContext.RouteID
				deleteRouteContext = nil
				return m, tea.Batch(
					DialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					DeleteRouteCmd(routeID),
				)
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case DIALOGDELETEMEDIA:
			if deleteMediaContext != nil {
				mediaID := deleteMediaContext.MediaID
				deleteMediaContext = nil
				return m, tea.Batch(
					DialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					DeleteMediaCmd(mediaID),
				)
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case DIALOGDELETEUSER:
			if deleteUserContext != nil {
				userID := deleteUserContext.UserID
				deleteUserContext = nil
				return m, tea.Batch(
					DialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					DeleteUserCmd(userID),
				)
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case DIALOGDELETEADMINROUTE:
			if deleteAdminRouteContext != nil {
				adminRouteID := deleteAdminRouteContext.AdminRouteID
				deleteAdminRouteContext = nil
				return m, tea.Batch(
					DialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					DeleteAdminRouteCmd(adminRouteID),
				)
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case DIALOGDELETEADMINDATATYPE:
			if deleteAdminDatatypeContext != nil {
				adminDatatypeID := deleteAdminDatatypeContext.AdminDatatypeID
				deleteAdminDatatypeContext = nil
				return m, tea.Batch(
					DialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					DeleteAdminDatatypeCmd(adminDatatypeID),
				)
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case DIALOGDELETEADMINFIELD:
			if deleteAdminFieldContext != nil {
				ctx := deleteAdminFieldContext
				deleteAdminFieldContext = nil
				return m, tea.Batch(
					DialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					DeleteAdminFieldCmd(ctx.AdminFieldID, ctx.AdminDatatypeID),
				)
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case DIALOGDELETEFIELDTYPE:
			if deleteFieldTypeContext != nil {
				fieldTypeID := deleteFieldTypeContext.FieldTypeID
				deleteFieldTypeContext = nil
				return m, tea.Batch(
					DialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					DeleteFieldTypeCmd(fieldTypeID),
				)
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case DIALOGDELETEADMINFIELDTYPE:
			if deleteAdminFieldTypeContext != nil {
				adminFieldTypeID := deleteAdminFieldTypeContext.AdminFieldTypeID
				deleteAdminFieldTypeContext = nil
				return m, tea.Batch(
					DialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					DeleteAdminFieldTypeCmd(adminFieldTypeID),
				)
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case DIALOGBACKUPRESTORE:
			if restoreBackupContext != nil {
				backupPath := restoreBackupContext.Path
				restoreBackupContext = nil
				cfg := m.Config
				return m, tea.Batch(
					DialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					func() tea.Msg {
						if err := backup.RestoreFromBackup(*cfg, backupPath); err != nil {
							return ActionResultMsg{
								Title:   "Restore Failed",
								Message: fmt.Sprintf("Failed to restore backup:\n%s", err),
								IsError: true,
							}
						}
						return BackupRestoreCompleteMsg{Path: backupPath}
					},
				)
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case DIALOGAPPROVEPLUGINROUTES:
			if approvePluginRoutesContext != nil {
				pluginName := approvePluginRoutesContext.PluginName
				approvePluginRoutesContext = nil
				return m, tea.Batch(
					DialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					func() tea.Msg {
						return PluginApproveAllRoutesRequestMsg{Name: pluginName}
					},
				)
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case DIALOGAPPROVEPLUGINSHOOKS:
			if approvePluginHooksContext != nil {
				pluginName := approvePluginHooksContext.PluginName
				approvePluginHooksContext = nil
				return m, tea.Batch(
					DialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					func() tea.Msg {
						return PluginApproveAllHooksRequestMsg{Name: pluginName}
					},
				)
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case DIALOGDELETECONTENTFIELD:
			if deleteContentFieldContext != nil {
				ctx := deleteContentFieldContext
				deleteContentFieldContext = nil
				return m, tea.Batch(
					DialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					m.HandleDeleteContentField(ctx.ContentFieldID, ctx.ContentID, ctx.RouteID, ctx.DatatypeID),
				)
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case DIALOGQUICKSTART:
			schemaIndex := 0
			if m.Dialog != nil {
				schemaIndex = m.Dialog.ActionIndex
			}
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				RunQuickstartInstallCmd(m.Config, m.UserID, schemaIndex),
			)
		default:
			return m, tea.Batch(
				DialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		}
	case DialogCancelMsg:
		// If a restore just completed, quit the application
		if restoreRequiresQuit {
			restoreRequiresQuit = false
			return m, tea.Quit
		}
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
	case ShowAdminFormDialogMsg:
		// Admin form dialog with parent options from admin datatypes
		parentOpts := []ParentOption{
			{Label: "ROOT (no parent)", Value: ""},
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
			FormDialogSetCmd(&dialog),
			FormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditAdminRouteDialogMsg:
		// Edit admin route dialog with pre-populated values
		dialog := NewRouteFormDialog("Edit Admin Route", FORMDIALOGEDITADMINROUTE)
		dialog.LabelInput.SetValue(msg.Route.Title)
		dialog.TypeInput.SetValue(string(msg.Route.Slug))
		dialog.EntityID = string(msg.Route.AdminRouteID)
		return m, tea.Batch(
			FormDialogSetCmd(&dialog),
			FormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditAdminDatatypeDialogMsg:
		// Edit admin datatype dialog with pre-populated values
		parentOpts := []ParentOption{
			{Label: "ROOT (no parent)", Value: ""},
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
			FormDialogSetCmd(&dialog),
			FormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowEditAdminFieldDialogMsg:
		// Edit admin field dialog with pre-populated values
		dialog := NewFieldFormDialog("Edit Admin Field", FORMDIALOGEDITADMINFIELD)
		dialog.LabelInput.SetValue(msg.Field.Label)
		dialog.TypeIndex = FieldInputTypeIndex(string(msg.Field.Type))
		dialog.EntityID = string(msg.Field.AdminFieldID)
		return m, tea.Batch(
			FormDialogSetCmd(&dialog),
			FormDialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case ShowCreateRouteWithContentDialogMsg:
		// Create route with initial content dialog
		dialog := NewRouteWithContentDialog("New Content", FORMDIALOGCREATEROUTEWITHCONTENT, msg.RootDatatypes)
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
	case ShowMoveContentDialogMsg:
		sourceID := string(msg.SourceNode.Instance.ContentDataID)
		dialog := NewMoveContentDialog("Move Content", sourceID, string(msg.RouteID), msg.ValidTargets)
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
			// Create a new route with initial content (ParentID=DatatypeID from carousel, Label=Title, Type=Slug)
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				CreateRouteWithContentCmd(msg.Label, msg.Type, msg.ParentID),
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
		case FORMDIALOGMOVECONTENT:
			// ParentID = selected target content ID, EntityID = "sourceContentID|routeID"
			parts := strings.SplitN(msg.EntityID, "|", 2)
			if len(parts) == 2 && msg.ParentID != "" {
				return m, tea.Batch(
					FormDialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					MoveContentCmd(types.ContentID(parts[0]), types.ContentID(msg.ParentID), types.RouteID(parts[1])),
				)
			}
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case FORMDIALOGADDCONTENTFIELD:
			// ParentID = selected field ID from the picker
			if addContentFieldContext != nil && msg.ParentID != "" {
				ctx := addContentFieldContext
				addContentFieldContext = nil
				return m, tea.Batch(
					FormDialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					m.HandleAddContentField(ctx.ContentID, types.FieldID(msg.ParentID), ctx.RouteID, ctx.DatatypeID),
				)
			}
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case FORMDIALOGCREATEADMINROUTE:
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				CreateAdminRouteFromDialogCmd(msg.Label, msg.Type),
			)
		case FORMDIALOGEDITADMINROUTE:
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				UpdateAdminRouteFromDialogCmd(msg.EntityID, msg.Label, msg.Type, msg.ParentID),
			)
		case FORMDIALOGCREATEADMINDATATYPE:
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				CreateAdminDatatypeFromDialogCmd(msg.Label, msg.Type, msg.ParentID),
			)
		case FORMDIALOGEDITADMINDATATYPE:
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				UpdateAdminDatatypeFromDialogCmd(msg.EntityID, msg.Label, msg.Type, msg.ParentID),
			)
		case FORMDIALOGCREATEADMINFIELD:
			// Create a field and link it to the selected admin datatype
			if len(m.AdminAllDatatypes) > 0 && m.Cursor < len(m.AdminAllDatatypes) {
				dt := m.AdminAllDatatypes[m.Cursor]
				return m, tea.Batch(
					FormDialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					CreateAdminFieldFromDialogCmd(msg.Label, msg.Type, dt.AdminDatatypeID),
				)
			}
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		case FORMDIALOGEDITADMINFIELD:
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				UpdateAdminFieldFromDialogCmd(msg.EntityID, msg.Label, msg.Type),
			)
		case FORMDIALOGCREATEFIELDTYPE:
			// Create a new field type (Label=Type value, Type=Label value)
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				CreateFieldTypeFromDialogCmd(msg.Label, msg.Type),
			)
		case FORMDIALOGEDITFIELDTYPE:
			// Update an existing field type (Label=Type value, Type=Label value)
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				UpdateFieldTypeFromDialogCmd(msg.EntityID, msg.Label, msg.Type),
			)
		case FORMDIALOGCREATEADMINFIELDTYPE:
			// Create a new admin field type (Label=Type value, Type=Label value)
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				CreateAdminFieldTypeFromDialogCmd(msg.Label, msg.Type),
			)
		case FORMDIALOGEDITADMINFIELDTYPE:
			// Update an existing admin field type (Label=Type value, Type=Label value)
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				UpdateAdminFieldTypeFromDialogCmd(msg.EntityID, msg.Label, msg.Type),
			)
		case FORMDIALOGCONFIGEDIT:
			// EntityID holds the JSON key, Label holds the new value
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
				func() tea.Msg {
					return ConfigFieldUpdateMsg{
						Key:   msg.EntityID,
						Value: msg.Label,
					}
				},
			)
		default:
			return m, tea.Batch(
				FormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
			)
		}
	case FormDialogCancelMsg:
		addContentFieldContext = nil
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
		dialog.Logger = logger
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
		case FORMDIALOGEDIITSINGLEFIELD:
			// Single-field edit: use stored context for ContentFieldID
			if editSingleFieldContext != nil {
				ctx := editSingleFieldContext
				editSingleFieldContext = nil
				var newValue string
				for _, val := range msg.FieldValues {
					newValue = val
					break
				}
				return m, tea.Batch(
					ContentFormDialogActiveSetCmd(false),
					FocusSetCmd(PAGEFOCUS),
					m.HandleEditSingleField(
						ctx.ContentFieldID,
						ctx.ContentID, ctx.FieldID, newValue, ctx.RouteID,
						ctx.DatatypeID,
					),
				)
			}
			return m, tea.Batch(
				ContentFormDialogActiveSetCmd(false),
				FocusSetCmd(PAGEFOCUS),
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
		editSingleFieldContext = nil
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
		dialog.Logger = logger
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
	case DatatypeDeletedMsg:
		// Refresh datatypes list after deletion
		newModel := m
		newModel.Cursor = 0
		newModel.FieldCursor = 0
		newModel.SelectedDatatypeFields = nil
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Datatype deleted: %s", msg.DatatypeID)),
			AllDatatypesFetchCmd(),
		)
	case FieldDeletedMsg:
		// Refresh fields list after deletion
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Field deleted: %s", msg.FieldID)),
			DatatypeFieldsFetchCmd(msg.DatatypeID),
		)
	case RouteDeletedMsg:
		// Refresh routes list after deletion
		newModel := m
		newModel.Cursor = 0
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Route deleted: %s", msg.RouteID)),
			RoutesFetchCmd(),
		)
	case MediaDeletedMsg:
		// Refresh media list after deletion
		newModel := m
		newModel.Cursor = 0
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Media deleted: %s", msg.MediaID)),
			MediaFetchCmd(),
		)

	// =========================================================================
	// DATABASE FORM DIALOG
	// =========================================================================
	// =========================================================================
	// UICONFIG FORM DIALOG
	// =========================================================================
	case ShowUIConfigFormDialogMsg:
		dialog := NewUIConfigFormDialog(msg.Title, msg.FieldID)
		m.UIConfigFormDialog = &dialog
		m.UIConfigFormDialogActive = true
		return m, FocusSetCmd(DIALOGFOCUS)
	case ShowEditUIConfigFormDialogMsg:
		dialog := NewEditUIConfigFormDialog(msg.Title, msg.FieldID, msg.Existing)
		m.UIConfigFormDialog = &dialog
		m.UIConfigFormDialogActive = true
		return m, FocusSetCmd(DIALOGFOCUS)
	case UIConfigFormDialogCancelMsg:
		m.UIConfigFormDialogActive = false
		m.UIConfigFormDialog = nil
		return m, FocusSetCmd(PAGEFOCUS)
	case UIConfigFormDialogAcceptMsg:
		m.UIConfigFormDialogActive = false
		m.UIConfigFormDialog = nil
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
			m.DatabaseFormDialog = &dialog
			m.DatabaseFormDialogActive = true
			return m, tea.Batch(
				FocusSetCmd(DIALOGFOCUS),
				StatusSetCmd(EDITING),
			)
		case FORMDIALOGDBUPDATE:
			// Get current row data
			if len(m.TableState.Rows) == 0 {
				return m, LogMessageCmd("No rows available for update")
			}
			recordIndex := (m.PageMod * m.MaxRows) + m.Cursor
			if recordIndex >= len(m.TableState.Rows) {
				return m, LogMessageCmd("Row index out of range")
			}
			currentRow := m.TableState.Rows[recordIndex]
			dialog := NewDatabaseUpdateDialog(msg.Title, msg.Table, columns, nil, currentRow)
			m.DatabaseFormDialog = &dialog
			m.DatabaseFormDialogActive = true
			return m, tea.Batch(
				FocusSetCmd(DIALOGFOCUS),
				StatusSetCmd(EDITING),
			)
		default:
			return m, LogMessageCmd(fmt.Sprintf("Unknown database form action: %s", msg.Action))
		}
	case DatabaseFormDialogAcceptMsg:
		m.DatabaseFormDialogActive = false
		m.DatabaseFormDialog = nil

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
		m.DatabaseFormDialogActive = false
		m.DatabaseFormDialog = nil
		return m, tea.Batch(
			FocusSetCmd(PAGEFOCUS),
			StatusSetCmd(OK),
		)

	case ConfigFieldUpdateMsg:
		// Apply config field update via the config manager
		if m.ConfigManager == nil {
			dialog := NewDialog("Error", "Config manager is not available.", false, DIALOGGENERIC)
			return m, tea.Batch(
				DialogSetCmd(&dialog),
				DialogActiveSetCmd(true),
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
				DialogSetCmd(&dialog),
				DialogActiveSetCmd(true),
				FocusSetCmd(DIALOGFOCUS),
			)
		}
		if len(msg.RestartRequired) > 0 {
			fields := strings.Join(msg.RestartRequired, ", ")
			dialog := NewDialog("Config Updated", fmt.Sprintf("Config saved. The following fields require a restart to take effect:\n\n%s", fields), false, DIALOGGENERIC)
			return m, tea.Batch(
				DialogSetCmd(&dialog),
				DialogActiveSetCmd(true),
				FocusSetCmd(DIALOGFOCUS),
			)
		}
		dialog := NewDialog("Config Updated", "Config saved successfully.", false, DIALOGGENERIC)
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	default:
		return m, nil
	}
}

// DatatypeCreatedFromDialogMsg is sent after a datatype is successfully created from dialog.
type DatatypeCreatedFromDialogMsg struct {
	DatatypeID types.DatatypeID
	Label      string
}

// CreateDatatypeFromDialogCmd creates a command to create a datatype from dialog input.
func CreateDatatypeFromDialogCmd(label, dtype, parentID string) tea.Cmd {
	return func() tea.Msg {
		return CreateDatatypeFromDialogRequestMsg{
			Label:    label,
			Type:     dtype,
			ParentID: parentID,
		}
	}
}

// CreateDatatypeFromDialogRequestMsg triggers datatype creation from dialog.
type CreateDatatypeFromDialogRequestMsg struct {
	Label    string
	Type     string
	ParentID string
}

// HandleCreateDatatypeFromDialog processes the datatype creation request.
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
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		// Prepare the type - default to ROOT if empty
		dtype := msg.Type
		if dtype == "" {
			dtype = "ROOT"
		}

		// Prepare parent ID
		var parentID types.NullableDatatypeID
		if msg.ParentID != "" {
			parentID = types.NullableDatatypeID{
				ID:    types.DatatypeID(msg.ParentID),
				Valid: true,
			}
		}

		// Create the datatype
		params := db.CreateDatatypeParams{
			DatatypeID:   types.NewDatatypeID(),
			ParentID:     parentID,
			Label:        msg.Label,
			Type:         dtype,
			AuthorID:     authorID,
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		dt, err := d.CreateDatatype(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to create datatype: %v", err),
			}
		}
		return DatatypeCreatedFromDialogMsg{
			DatatypeID: dt.DatatypeID,
			Label:      dt.Label,
		}
	}
}

// FieldCreatedFromDialogMsg is sent after a field is successfully created from dialog.
type FieldCreatedFromDialogMsg struct {
	FieldID    types.FieldID
	DatatypeID types.DatatypeID
	Label      string
}

// CreateFieldFromDialogCmd creates a command to create a field and link it to a datatype.
func CreateFieldFromDialogCmd(label, fieldType string, datatypeID types.DatatypeID) tea.Cmd {
	return func() tea.Msg {
		return CreateFieldFromDialogRequestMsg{
			Label:      label,
			Type:       fieldType,
			DatatypeID: datatypeID,
		}
	}
}

// CreateFieldFromDialogRequestMsg triggers field creation from dialog.
type CreateFieldFromDialogRequestMsg struct {
	Label      string
	Type       string
	DatatypeID types.DatatypeID
}

// HandleCreateFieldFromDialog processes the field creation request.
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
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		// Verify the author user exists in the database before creating the field
		authorUser, userErr := d.GetUser(authorID)
		if userErr != nil || authorUser == nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Cannot create field: author user %s not found in database (run --install to bootstrap)", authorID),
			}
		}

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
			Validation:   types.EmptyJSON,
			UIConfig:     types.EmptyJSON,
			Type:         fieldType,
			AuthorID:     nullableAuthorID,
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		field, err := d.CreateField(ctx, ac, fieldParams)
		if err != nil || field.FieldID.IsZero() {
			errMsg := "Failed to create field in database"
			if err != nil {
				errMsg = fmt.Sprintf("Failed to create field: %v", err)
			}
			return ActionResultMsg{
				Title:   "Error",
				Message: errMsg,
			}
		}

		// Link field to datatype via datatypes_fields join table
		dtFieldID := string(types.NewDatatypeFieldID())
		maxSort, sortErr := d.GetMaxSortOrderByDatatypeID(msg.DatatypeID)
		if sortErr != nil {
			maxSort = -1
		}
		dtFieldParams := db.CreateDatatypeFieldParams{
			ID:         dtFieldID,
			DatatypeID: msg.DatatypeID,
			FieldID:    field.FieldID,
			SortOrder:  maxSort + 1,
		}

		_, dtfErr := d.CreateDatatypeField(ctx, ac, dtFieldParams)
		if dtfErr != nil {
			return ActionResultMsg{
				Title:   "Warning",
				Message: fmt.Sprintf("Field created but failed to link to datatype: %v", dtfErr),
			}
		}

		return FieldCreatedFromDialogMsg{
			FieldID:    field.FieldID,
			DatatypeID: msg.DatatypeID,
			Label:      field.Label,
		}
	}
}

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

// HandleCreateRouteFromDialog processes the route creation request.
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
			Slug:   validSlug,
			Title:  msg.Title,
			Status: 1, // Active by default
			AuthorID: types.NullableUserID{
				ID:    authorID,
				Valid: true,
			},
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		route, err := d.CreateRoute(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to create route: %v", err),
			}
		}
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

// DatatypeUpdatedFromDialogMsg is sent after a datatype is successfully updated from dialog.
type DatatypeUpdatedFromDialogMsg struct {
	DatatypeID types.DatatypeID
	Label      string
}

// UpdateDatatypeFromDialogRequestMsg triggers datatype update from dialog.
type UpdateDatatypeFromDialogRequestMsg struct {
	DatatypeID string
	Label      string
	Type       string
	ParentID   string
}

// UpdateDatatypeFromDialogCmd creates a command to update a datatype from dialog input.
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

// HandleUpdateDatatypeFromDialog processes the datatype update request.
func (m Model) HandleUpdateDatatypeFromDialog(msg UpdateDatatypeFromDialogRequestMsg) tea.Cmd {
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

	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		// Fetch existing datatype to preserve unchanged values
		datatypeID := types.DatatypeID(msg.DatatypeID)
		existing, err := d.GetDatatype(datatypeID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to get datatype for update: %v", err),
			}
		}

		// Prepare the type - default to existing type if empty
		dtype := msg.Type
		if dtype == "" {
			dtype = existing.Type
		}

		// Prepare parent ID - use provided value or preserve existing
		parentID := existing.ParentID
		if msg.ParentID != "" {
			parentID = types.NullableDatatypeID{
				ID:    types.DatatypeID(msg.ParentID),
				Valid: true,
			}
		}

		// Update only changed fields; preserve author and date_created
		params := db.UpdateDatatypeParams{
			DatatypeID:   datatypeID,
			ParentID:     parentID,
			Label:        msg.Label,
			Type:         dtype,
			AuthorID:     existing.AuthorID,
			DateCreated:  existing.DateCreated,
			DateModified: types.TimestampNow(),
		}

		_, err = d.UpdateDatatype(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update datatype: %v", err),
			}
		}

		return DatatypeUpdatedFromDialogMsg{
			DatatypeID: datatypeID,
			Label:      msg.Label,
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
	Label   string
	Type    string
}

// UpdateFieldFromDialogCmd creates a command to update a field from dialog input.
func UpdateFieldFromDialogCmd(fieldID, label, fieldType string) tea.Cmd {
	return func() tea.Msg {
		return UpdateFieldFromDialogRequestMsg{
			FieldID: fieldID,
			Label:   label,
			Type:    fieldType,
		}
	}
}

// HandleUpdateFieldFromDialog processes the field update request.
func (m Model) HandleUpdateFieldFromDialog(msg UpdateFieldFromDialogRequestMsg) tea.Cmd {
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

	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		// Fetch existing field to preserve unchanged values
		fieldID := types.FieldID(msg.FieldID)
		existing, err := d.GetField(fieldID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to get field for update: %v", err),
			}
		}

		// Prepare the field type - default to existing type if empty
		fieldTypeStr := msg.Type
		if fieldTypeStr == "" {
			fieldTypeStr = string(existing.Type)
		}
		fieldType := types.FieldType(fieldTypeStr)

		// Update only label, type, and date_modified; preserve everything else
		params := db.UpdateFieldParams{
			FieldID:      fieldID,
			ParentID:     existing.ParentID,
			Label:        msg.Label,
			Data:         existing.Data,
			Validation:   existing.Validation,
			UIConfig:     existing.UIConfig,
			Type:         fieldType,
			AuthorID:     existing.AuthorID,
			DateCreated:  existing.DateCreated,
			DateModified: types.TimestampNow(),
		}

		_, err = d.UpdateField(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update field: %v", err),
			}
		}

		return FieldUpdatedFromDialogMsg{
			FieldID:    fieldID,
			DatatypeID: datatypeID,
			Label:      msg.Label,
		}
	}
}

// =============================================================================
// UPDATE FIELD UICONFIG
// =============================================================================

// HandleUpdateFieldUIConfig processes a field UIConfig update request.
func (m Model) HandleUpdateFieldUIConfig(msg UpdateFieldUIConfigRequestMsg) tea.Cmd {
	cfg := m.Config
	var datatypeID types.DatatypeID
	if len(m.AllDatatypes) > 0 && m.Cursor < len(m.AllDatatypes) {
		datatypeID = m.AllDatatypes[m.Cursor].DatatypeID
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot update field UIConfig: configuration not loaded",
			}
		}
	}

	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		fieldID := types.FieldID(msg.FieldID)
		existing, err := d.GetField(fieldID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to get field for UIConfig update: %v", err),
			}
		}

		params := db.UpdateFieldParams{
			FieldID:      fieldID,
			ParentID:     existing.ParentID,
			Label:        existing.Label,
			Data:         existing.Data,
			Validation:   existing.Validation,
			UIConfig:     msg.UIConfigJSON,
			Type:         existing.Type,
			AuthorID:     existing.AuthorID,
			DateCreated:  existing.DateCreated,
			DateModified: types.TimestampNow(),
		}

		_, err = d.UpdateField(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update field UIConfig: %v", err),
			}
		}

		return FieldUIConfigUpdatedMsg{
			FieldID:    fieldID,
			DatatypeID: datatypeID,
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

// HandleUpdateRouteFromDialog processes the route update request.
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

		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		_, err = d.UpdateRoute(ctx, ac, params)
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

// HandleCreateRouteWithContent processes the route with content creation request.
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
			Slug:   validSlug,
			Title:  msg.Title,
			Status: 1, // Active by default
			AuthorID: types.NullableUserID{
				ID:    authorID,
				Valid: true,
			},
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		route, routeErr := d.CreateRoute(ctx, ac, routeParams)
		if routeErr != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to create route: %v", routeErr),
			}
		}
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
			AuthorID:     authorID,
			Status:       types.ContentStatusDraft,
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		contentData, contentErr := d.CreateContentData(ctx, ac, contentParams)
		if contentErr != nil || contentData.ContentDataID.IsZero() {
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

// InitializeRouteContentContext stores context for route content initialization.
type InitializeRouteContentContext struct {
	Route      db.Routes
	DatatypeID string
}

// Global variable to store the context for route content initialization.
var initializeRouteContentContext *InitializeRouteContentContext

// =============================================================================
// DELETE CONTENT
// =============================================================================

// DeleteContentContext stores context for a content deletion operation.
type DeleteContentContext struct {
	ContentID string
	RouteID   string
}

// RestoreBackupContext stores context for a backup restore operation.
type RestoreBackupContext struct {
	Path string
}

var restoreBackupContext *RestoreBackupContext
var restoreRequiresQuit bool

// Global variable to store delete context for content deletion.
var deleteContentContext *DeleteContentContext

// =============================================================================
// APPROVE PLUGIN ROUTES / HOOKS
// =============================================================================

// ApprovePluginContext stores context for a plugin approval confirmation dialog.
type ApprovePluginContext struct {
	PluginName string
}

// Global variables to store context for plugin approval dialogs.
var approvePluginRoutesContext *ApprovePluginContext
var approvePluginHooksContext *ApprovePluginContext

// Global variable to store delete content field context.
var deleteContentFieldContext *DeleteContentFieldContext

// editSingleFieldCtx stores context for the single-field edit dialog.
type editSingleFieldCtx struct {
	ContentFieldID types.ContentFieldID
	ContentID      types.ContentID
	FieldID        types.FieldID
	RouteID        types.RouteID
	DatatypeID     types.NullableDatatypeID
}

var editSingleFieldContext *editSingleFieldCtx

// addContentFieldCtx stores context for the add content field operation.
type addContentFieldCtx struct {
	ContentID  types.ContentID
	RouteID    types.RouteID
	DatatypeID types.NullableDatatypeID
}

var addContentFieldContext *addContentFieldCtx

// =============================================================================
// DELETE DATATYPE
// =============================================================================

// DeleteDatatypeContext stores context for a datatype deletion operation.
type DeleteDatatypeContext struct {
	DatatypeID types.DatatypeID
	Label      string
}

var deleteDatatypeContext *DeleteDatatypeContext

// ShowDeleteDatatypeDialogMsg triggers showing a delete datatype confirmation dialog.
type ShowDeleteDatatypeDialogMsg struct {
	DatatypeID  types.DatatypeID
	Label       string
	HasChildren bool
}

// ShowDeleteDatatypeDialogCmd creates a command to show a delete datatype confirmation dialog.
func ShowDeleteDatatypeDialogCmd(datatypeID types.DatatypeID, label string, hasChildren bool) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteDatatypeDialogMsg{
			DatatypeID:  datatypeID,
			Label:       label,
			HasChildren: hasChildren,
		}
	}
}

// DeleteDatatypeRequestMsg triggers datatype deletion.
type DeleteDatatypeRequestMsg struct {
	DatatypeID types.DatatypeID
}

// DatatypeDeletedMsg is sent after a datatype is successfully deleted.
type DatatypeDeletedMsg struct {
	DatatypeID types.DatatypeID
}

// DeleteDatatypeCmd creates a command to delete a datatype.
func DeleteDatatypeCmd(datatypeID types.DatatypeID) tea.Cmd {
	return func() tea.Msg {
		return DeleteDatatypeRequestMsg{DatatypeID: datatypeID}
	}
}

// =============================================================================
// DELETE FIELD
// =============================================================================

// DeleteFieldContext stores context for a field deletion operation.
type DeleteFieldContext struct {
	FieldID    types.FieldID
	DatatypeID types.DatatypeID
	Label      string
}

var deleteFieldContext *DeleteFieldContext

// ShowDeleteFieldDialogMsg triggers showing a delete field confirmation dialog.
type ShowDeleteFieldDialogMsg struct {
	FieldID    types.FieldID
	DatatypeID types.DatatypeID
	Label      string
}

// ShowDeleteFieldDialogCmd creates a command to show a delete field confirmation dialog.
func ShowDeleteFieldDialogCmd(fieldID types.FieldID, datatypeID types.DatatypeID, label string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteFieldDialogMsg{
			FieldID:    fieldID,
			DatatypeID: datatypeID,
			Label:      label,
		}
	}
}

// DeleteFieldRequestMsg triggers field deletion.
type DeleteFieldRequestMsg struct {
	FieldID    types.FieldID
	DatatypeID types.DatatypeID
}

// FieldDeletedMsg is sent after a field is successfully deleted.
type FieldDeletedMsg struct {
	FieldID    types.FieldID
	DatatypeID types.DatatypeID
}

// DeleteFieldCmd creates a command to delete a field.
func DeleteFieldCmd(fieldID types.FieldID, datatypeID types.DatatypeID) tea.Cmd {
	return func() tea.Msg {
		return DeleteFieldRequestMsg{FieldID: fieldID, DatatypeID: datatypeID}
	}
}

// =============================================================================
// DELETE ROUTE
// =============================================================================

// DeleteRouteContext stores context for a route deletion operation.
type DeleteRouteContext struct {
	RouteID types.RouteID
	Title   string
}

var deleteRouteContext *DeleteRouteContext

// ShowDeleteRouteDialogMsg triggers showing a delete route confirmation dialog.
type ShowDeleteRouteDialogMsg struct {
	RouteID types.RouteID
	Title   string
}

// ShowDeleteRouteDialogCmd creates a command to show a delete route confirmation dialog.
func ShowDeleteRouteDialogCmd(routeID types.RouteID, title string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteRouteDialogMsg{
			RouteID: routeID,
			Title:   title,
		}
	}
}

// DeleteRouteRequestMsg triggers route deletion.
type DeleteRouteRequestMsg struct {
	RouteID types.RouteID
}

// RouteDeletedMsg is sent after a route is successfully deleted.
type RouteDeletedMsg struct {
	RouteID types.RouteID
}

// DeleteRouteCmd creates a command to delete a route.
func DeleteRouteCmd(routeID types.RouteID) tea.Cmd {
	return func() tea.Msg {
		return DeleteRouteRequestMsg{RouteID: routeID}
	}
}

// =============================================================================
// DELETE MEDIA
// =============================================================================

// DeleteMediaContext stores context for a media deletion operation.
type DeleteMediaContext struct {
	MediaID types.MediaID
	Label   string
}

var deleteMediaContext *DeleteMediaContext

// ShowDeleteMediaDialogMsg triggers showing a delete media confirmation dialog.
type ShowDeleteMediaDialogMsg struct {
	MediaID types.MediaID
	Label   string
}

// ShowDeleteMediaDialogCmd creates a command to show a delete media confirmation dialog.
func ShowDeleteMediaDialogCmd(mediaID types.MediaID, label string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteMediaDialogMsg{
			MediaID: mediaID,
			Label:   label,
		}
	}
}

// DeleteMediaRequestMsg triggers media deletion.
type DeleteMediaRequestMsg struct {
	MediaID types.MediaID
}

// MediaDeletedMsg is sent after a media item is successfully deleted.
type MediaDeletedMsg struct {
	MediaID types.MediaID
}

// DeleteMediaCmd creates a command to delete a media item.
func DeleteMediaCmd(mediaID types.MediaID) tea.Cmd {
	return func() tea.Msg {
		return DeleteMediaRequestMsg{MediaID: mediaID}
	}
}

// =============================================================================
// DELETE USER
// =============================================================================

// DeleteUserContext stores context for a user deletion operation.
type DeleteUserContext struct {
	UserID   types.UserID
	Username string
}

var deleteUserContext *DeleteUserContext

// ShowDeleteUserDialogMsg triggers showing a delete user confirmation dialog.
type ShowDeleteUserDialogMsg struct {
	UserID   types.UserID
	Username string
}

// ShowDeleteUserDialogCmd creates a command to show a delete user confirmation dialog.
func ShowDeleteUserDialogCmd(userID types.UserID, username string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteUserDialogMsg{
			UserID:   userID,
			Username: username,
		}
	}
}

// DeleteUserRequestMsg triggers user deletion.
type DeleteUserRequestMsg struct {
	UserID types.UserID
}

// UserDeletedMsg is sent after a user is successfully deleted.
type UserDeletedMsg struct {
	UserID types.UserID
}

// DeleteUserCmd creates a command to delete a user.
func DeleteUserCmd(userID types.UserID) tea.Cmd {
	return func() tea.Msg {
		return DeleteUserRequestMsg{UserID: userID}
	}
}

// =============================================================================
// CREATE/UPDATE USER
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

// InitializeRouteContentContextCmd creates a command to store the context for route content initialization.
func InitializeRouteContentContextCmd(route db.Routes, datatypeID string) tea.Cmd {
	return func() tea.Msg {
		initializeRouteContentContext = &InitializeRouteContentContext{
			Route:      route,
			DatatypeID: datatypeID,
		}
		return nil
	}
}

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

// HandleInitializeRouteContent processes the route content initialization request.
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
			AuthorID:     authorID,
			Status:       types.ContentStatusDraft,
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		contentData, contentErr := d.CreateContentData(ctx, ac, contentParams)
		if contentErr != nil || contentData.ContentDataID.IsZero() {
			errMsg := "Failed to create content in database"
			if contentErr != nil {
				errMsg = fmt.Sprintf("Failed to create content: %v", contentErr)
			}
			return ActionResultMsg{
				Title:   "Error",
				Message: errMsg,
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

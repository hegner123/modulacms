package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// handleAdminDialogMsg handles admin-specific dialog show/setup messages.
// Returns (model, cmd, handled). If handled is false, the caller should
// continue dispatching through the main switch.
func (m Model) handleAdminDialogMsg(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case ShowDeleteAdminContentDialogMsg:
		if msg.HasChildren {
			dialog := NewDialog("Cannot Delete", fmt.Sprintf("Cannot delete '%s' because it has children.\nDelete child content first.", msg.ContentName), false, DIALOGGENERIC)
			return m, tea.Batch(
				OverlaySetCmd(&dialog),
				FocusSetCmd(DIALOGFOCUS),
			), true
		}
		dialog := NewDialog("Delete Admin Content", fmt.Sprintf("Delete '%s'?\nAll field values will be deleted.", msg.ContentName), true, DIALOGDELETECONTENT)
		dialog.SetButtons("Delete", "Cancel")
		m.DCtx.Active = &DeleteContentContext{
			ContentID: string(msg.AdminContentID),
			RouteID:   string(msg.AdminRouteID),
			AdminMode: true,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		), true
	case ShowPublishAdminContentDialogMsg:
		var dialogMsg string
		var dialogAction DialogAction
		var title string
		if msg.IsPublished {
			title = "Unpublish Admin Content"
			dialogMsg = fmt.Sprintf("Unpublish '%s'?\nIt will no longer be publicly accessible.", msg.Name)
			dialogAction = DIALOGUNPUBLISHCONTENT
		} else {
			title = "Publish Admin Content"
			dialogMsg = fmt.Sprintf("Publish '%s'?\nThis creates a public snapshot.", msg.Name)
			dialogAction = DIALOGPUBLISHCONTENT
		}
		dialog := NewDialog(title, dialogMsg, true, dialogAction)
		if msg.IsPublished {
			dialog.SetButtons("Unpublish", "Cancel")
		} else {
			dialog.SetButtons("Publish", "Cancel")
		}
		m.DCtx.Active = &PublishContentContext{
			ContentID: types.ContentID(msg.AdminContentID),
			RouteID:   types.RouteID(msg.AdminRouteID),
			AdminMode: true,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		), true
	case ShowRestoreVersionDialogMsg:
		dialog := NewDialog("Restore Version", fmt.Sprintf("Restore version %d?\nCurrent field values will be overwritten.", msg.VersionNumber), true, DIALOGRESTOREVERSION)
		dialog.SetButtons("Restore", "Cancel")
		m.DCtx.Active = &RestoreVersionContext{
			ContentID: msg.ContentID,
			VersionID: msg.VersionID,
			RouteID:   msg.RouteID,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		), true
	case ShowRestoreAdminVersionDialogMsg:
		dialog := NewDialog("Restore Version", fmt.Sprintf("Restore version %d?\nCurrent field values will be overwritten.", msg.VersionNumber), true, DIALOGRESTOREVERSION)
		dialog.SetButtons("Restore", "Cancel")
		m.DCtx.Active = &RestoreVersionContext{
			ContentID: types.ContentID(msg.AdminContentID),
			VersionID: types.ContentVersionID(msg.VersionID),
			RouteID:   types.RouteID(msg.AdminRouteID),
			AdminMode: true,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		), true
	case ShowDeleteAdminContentFieldDialogMsg:
		dialog := NewDialog("Delete Field Value", fmt.Sprintf("Delete value for field '%s'?\nThis removes the stored value.", msg.Label), true, DIALOGDELETECONTENTFIELD)
		dialog.SetButtons("Delete", "Cancel")
		m.DCtx.Active = &DeleteContentFieldContext{
			ContentFieldID: types.ContentFieldID(msg.AdminContentFieldID),
			ContentID:      types.ContentID(msg.AdminContentID),
			RouteID:        types.RouteID(msg.AdminRouteID),
			DatatypeID:     types.NullableDatatypeID{ID: types.DatatypeID(msg.AdminDatatypeID.ID), Valid: msg.AdminDatatypeID.Valid},
			AdminMode:      true,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		), true
	case ShowMoveAdminContentDialogMsg:
		if len(msg.Targets) == 0 {
			dialog := NewDialog("Cannot Move", "No valid move targets available.", false, DIALOGGENERIC)
			return m, tea.Batch(
				OverlaySetCmd(&dialog),
				FocusSetCmd(DIALOGFOCUS),
			), true
		}
		m.DCtx.Active = &MoveAdminContentContext{
			SourceNode:   msg.SourceNode,
			AdminRouteID: msg.AdminRouteID,
		}
		dialog := FormDialogModel{
			dialogStyles:  newDialogStyles(),
			Title:         "Move Content",
			Width:         50,
			Action:        FORMDIALOGMOVEADMINCONTENT,
			LabelInput:    textinput.New(),
			TypeInput:     textinput.New(),
			ParentOptions: msg.Targets,
			ParentIndex:   0,
			focusIndex:    FormDialogFieldParent,
			EntityID:      string(msg.SourceNode.Instance.ContentDataID) + "|" + string(msg.AdminRouteID),
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		), true

	// =========================================================================
	// ADMIN SINGLE-FIELD AND ADD-FIELD DIALOGS
	// =========================================================================
	case ShowEditAdminSingleFieldDialogMsg:
		m.DCtx.Active = &editAdminSingleFieldCtx{
			AdminContentFieldID: msg.Field.AdminContentFieldID,
			AdminContentID:      msg.AdminContentID,
			AdminFieldID:        msg.Field.AdminFieldID,
			AdminRouteID:        msg.AdminRouteID,
			AdminDatatypeID:     msg.AdminDatatypeID,
			Label:               msg.Field.Label,
			Type:                msg.Field.Type,
			Value:               msg.Field.Value,
		}
		// Build a single-field edit form using the regular ContentFormDialog
		// with admin-mapped types and FORMDIALOGEDITADMINSINGLEFIELD action.
		existingFields := []ExistingContentField{{
			ContentFieldID: types.ContentFieldID(msg.Field.AdminContentFieldID),
			FieldID:        types.FieldID(msg.Field.AdminFieldID),
			Label:          msg.Field.Label,
			Type:           msg.Field.Type,
			Value:          msg.Field.Value,
			ValidationJSON: msg.Field.ValidationJSON,
			DataJSON:       msg.Field.DataJSON,
		}}
		dialog := NewEditContentFormDialog(
			fmt.Sprintf("Edit: %s", msg.Field.Label),
			types.ContentID(msg.AdminContentID),
			types.DatatypeID(msg.AdminDatatypeID.ID),
			types.RouteID(msg.AdminRouteID),
			existingFields,
		)
		dialog.Action = FORMDIALOGEDITADMINSINGLEFIELD
		dialog.Logger = m.Logger
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		), true
	case ShowAddAdminContentFieldDialogMsg:
		m.DCtx.Active = &addAdminContentFieldCtx{
			AdminContentID:  msg.AdminContentID,
			AdminRouteID:    msg.AdminRouteID,
			AdminDatatypeID: msg.AdminDatatypeID,
		}
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
			Action:        FORMDIALOGADDADMINCONTENTFIELD,
			LabelInput:    textinput.New(),
			TypeInput:     textinput.New(),
			ParentOptions: parents,
			ParentIndex:   0,
			focusIndex:    FormDialogFieldParent,
		}
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		), true

	// =========================================================================
	// ADMIN CONTENT FORM BUILD/EDIT DIALOGS
	// =========================================================================
	case AdminBuildContentFormMsg:
		// Build creation form using regular ContentFormDialog with admin action.
		var dbFields []db.Fields
		for _, f := range msg.Fields {
			dbFields = append(dbFields, db.Fields{
				FieldID: types.FieldID(f.AdminFieldID),
				Label:   f.Label,
				Type:    f.Type,
				Data:    f.Data,
			})
		}
		dialog := NewContentFormDialog(
			"New Admin Content",
			FORMDIALOGCREATEADMINCONTENT,
			types.DatatypeID(msg.AdminDatatypeID),
			types.RouteID(msg.AdminRouteID),
			dbFields,
		)
		dialog.Logger = m.Logger
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		), true
	case ShowEditAdminContentFormDialogMsg:
		// Build edit form using regular ContentFormDialog with admin action.
		var existingFields []ExistingContentField
		for _, f := range msg.Fields {
			existingFields = append(existingFields, ExistingContentField{
				ContentFieldID: types.ContentFieldID(f.AdminContentFieldID),
				FieldID:        types.FieldID(f.AdminFieldID),
				Label:          f.Label,
				Type:           f.Type,
				Value:          f.Value,
				ValidationJSON: f.ValidationJSON,
				DataJSON:       f.DataJSON,
			})
		}
		dialog := NewEditContentFormDialog(
			"Edit Admin Content",
			types.ContentID(msg.AdminContentID),
			types.DatatypeID(msg.AdminDatatypeID),
			types.RouteID(msg.AdminRouteID),
			existingFields,
		)
		dialog.Action = FORMDIALOGEDITADMINCONTENT
		dialog.Logger = m.Logger
		return m, tea.Batch(
			OverlaySetCmd(&dialog),
			FocusSetCmd(DIALOGFOCUS),
		), true

	// =========================================================================
	// ADMIN CONTENT FORM DIALOG HANDLING
	// =========================================================================
	case AdminContentFormDialogAcceptMsg:
		newM, cmd := m.HandleAdminContentFormDialogAccept(msg)
		return newM, cmd, true
	case AdminContentFormDialogCancelMsg:
		m.DCtx.Active = nil
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		), true

	default:
		return m, nil, false
	}
}

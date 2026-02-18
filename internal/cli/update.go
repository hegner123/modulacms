package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Update dispatches messages through the update handler chain and returns the updated model and command.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle user provisioning first if needed
	if m, cmd := m.UpdateProvisioning(msg); cmd != nil {
		return m, cmd
	}

	if m, cmd := m.UpdateLog(msg); cmd != nil {
		return m, cmd
	}
	if m, cmd := m.UpdateTea(msg); cmd != nil {
		return m, cmd
	}
	if m, cmd := m.UpdateState(msg); cmd != nil {
		return m, cmd
	}

	if m, cmd := m.UpdateNavigation(msg); cmd != nil {
		return m, cmd
	}
	if m, cmd := m.UpdateFetch(msg); cmd != nil {
		return m, cmd
	}
	if m, cmd := m.UpdateAdminFetch(msg); cmd != nil {
		return m, cmd
	}
	if m, cmd := m.UpdateForm(msg); cmd != nil {
		return m, cmd
	}
	if m, cmd := m.UpdateDialog(msg); cmd != nil {
		return m, cmd
	}
	if m, cmd := m.UpdateDatabase(msg); cmd != nil {
		return m, cmd
	}
	if m, cmd := m.UpdateCms(msg); cmd != nil {
		return m, cmd
	}
	if m, cmd := m.UpdateAdminCms(msg); cmd != nil {
		return m, cmd
	}
	// Handle editor finished: update the content form dialog field with edited content.
	if editorMsg, ok := msg.(EditorFinishedMsg); ok {
		if m.Logger != nil {
			m.Logger.Finfo(fmt.Sprintf("[editor] EditorFinishedMsg received: fieldIndex=%d, err=%v, contentLen=%d", editorMsg.FieldIndex, editorMsg.Err, len(editorMsg.Content)))
		}
		if m.ContentFormDialogActive && m.ContentFormDialog != nil {
			if editorMsg.Err != nil {
				if m.Logger != nil {
					m.Logger.Ferror(fmt.Sprintf("[editor] editor returned error for field %d", editorMsg.FieldIndex), editorMsg.Err)
				}
				return m, nil
			}
			if editorMsg.FieldIndex < len(m.ContentFormDialog.Fields) {
				field := m.ContentFormDialog.Fields[editorMsg.FieldIndex]
				if m.Logger != nil {
					m.Logger.Finfo(fmt.Sprintf("[editor] applying editor content (%d bytes) to field %d (%s)", len(editorMsg.Content), editorMsg.FieldIndex, field.Label))
				}
				m.ContentFormDialog.Fields[editorMsg.FieldIndex].Bubble.SetValue(editorMsg.Content)
			} else {
				if m.Logger != nil {
					m.Logger.Finfo(fmt.Sprintf("[editor] fieldIndex %d out of range (dialog has %d fields), ignoring", editorMsg.FieldIndex, len(m.ContentFormDialog.Fields)))
				}
			}
			return m, nil
		}
		if m.Logger != nil {
			m.Logger.Finfo("[editor] EditorFinishedMsg received but no active content form dialog, ignoring")
		}
	}

	// When file picker is active, route all input to it.
	if m.FilePickerActive {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "esc" || keyMsg.String() == "ctrl+c" {
				m.FilePickerActive = false
				return m, nil
			}
			var cmd tea.Cmd
			m.FilePicker, cmd = m.FilePicker.Update(msg)

			if didSelect, path := m.FilePicker.DidSelectFile(msg); didSelect {
				m.FilePickerActive = false
				switch m.FilePickerPurpose {
				case FILEPICKER_RESTORE:
					return m, RestoreBackupFromPathCmd(path)
				default:
					return m, MediaUploadCmd(path)
				}
			}

			// Disabled file selection stays in picker (filepicker shows its own error)
			if didSelect, _ := m.FilePicker.DidSelectDisabledFile(msg); didSelect {
				return m, cmd
			}

			return m, cmd
		}
		// Non-key messages still forwarded to filepicker (for directory reads)
		var cmd tea.Cmd
		m.FilePicker, cmd = m.FilePicker.Update(msg)
		return m, cmd
	}

	// When dialog is active, route all key input to the dialog and stop.
	if m.DialogActive && m.Dialog != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			dialog, cmd := m.Dialog.Update(keyMsg)
			m.Dialog = &dialog
			return m, cmd
		}
	}

	// When form dialog is active, route all key input to the form dialog and stop.
	if m.FormDialogActive && m.FormDialog != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			formDialog, cmd := m.FormDialog.Update(keyMsg)
			m.FormDialog = &formDialog
			return m, cmd
		}
	}

	// When content form dialog is active, route all key input to the content form dialog and stop.
	if m.ContentFormDialogActive && m.ContentFormDialog != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			contentFormDialog, cmd := m.ContentFormDialog.Update(keyMsg)
			m.ContentFormDialog = &contentFormDialog
			return m, cmd
		}
	}

	// When user form dialog is active, route all key input to the user form dialog and stop.
	if m.UserFormDialogActive && m.UserFormDialog != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			userFormDialog, cmd := m.UserFormDialog.Update(keyMsg)
			m.UserFormDialog = &userFormDialog
			return m, cmd
		}
	}

	// When database form dialog is active, route all key input to the database form dialog and stop.
	if m.DatabaseFormDialogActive && m.DatabaseFormDialog != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			dbFormDialog, cmd := m.DatabaseFormDialog.Update(keyMsg)
			m.DatabaseFormDialog = &dbFormDialog
			return m, cmd
		}
	}

	// When UIConfig form dialog is active, route all key input to the UIConfig form dialog and stop.
	if m.UIConfigFormDialogActive && m.UIConfigFormDialog != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			uiConfigDialog, cmd := m.UIConfigFormDialog.Update(keyMsg)
			m.UIConfigFormDialog = &uiConfigDialog
			return m, cmd
		}
	}

	return m.PageSpecificMsgHandlers(nil, msg)
}
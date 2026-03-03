package tui

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
	if m, cmd := m.UpdateDeployFetch(msg); cmd != nil {
		return m, cmd
	}
	if m, cmd := m.UpdateDeployCms(msg); cmd != nil {
		return m, cmd
	}
	// Handle editor finished: update the content form dialog field with edited content.
	if editorMsg, ok := msg.(EditorFinishedMsg); ok {
		if m.Logger != nil {
			m.Logger.Finfo(fmt.Sprintf("[editor] EditorFinishedMsg received: fieldIndex=%d, err=%v, contentLen=%d", editorMsg.FieldIndex, editorMsg.Err, len(editorMsg.Content)))
		}
		if cfd, ok := m.ActiveOverlay.(*ContentFormDialogModel); ok && cfd != nil {
			if editorMsg.Err != nil {
				if m.Logger != nil {
					m.Logger.Ferror(fmt.Sprintf("[editor] editor returned error for field %d", editorMsg.FieldIndex), editorMsg.Err)
				}
				return m, nil
			}
			if editorMsg.FieldIndex < len(cfd.Fields) {
				field := cfd.Fields[editorMsg.FieldIndex]
				if m.Logger != nil {
					m.Logger.Finfo(fmt.Sprintf("[editor] applying editor content (%d bytes) to field %d (%s)", len(editorMsg.Content), editorMsg.FieldIndex, field.Label))
				}
				cfd.Fields[editorMsg.FieldIndex].Bubble.SetValue(editorMsg.Content)
			} else {
				if m.Logger != nil {
					m.Logger.Finfo(fmt.Sprintf("[editor] fieldIndex %d out of range (dialog has %d fields), ignoring", editorMsg.FieldIndex, len(cfd.Fields)))
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
		// Recalculate filepicker height on terminal resize.
		if _, ok := msg.(tea.WindowSizeMsg); ok {
			m.FilePicker.SetHeight(filePickerHeight(m.Height))
		}
		// Non-key messages still forwarded to filepicker (for directory reads)
		var cmd tea.Cmd
		m.FilePicker, cmd = m.FilePicker.Update(msg)
		return m, cmd
	}

	// When any modal overlay is active, route all key input to it and stop.
	if m.ActiveOverlay != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			overlay, cmd := m.ActiveOverlay.OverlayUpdate(keyMsg)
			m.ActiveOverlay = overlay
			return m, cmd
		}
	}

	return m.PageSpecificMsgHandlers(nil, msg)
}

package cli

import (
	tea "github.com/charmbracelet/bubbletea"
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
				RunDestructiveActionCmd(m.Config, actionIndex),
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
	default:
		return m, nil
	}
}

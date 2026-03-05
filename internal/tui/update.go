package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
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

	// Screen-based pages: ActiveScreen handles all messages except
	// WindowSizeMsg (processed by UpdateTea above) and provisioning.
	if m.ActiveScreen != nil {
		// Handle messages that mutate root Model state.
		switch typedMsg := msg.(type) {
		case OverlaySetMsg:
			m.ActiveOverlay = typedMsg.Overlay
			return m, nil
		case OverlayClearMsg:
			m.ActiveOverlay = nil
			return m, nil
		case SetPanelFocusMsg:
			m.PanelFocus = typedMsg.Panel
			return m, nil
		case OpenFilePickerMsg:
			fp := filepicker.New()
			switch typedMsg.Purpose {
			case FILEPICKER_RESTORE:
				fp.AllowedTypes = []string{".zip"}
			default:
				fp.AllowedTypes = []string{".png", ".jpg", ".jpeg", ".webp", ".gif"}
			}
			fp.CurrentDirectory, _ = os.UserHomeDir()
			fp.AutoHeight = false
			fp.Height = filePickerHeight(m.Height)
			m.FilePicker = fp
			m.FilePickerActive = true
			m.FilePickerPurpose = typedMsg.Purpose
			return m, m.FilePicker.Init()

		// Navigation messages: delegate to UpdateNavigation so page
		// transitions, history, and ActiveScreen switching all work.
		case NavigateToPage:
			return m.UpdateNavigation(msg)
		case HistoryPop:
			return m.UpdateNavigation(msg)
		case HistoryPush:
			return m.UpdateNavigation(msg)

		// State messages that Screens may emit via constructors.
		case SetLoadingMsg:
			m.Loading = typedMsg.Loading
			return m, nil
		case PageSet:
			m.Page = typedMsg.Page
			m.PageMenu = m.HomepageMenuInit()
			m.ActiveScreen = m.screenForPage(m.Page)
			return m, nil

		// Plugin selection + navigation (emitted by PluginsScreen).
		case SelectPluginAndNavigateMsg:
			m.SelectedPlugin = typedMsg.PluginName
			return m, NavigateToPageCmd(m.PageMap[PLUGINDETAILPAGE])

		// Pipeline selection + navigation (emitted by PipelinesScreen).
		case SelectPipelineAndNavigateMsg:
			// Screen receives the key via PipelineEntriesSet; no Model field needed.
			return m, tea.Batch(
				PipelineEntriesFetchCmd(typedMsg.PipelineKey),
				NavigateToPageCmd(m.PageMap[PIPELINEDETAILPAGE]),
			)

		// Database state messages.
		case SetDatabaseModeMsg:
			m.DatabaseMode = typedMsg.Mode
			return m, nil
		case CursorMsg:
			return m.UpdateState(msg)
		case PageModMsg:
			return m.UpdateState(msg)

		// Fetch error: show dialog
		case FetchErrMsg:
			return m, tea.Batch(
				ErrorSetCmd(typedMsg.Error),
				LoadingStopCmd(),
				LogMessageCmd(fmt.Sprintf("Database fetch error: %s", typedMsg.Error.Error())),
				func() tea.Msg {
					return ActionResultMsg{
						Title:   "Fetch Error",
						Message: typedMsg.Error.Error(),
					}
				},
			)

		// Database and dialog messages need to flow through their handlers.
		case ShowDatabaseFormDialogMsg:
			return m.UpdateDialog(msg)
		case ShowDialogMsg:
			return m.UpdateDialog(msg)
		case DbResMsg:
			return m.UpdateDatabase(msg)

		// Config field edit dialog.
		case ShowConfigFieldEditMsg:
			return m.UpdateDialog(msg)

		// Deploy confirmation dialogs flow through UpdateDialog.
		case DeployConfirmPullMsg:
			return m.UpdateDialog(msg)
		case DeployConfirmPushMsg:
			return m.UpdateDialog(msg)

		// Deploy fetch and operation messages flow through shared handlers,
		// then forward results to the screen for display state updates.
		case DeployEnvsFetchMsg:
			m, cmd := m.UpdateDeployFetch(msg)
			return m, cmd
		case DeployTestConnectionRequestMsg:
			m, cmd := m.UpdateDeployCms(msg)
			if m.ActiveScreen != nil {
				ctx := m.AppCtx()
				screen, screenCmd := m.ActiveScreen.Update(ctx, msg)
				m.ActiveScreen = screen
				if screenCmd != nil {
					return m, tea.Batch(cmd, screenCmd)
				}
			}
			return m, cmd
		case DeployTestConnectionResultMsg:
			m, cmd := m.UpdateDeployCms(msg)
			if m.ActiveScreen != nil {
				ctx := m.AppCtx()
				screen, screenCmd := m.ActiveScreen.Update(ctx, msg)
				m.ActiveScreen = screen
				if screenCmd != nil {
					return m, tea.Batch(cmd, screenCmd)
				}
			}
			return m, cmd
		case DeployPullRequestMsg:
			m, cmd := m.UpdateDeployCms(msg)
			if m.ActiveScreen != nil {
				ctx := m.AppCtx()
				screen, screenCmd := m.ActiveScreen.Update(ctx, msg)
				m.ActiveScreen = screen
				if screenCmd != nil {
					return m, tea.Batch(cmd, screenCmd)
				}
			}
			return m, cmd
		case DeployPullResultMsg:
			m, cmd := m.UpdateDeployCms(msg)
			if m.ActiveScreen != nil {
				ctx := m.AppCtx()
				screen, screenCmd := m.ActiveScreen.Update(ctx, msg)
				m.ActiveScreen = screen
				if screenCmd != nil {
					return m, tea.Batch(cmd, screenCmd)
				}
			}
			return m, cmd
		case DeployPushRequestMsg:
			m, cmd := m.UpdateDeployCms(msg)
			if m.ActiveScreen != nil {
				ctx := m.AppCtx()
				screen, screenCmd := m.ActiveScreen.Update(ctx, msg)
				m.ActiveScreen = screen
				if screenCmd != nil {
					return m, tea.Batch(cmd, screenCmd)
				}
			}
			return m, cmd
		case DeployPushResultMsg:
			m, cmd := m.UpdateDeployCms(msg)
			if m.ActiveScreen != nil {
				ctx := m.AppCtx()
				screen, screenCmd := m.ActiveScreen.Update(ctx, msg)
				m.ActiveScreen = screen
				if screenCmd != nil {
					return m, tea.Batch(cmd, screenCmd)
				}
			}
			return m, cmd

		// Deploy envs set goes directly to screen (no Model-level state).
		case DeployEnvsSet:
			if m.ActiveScreen != nil {
				ctx := m.AppCtx()
				screen, screenCmd := m.ActiveScreen.Update(ctx, msg)
				m.ActiveScreen = screen
				return m, screenCmd
			}
			return m, nil

		// Plugin action results flow through shared handlers.
		case PluginActionRequestMsg:
			return m.UpdateCms(msg)
		case PluginActionCompleteMsg:
			return m.UpdateCms(msg)
		case PluginSyncCapabilitiesRequestMsg:
			return m.UpdateCms(msg)
		case ActionConfirmMsg:
			return m.UpdateDialog(msg)
		case ActionResultMsg:
			return m.UpdateDialog(msg)
		case ShowApproveAllRoutesDialogMsg:
			return m.UpdateDialog(msg)
		case ShowApproveAllHooksDialogMsg:
			return m.UpdateDialog(msg)

		// Content screen: dialog-flow messages → UpdateDialog.
		case ShowRestoreVersionDialogMsg:
			return m.UpdateDialog(msg)
		case ShowRestoreAdminVersionDialogMsg:
			return m.UpdateDialog(msg)
		case ShowDeleteAdminContentDialogMsg:
			return m.UpdateDialog(msg)
		case ShowPublishAdminContentDialogMsg:
			return m.UpdateDialog(msg)
		case ShowDeleteAdminContentFieldDialogMsg:
			return m.UpdateDialog(msg)
		case ShowDeleteContentDialogMsg:
			return m.UpdateDialog(msg)
		case ShowEditSingleFieldDialogMsg:
			return m.UpdateDialog(msg)
		case ShowAddContentFieldDialogMsg:
			return m.UpdateDialog(msg)
		case ShowDeleteContentFieldDialogMsg:
			return m.UpdateDialog(msg)
		case ShowMoveAdminContentDialogMsg:
			return m.UpdateDialog(msg)
		case ShowMoveContentDialogMsg:
			return m.UpdateDialog(msg)
		case ShowEditAdminSingleFieldDialogMsg:
			return m.UpdateDialog(msg)
		case ShowAddAdminContentFieldDialogMsg:
			return m.UpdateDialog(msg)
		case ShowCreateRouteWithContentDialogMsg:
			return m.UpdateDialog(msg)
		case AdminContentFormDialogAcceptMsg:
			return m.UpdateDialog(msg)

		// Content screen: CMS operations → UpdateCms/UpdateAdminCms.
		case FetchContentForEditMsg:
			return m.UpdateCms(msg)
		case BuildContentFormMsg:
			return m.UpdateCms(msg)
		case ListVersionsRequestMsg:
			return m.UpdateCms(msg)
		case ReorderSiblingRequestMsg:
			return m.UpdateCms(msg)
		case CopyContentRequestMsg:
			return m.UpdateCms(msg)
		case ConfirmedRestoreVersionMsg:
			return m.UpdateCms(msg)
		case UpdateContentFromDialogRequestMsg:
			return m.UpdateCms(msg)
		case ConfirmedDeleteAdminContentMsg:
			return m.UpdateAdminCms(msg)
		case AdminReorderSiblingRequestMsg:
			return m.UpdateAdminCms(msg)
		case AdminCopyContentRequestMsg:
			return m.UpdateAdminCms(msg)
		case AdminMoveContentRequestMsg:
			return m.UpdateAdminCms(msg)
		case ConfirmedPublishAdminContentMsg:
			return m.UpdateAdminCms(msg)
		case ConfirmedUnpublishAdminContentMsg:
			return m.UpdateAdminCms(msg)
		case AdminListVersionsRequestMsg:
			return m.UpdateAdminCms(msg)
		case ConfirmedRestoreAdminVersionMsg:
			return m.UpdateAdminCms(msg)
		case ConfirmedDeleteAdminContentFieldMsg:
			return m.UpdateAdminCms(msg)

		// Content screen: admin form building.
		case AdminBuildContentFormMsg:
			return m.UpdateDialog(msg)
		case AdminFetchContentForEditMsg:
			return m.UpdateAdminCms(msg)
		case ShowEditAdminContentFormDialogMsg:
			return m.UpdateDialog(msg)
		case ContentFormDialogAcceptMsg:
			return m.UpdateDialog(msg)
		case ContentFormDialogCancelMsg:
			return m.UpdateDialog(msg)
		case ContentCreatedFromDialogMsg:
			return m.UpdateDialog(msg)
		case ContentUpdatedFromDialogMsg:
			return m.UpdateDialog(msg)
		case ShowEditContentFormDialogMsg:
			return m.UpdateDialog(msg)
		}

		// Global key handling for Screen-based pages (screen mode + accordion).
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			km := m.Config.KeyBindings
			key := keyMsg.String()
			if km.Matches(key, config.ActionAccordion) {
				m.AccordionEnabled = !m.AccordionEnabled
				return m, nil
			}
			if km.Matches(key, config.ActionScreenToggle) {
				if m.ScreenMode == ScreenFull {
					m.ScreenMode = ScreenNormal
				} else {
					m.ScreenMode = ScreenFull
				}
				m.ScreenModeManual = true
				return m, nil
			}
			if km.Matches(key, config.ActionScreenNext) {
				m.ScreenMode = (m.ScreenMode + 1) % 3
				m.ScreenModeManual = true
				return m, nil
			}
			if km.Matches(key, config.ActionScreenReset) {
				m.ScreenMode = ScreenNormal
				m.ScreenModeManual = false
				return m, nil
			}
			if km.Matches(key, config.ActionAdminToggle) {
				m.AdminMode = !m.AdminMode
				return m, nil
			}
		}

		// Overlay intercepts key input even for Screen-based pages
		if m.ActiveOverlay != nil {
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				overlay, cmd := m.ActiveOverlay.OverlayUpdate(keyMsg)
				m.ActiveOverlay = overlay
				return m, cmd
			}
		}

		// File picker intercepts key input
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
				if didSelect, _ := m.FilePicker.DidSelectDisabledFile(msg); didSelect {
					return m, cmd
				}
				return m, cmd
			}
		}

		// Delegate everything else to the screen
		ctx := m.AppCtx()
		screen, cmd := m.ActiveScreen.Update(ctx, msg)
		m.ActiveScreen = screen
		return m, cmd
	}

	if m, cmd := m.UpdateState(msg); cmd != nil {
		return m, cmd
	}

	if m, cmd := m.UpdateNavigation(msg); cmd != nil {
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

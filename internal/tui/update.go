package tui

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/filepicker"
	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/utility"
)

// handleFilePicker routes messages to the file picker when it is active.
// Returns (model, cmd, true) if the file picker consumed the message.
func (m Model) handleFilePicker(msg tea.Msg) (Model, tea.Cmd, bool) {
	if !m.FilePickerActive {
		return m, nil, false
	}
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if keyMsg.String() == "esc" || keyMsg.String() == "ctrl+c" {
			m.FilePickerActive = false
			return m, nil, true
		}
		var cmd tea.Cmd
		m.FilePicker, cmd = m.FilePicker.Update(msg)
		if didSelect, path := m.FilePicker.DidSelectFile(msg); didSelect {
			m.FilePickerActive = false
			switch m.FilePickerPurpose {
			case FILEPICKER_RESTORE:
				return m, RestoreBackupFromPathCmd(path), true
			default:
				return m, MediaUploadCmd(path), true
			}
		}
		if didSelect, _ := m.FilePicker.DidSelectDisabledFile(msg); didSelect {
			return m, cmd, true
		}
		return m, cmd, true
	}
	// Recalculate filepicker height on terminal resize.
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		m.FilePicker.SetHeight(filePickerHeight(m.Height))
	}
	// Non-key messages still forwarded to filepicker (for directory reads)
	var cmd tea.Cmd
	m.FilePicker, cmd = m.FilePicker.Update(msg)
	return m, cmd, true
}

// Update dispatches messages through the update handler chain and returns the updated model and command.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Ctrl+c always force-quits, regardless of overlay, focus, or screen state.
	// Esc shows a quit confirmation dialog (or quits immediately if one is already showing).
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			// Double-esc from quit dialog exits immediately.
			if d, ok := m.ActiveOverlay.(*DialogModel); ok && d != nil && d.Action == DIALOGQUITCONFIRM {
				return m, tea.Quit
			}
			// Let active overlays (form dialogs, confirmation dialogs) handle
			// their own escape — they return cancel messages.
			if m.ActiveOverlay != nil {
				break
			}
			// Let the file picker handle its own escape.
			if m.FilePickerActive {
				break
			}
			// Let screen-level search handle escape before showing quit.
			if ds, ok := m.ActiveScreen.(*DatatypesScreen); ok && ds.Searching {
				break
			}
			if ms, ok := m.ActiveScreen.(*MediaScreen); ok && ms.Searching {
				break
			}
			// Default: show quit confirmation.
			return m.UpdateDialog(ShowQuitConfirmDialogMsg{})
		}
	}

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

	// All pages use the Screen interface. ActiveScreen is set at startup
	// and on every page navigation; it should never be nil here.
	if m.ActiveScreen == nil {
		return m, nil
	}

	// Handle messages that mutate root Model state.
	switch typedMsg := msg.(type) {
	case OverlaySetMsg:
		m.ActiveOverlay = typedMsg.Overlay
		return m, nil
	case OverlayClearMsg:
		m.ActiveOverlay = nil
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
		fp.SetHeight(filePickerHeight(m.Height))
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
		return m.UpdateState(msg)

	// State messages that Screens may emit via constructors.
	case SetLoadingMsg:
		m.Loading = typedMsg.Loading
		return m, nil
	case PageSet:
		m.Page = typedMsg.Page
		m.PageMenu = m.HomepageMenuInit()
		m.ActiveScreen = m.screenForPage(m.Page)
		if m.Page.Index == HOMEPAGE {
			return m, HomeDashboardFetchCmd(m.DB)
		}
		return m, nil

	// Plugin selection + navigation (emitted by PluginsScreen).
	case SelectPluginAndNavigateMsg:
		m.SelectedPlugin = typedMsg.PluginName
		return m, NavigateToPageCmd(m.PageMap[PLUGINDETAILPAGE])

	// Plugin TUI screen navigation (emitted by PluginTUIScreen or sidebar).
	case NavigateToPluginScreenMsg:
		screen := NewPluginTUIScreen(typedMsg.PluginName, typedMsg.ScreenName, typedMsg.Params)
		// Push current state onto history, then set the plugin screen as active.
		m.ActiveScreen = screen
		return m, tea.Batch(
			HistoryPushCmd(PageHistory{Page: m.Page, Cursor: m.Cursor, Menu: m.PageMenu, Screen: m.ActiveScreen}),
			PageSetCmd(m.PageMap[PLUGINTUIPAGE]),
			PluginScreenSetupCmd(typedMsg.PluginName, typedMsg.ScreenName, typedMsg.Params, m.Width, m.Height, m.PluginManager),
		)

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
	case DatabaseDeleteEntry:
		return m.UpdateDatabase(msg)
	case DatabaseInsertEntry:
		return m.UpdateDatabase(msg)
	case DatabaseUpdateEntry:
		return m.UpdateDatabase(msg)
	case DatabaseFormDialogAcceptMsg:
		return m.UpdateDialog(msg)

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
		ctx := m.AppCtx()
		screen, screenCmd := m.ActiveScreen.Update(ctx, msg)
		m.ActiveScreen = screen
		if screenCmd != nil {
			return m, tea.Batch(cmd, screenCmd)
		}
		return m, cmd
	case DeployTestConnectionResultMsg:
		m, cmd := m.UpdateDeployCms(msg)
		ctx := m.AppCtx()
		screen, screenCmd := m.ActiveScreen.Update(ctx, msg)
		m.ActiveScreen = screen
		if screenCmd != nil {
			return m, tea.Batch(cmd, screenCmd)
		}
		return m, cmd
	case DeployPullRequestMsg:
		m, cmd := m.UpdateDeployCms(msg)
		ctx := m.AppCtx()
		screen, screenCmd := m.ActiveScreen.Update(ctx, msg)
		m.ActiveScreen = screen
		if screenCmd != nil {
			return m, tea.Batch(cmd, screenCmd)
		}
		return m, cmd
	case DeployPullResultMsg:
		m, cmd := m.UpdateDeployCms(msg)
		ctx := m.AppCtx()
		screen, screenCmd := m.ActiveScreen.Update(ctx, msg)
		m.ActiveScreen = screen
		if screenCmd != nil {
			return m, tea.Batch(cmd, screenCmd)
		}
		return m, cmd
	case DeployPushRequestMsg:
		m, cmd := m.UpdateDeployCms(msg)
		ctx := m.AppCtx()
		screen, screenCmd := m.ActiveScreen.Update(ctx, msg)
		m.ActiveScreen = screen
		if screenCmd != nil {
			return m, tea.Batch(cmd, screenCmd)
		}
		return m, cmd
	case DeployPushResultMsg:
		m, cmd := m.UpdateDeployCms(msg)
		ctx := m.AppCtx()
		screen, screenCmd := m.ActiveScreen.Update(ctx, msg)
		m.ActiveScreen = screen
		if screenCmd != nil {
			return m, tea.Batch(cmd, screenCmd)
		}
		return m, cmd

	// Deploy envs set goes directly to screen (no Model-level state).
	case DeployEnvsSet:
		ctx := m.AppCtx()
		screen, screenCmd := m.ActiveScreen.Update(ctx, msg)
		m.ActiveScreen = screen
		return m, screenCmd

	// Datatype/field dialog show messages → UpdateDialog.
	case ShowFormDialogMsg:
		return m.UpdateDialog(msg)
	case ShowFieldFormDialogMsg:
		return m.UpdateDialog(msg)
	case ShowEditDatatypeDialogMsg:
		return m.UpdateDialog(msg)
	case ShowDeleteDatatypeDialogMsg:
		return m.UpdateDialog(msg)
	case ShowEditFieldDialogMsg:
		return m.UpdateDialog(msg)
	case ShowDeleteFieldDialogMsg:
		return m.UpdateDialog(msg)
	case ShowAdminFormDialogMsg:
		return m.UpdateDialog(msg)
	case ShowEditAdminDatatypeDialogMsg:
		return m.UpdateDialog(msg)
	case ShowDeleteAdminDatatypeDialogMsg:
		return m.UpdateDialog(msg)
	case ShowEditAdminFieldDialogMsg:
		return m.UpdateDialog(msg)
	case ShowDeleteAdminFieldDialogMsg:
		return m.UpdateDialog(msg)

	// Datatype/field CRUD requests → UpdateCms/UpdateAdminCms.
	case CreateDatatypeFromDialogRequestMsg:
		return m.UpdateCms(msg)
	case UpdateDatatypeFromDialogRequestMsg:
		return m.UpdateCms(msg)
	case DeleteDatatypeRequestMsg:
		return m.UpdateCms(msg)
	case CreateFieldFromDialogRequestMsg:
		return m.UpdateCms(msg)
	case UpdateFieldFromDialogRequestMsg:
		return m.UpdateCms(msg)
	case DeleteFieldRequestMsg:
		return m.UpdateCms(msg)
	case CreateAdminDatatypeFromDialogRequestMsg:
		return m.UpdateAdminCms(msg)
	case UpdateAdminDatatypeFromDialogRequestMsg:
		return m.UpdateAdminCms(msg)
	case DeleteAdminDatatypeRequestMsg:
		return m.UpdateAdminCms(msg)
	case CreateAdminFieldFromDialogRequestMsg:
		return m.UpdateAdminCms(msg)
	case UpdateAdminFieldFromDialogRequestMsg:
		return m.UpdateAdminCms(msg)
	case DeleteAdminFieldRequestMsg:
		return m.UpdateAdminCms(msg)

	// Datatype/field CRUD results → UpdateDialog/UpdateAdminCms.
	case DatatypeCreatedFromDialogMsg:
		return m.UpdateDialog(msg)
	case DatatypeUpdatedFromDialogMsg:
		return m.UpdateDialog(msg)
	case DatatypeDeletedMsg:
		return m.UpdateDialog(msg)
	case FieldCreatedFromDialogMsg:
		return m.UpdateDialog(msg)
	case FieldUpdatedFromDialogMsg:
		return m.UpdateDialog(msg)
	case FieldDeletedMsg:
		return m.UpdateDialog(msg)
	case AdminDatatypeCreatedFromDialogMsg:
		return m.UpdateAdminCms(msg)
	case AdminDatatypeUpdatedFromDialogMsg:
		return m.UpdateAdminCms(msg)
	case AdminDatatypeDeletedMsg:
		return m.UpdateAdminCms(msg)
	case AdminFieldCreatedFromDialogMsg:
		return m.UpdateAdminCms(msg)
	case AdminFieldUpdatedFromDialogMsg:
		return m.UpdateAdminCms(msg)
	case AdminFieldDeletedMsg:
		return m.UpdateAdminCms(msg)

	// Route dialog show messages → UpdateDialog.
	case ShowRouteFormDialogMsg:
		return m.UpdateDialog(msg)
	case ShowEditRouteDialogMsg:
		return m.UpdateDialog(msg)
	case ShowDeleteRouteDialogMsg:
		return m.UpdateDialog(msg)
	case ShowEditAdminRouteDialogMsg:
		return m.UpdateDialog(msg)
	case ShowDeleteAdminRouteDialogMsg:
		return m.UpdateDialog(msg)

	// Route CRUD requests → UpdateCms/UpdateAdminCms.
	case CreateRouteFromDialogRequestMsg:
		return m.UpdateCms(msg)
	case UpdateRouteFromDialogRequestMsg:
		return m.UpdateCms(msg)
	case DeleteRouteRequestMsg:
		return m.UpdateCms(msg)
	case CreateAdminRouteFromDialogRequestMsg:
		return m.UpdateAdminCms(msg)
	case UpdateAdminRouteFromDialogRequestMsg:
		return m.UpdateAdminCms(msg)
	case DeleteAdminRouteRequestMsg:
		return m.UpdateAdminCms(msg)

	// Route CRUD results → UpdateDialog/UpdateAdminCms.
	case RouteCreatedFromDialogMsg:
		return m.UpdateDialog(msg)
	case RouteUpdatedFromDialogMsg:
		return m.UpdateDialog(msg)
	case RouteDeletedMsg:
		return m.UpdateDialog(msg)
	case AdminRouteCreatedFromDialogMsg:
		return m.UpdateAdminCms(msg)
	case AdminRouteUpdatedFromDialogMsg:
		return m.UpdateAdminCms(msg)
	case AdminRouteDeletedMsg:
		return m.UpdateAdminCms(msg)

	// User dialog show messages → UpdateDialog.
	case ShowUserFormDialogMsg:
		return m.UpdateDialog(msg)
	case ShowEditUserDialogMsg:
		return m.UpdateDialog(msg)
	case ShowDeleteUserDialogMsg:
		return m.UpdateDialog(msg)
	case UserFormDialogAcceptMsg:
		return m.UpdateDialog(msg)

	// User CRUD requests → UpdateCms.
	case CreateUserFromDialogRequestMsg:
		return m.UpdateCms(msg)
	case UpdateUserFromDialogRequestMsg:
		return m.UpdateCms(msg)
	case DeleteUserRequestMsg:
		return m.UpdateCms(msg)

	// User CRUD results → UpdateDialog.
	case UserCreatedFromDialogMsg:
		return m.UpdateDialog(msg)
	case UserUpdatedFromDialogMsg:
		return m.UpdateDialog(msg)
	case UserDeletedMsg:
		return m.UpdateDialog(msg)

	// Media dialog + CRUD messages.
	case ShowDeleteMediaDialogMsg:
		return m.UpdateDialog(msg)
	case DeleteMediaRequestMsg:
		return m.UpdateCms(msg)
	case MediaDeletedMsg:
		return m.UpdateDialog(msg)

	// Media folder dialog + CRUD messages.
	case ShowCreateMediaFolderDialogMsg:
		return m.UpdateDialog(msg)
	case ShowRenameMediaFolderDialogMsg:
		return m.UpdateDialog(msg)
	case ShowDeleteMediaFolderDialogMsg:
		return m.UpdateDialog(msg)
	case ShowMoveMediaToFolderDialogMsg:
		return m.UpdateDialog(msg)
	case ShowMoveMediaToFolderPickerMsg:
		return m.UpdateDialog(msg)
	case MediaFolderNameDialogCancelMsg:
		return m.UpdateDialog(msg)
	case MoveMediaFolderDialogCancelMsg:
		return m.UpdateDialog(msg)
	case CreateMediaFolderRequestMsg:
		return m.UpdateDialog(msg)
	case RenameMediaFolderRequestMsg:
		return m.UpdateDialog(msg)
	case DeleteMediaFolderRequestMsg:
		return m.UpdateDialog(msg)
	case MoveMediaToFolderRequestMsg:
		return m.UpdateDialog(msg)
	case MediaFolderCreatedMsg:
		return m.UpdateDialog(msg)
	case MediaFolderRenamedMsg:
		return m.UpdateDialog(msg)
	case MediaFolderDeletedMsg:
		return m.UpdateDialog(msg)
	case MediaMovedToFolderMsg:
		return m.UpdateDialog(msg)

	// Webhook dialog show messages → UpdateDialog.
	case ShowWebhookFormDialogMsg:
		return m.UpdateDialog(msg)
	case ShowEditWebhookDialogMsg:
		return m.UpdateDialog(msg)
	case ShowDeleteWebhookDialogMsg:
		return m.UpdateDialog(msg)
	case WebhookFormDialogAcceptMsg:
		return m.UpdateDialog(msg)
	case WebhookFormDialogCancelMsg:
		return m.UpdateDialog(msg)

	// Webhook CRUD requests → UpdateCms.
	case CreateWebhookFromDialogRequestMsg:
		return m.UpdateCms(msg)
	case UpdateWebhookFromDialogRequestMsg:
		return m.UpdateCms(msg)
	case DeleteWebhookRequestMsg:
		return m.UpdateCms(msg)

	// Webhook CRUD results → UpdateDialog.
	case WebhookCreatedMsg:
		return m.UpdateDialog(msg)
	case WebhookUpdatedMsg:
		return m.UpdateDialog(msg)
	case WebhookDeletedMsg:
		return m.UpdateDialog(msg)

	// Token dialog show messages → UpdateDialog.
	case ShowTokenFormDialogMsg:
		return m.UpdateDialog(msg)
	case TokenFormDialogAcceptMsg:
		return m.UpdateDialog(msg)
	case TokenFormDialogCancelMsg:
		return m.UpdateDialog(msg)
	case ShowDeleteTokenDialogMsg:
		return m.UpdateDialog(msg)

	// Token CRUD requests → UpdateCms.
	case CreateTokenFromDialogRequestMsg:
		return m.UpdateCms(msg)
	case DeleteTokenRequestMsg:
		return m.UpdateCms(msg)

	// Token CRUD results → screen (fall through to ActiveScreen.Update).
	// TokenCreatedFromDialogMsg carries the one-time raw token value that
	// the screen needs to display, so it must reach the screen directly.
	// TokenDeletedMsg similarly refreshes the list via the screen.

	// User OAuth dialog messages → UpdateDialog.
	case ShowUnlinkOauthDialogMsg:
		return m.UpdateDialog(msg)

	// User OAuth CRUD requests → UpdateCms.
	case UnlinkOauthRequestMsg:
		return m.UpdateCms(msg)

	// User OAuth results → screen (fall through to ActiveScreen.Update).

	// Media dimension dialog messages → UpdateDialog.
	case ShowMediaDimensionFormDialogMsg:
		return m.UpdateDialog(msg)
	case ShowEditMediaDimensionDialogMsg:
		return m.UpdateDialog(msg)
	case ShowDeleteMediaDimensionDialogMsg:
		return m.UpdateDialog(msg)
	case MediaDimensionFormDialogAcceptMsg:
		return m.UpdateDialog(msg)
	case MediaDimensionFormDialogCancelMsg:
		return m.UpdateDialog(msg)

	// Media dimension CRUD requests → UpdateCms.
	case CreateMediaDimensionFromDialogRequestMsg:
		return m.UpdateCms(msg)
	case UpdateMediaDimensionFromDialogRequestMsg:
		return m.UpdateCms(msg)
	case DeleteMediaDimensionRequestMsg:
		return m.UpdateCms(msg)

	// Media dimension CRUD results → screen (fall through to ActiveScreen.Update).

	// Session dialog show messages → UpdateDialog.
	case ShowDeleteSessionDialogMsg:
		return m.UpdateDialog(msg)

	// Session CRUD requests → UpdateCms.
	case DeleteSessionRequestMsg:
		return m.UpdateCms(msg)

	// Session CRUD results → screen (fall through to ActiveScreen.Update).
	// SessionDeletedMsg refreshes the list via the screen.

	// Field type dialog show messages → UpdateDialog.
	case ShowDeleteFieldTypeDialogMsg:
		return m.UpdateDialog(msg)
	case ShowEditFieldTypeDialogMsg:
		return m.UpdateDialog(msg)
	case ShowDeleteAdminFieldTypeDialogMsg:
		return m.UpdateDialog(msg)
	case ShowEditAdminFieldTypeDialogMsg:
		return m.UpdateDialog(msg)

	// Field type CRUD → UpdateAdminCms.
	case CreateFieldTypeFromDialogRequestMsg:
		return m.UpdateAdminCms(msg)
	case UpdateFieldTypeFromDialogRequestMsg:
		return m.UpdateAdminCms(msg)
	case DeleteFieldTypeRequestMsg:
		return m.UpdateAdminCms(msg)
	case FieldTypeCreatedFromDialogMsg:
		return m.UpdateAdminCms(msg)
	case FieldTypeUpdatedFromDialogMsg:
		return m.UpdateAdminCms(msg)
	case FieldTypeDeletedMsg:
		return m.UpdateAdminCms(msg)
	case CreateAdminFieldTypeFromDialogRequestMsg:
		return m.UpdateAdminCms(msg)
	case UpdateAdminFieldTypeFromDialogRequestMsg:
		return m.UpdateAdminCms(msg)
	case DeleteAdminFieldTypeRequestMsg:
		return m.UpdateAdminCms(msg)
	case AdminFieldTypeCreatedFromDialogMsg:
		return m.UpdateAdminCms(msg)
	case AdminFieldTypeUpdatedFromDialogMsg:
		return m.UpdateAdminCms(msg)
	case AdminFieldTypeDeletedMsg:
		return m.UpdateAdminCms(msg)

	// Validation dialog show messages -> UpdateDialog.
	case ShowDeleteValidationDialogMsg:
		return m.UpdateDialog(msg)
	case ShowEditValidationDialogMsg:
		return m.UpdateDialog(msg)
	case ShowDeleteAdminValidationDialogMsg:
		return m.UpdateDialog(msg)
	case ShowEditAdminValidationDialogMsg:
		return m.UpdateDialog(msg)

	// Validation CRUD -> UpdateAdminCms.
	case CreateValidationFromDialogRequestMsg:
		return m.UpdateAdminCms(msg)
	case UpdateValidationFromDialogRequestMsg:
		return m.UpdateAdminCms(msg)
	case DeleteValidationRequestMsg:
		return m.UpdateAdminCms(msg)
	case ValidationCreatedFromDialogMsg:
		return m.UpdateAdminCms(msg)
	case ValidationUpdatedFromDialogMsg:
		return m.UpdateAdminCms(msg)
	case ValidationDeletedMsg:
		return m.UpdateAdminCms(msg)
	case CreateAdminValidationFromDialogRequestMsg:
		return m.UpdateAdminCms(msg)
	case UpdateAdminValidationFromDialogRequestMsg:
		return m.UpdateAdminCms(msg)
	case DeleteAdminValidationRequestMsg:
		return m.UpdateAdminCms(msg)
	case AdminValidationCreatedFromDialogMsg:
		return m.UpdateAdminCms(msg)
	case AdminValidationUpdatedFromDialogMsg:
		return m.UpdateAdminCms(msg)
	case AdminValidationDeletedMsg:
		return m.UpdateAdminCms(msg)

	// Title font cycling (Home + CMS Menu screens).
	case TitleFontMsg:
		return m.UpdateState(msg)

	// Actions screen: file picker for backup restore.
	case OpenFilePickerForRestoreMsg:
		return m.UpdateDialog(msg)

	// Datatype/field reorder requests flow through CMS handlers.
	case ReorderDatatypeRequestMsg:
		return m.UpdateCms(msg)
	case ReorderAdminDatatypeRequestMsg:
		return m.UpdateAdminCms(msg)
	case ReorderFieldRequestMsg:
		return m, HandleReorderFieldStandalone(*m.Config, m.UserID, m.DB, typedMsg)
	case ReorderAdminFieldRequestMsg:
		return m, HandleReorderAdminFieldStandalone(*m.Config, m.UserID, m.DB, typedMsg)

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
	case QuickstartConfirmMsg:
		return m.UpdateDialog(msg)
	case DialogAcceptMsg:
		return m.UpdateDialog(msg)
	case DialogCancelMsg:
		return m.UpdateDialog(msg)
	case FocusSet:
		return m.UpdateState(msg)
	case ShowApproveAllRoutesDialogMsg:
		return m.UpdateDialog(msg)
	case ShowApproveAllHooksDialogMsg:
		return m.UpdateDialog(msg)
	case ShowPluginConfirmDialogMsg:
		return m.UpdateDialog(msg)

	// Plugin TUI screen messages — forward to ActiveScreen.
	case PluginScreenInitMsg, PluginScreenErrorMsg, PluginDataMsg, PluginDialogResponseMsg:
		ctx := m.AppCtx()
		screen, cmd := m.ActiveScreen.Update(ctx, msg)
		m.ActiveScreen = screen
		return m, cmd

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
	case ShowCreateAdminRouteWithContentDialogMsg:
		return m.UpdateDialog(msg)
	case CreateAdminRouteWithContentRequestMsg:
		return m.UpdateAdminCms(msg)
	case AdminRouteWithContentCreatedMsg:
		return m.UpdateAdminCms(msg)
	case FormDialogAcceptMsg:
		return m.UpdateDialog(msg)
	case FormDialogCancelMsg:
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

	// Content screen: route+content creation flow → UpdateCms.
	case CreateRouteWithContentRequestMsg:
		return m.UpdateCms(msg)
	case RouteWithContentCreatedMsg:
		return m.UpdateDialog(msg)

	// Content screen: content CRUD operations → UpdateCms.
	case CreateContentFromDialogRequestMsg:
		return m.UpdateCms(msg)
	case ContentCreatedMsg:
		return m.UpdateCms(msg)
	case ContentCreatedWithErrorsMsg:
		return m.UpdateCms(msg)
	case MoveContentRequestMsg:
		return m.UpdateCms(msg)
	case ContentMovedMsg:
		return m.UpdateCms(msg)
	case DeleteContentRequestMsg:
		return m.UpdateCms(msg)
	case ContentDeletedMsg:
		return m.UpdateCms(msg)

	// Content screen: publish flow → UpdateDialog/UpdateCms.
	case ShowPublishDialogMsg:
		return m.UpdateDialog(msg)
	case TogglePublishRequestMsg:
		return m.UpdateCms(msg)
	case ConfirmedPublishMsg:
		return m.UpdateCms(msg)
	case ConfirmedUnpublishMsg:
		return m.UpdateCms(msg)
	case PublishCompletedMsg:
		return m.UpdateCms(msg)
	case UnpublishCompletedMsg:
		return m.UpdateCms(msg)
	case ContentPublishToggledMsg:
		return m.UpdateCms(msg)

	// Content screen: child datatype selection → UpdateDialog.
	case ShowChildDatatypeDialogMsg:
		return m.UpdateDialog(msg)
	case ChildDatatypeSelectedMsg:
		// Route to screen — the ContentScreen needs cursor/tree context
		// to determine the parent content ID for the new child node.
		ctx := m.AppCtx()
		screen, cmd := m.ActiveScreen.Update(ctx, msg)
		m.ActiveScreen = screen
		return m, cmd
	case ShowContentFormDialogMsg:
		return m.UpdateDialog(msg)

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

	// Editor finished: update the content form dialog field with edited content.
	case EditorFinishedMsg:
		if m.Logger != nil {
			m.Logger.Finfo(fmt.Sprintf("[editor] EditorFinishedMsg received: fieldIndex=%d, err=%v, contentLen=%d", typedMsg.FieldIndex, typedMsg.Err, len(typedMsg.Content)))
		}
		if cfd, ok := m.ActiveOverlay.(*ContentFormDialogModel); ok && cfd != nil {
			if typedMsg.Err != nil {
				if m.Logger != nil {
					m.Logger.Ferror(fmt.Sprintf("[editor] editor returned error for field %d", typedMsg.FieldIndex), typedMsg.Err)
				}
				return m, nil
			}
			if typedMsg.FieldIndex < len(cfd.Fields) {
				field := cfd.Fields[typedMsg.FieldIndex]
				if m.Logger != nil {
					m.Logger.Finfo(fmt.Sprintf("[editor] applying editor content (%d bytes) to field %d (%s)", len(typedMsg.Content), typedMsg.FieldIndex, field.Label))
				}
				cfd.Fields[typedMsg.FieldIndex].Bubble.SetValue(typedMsg.Content)
			} else if m.Logger != nil {
				m.Logger.Finfo(fmt.Sprintf("[editor] fieldIndex %d out of range (dialog has %d fields), ignoring", typedMsg.FieldIndex, len(cfd.Fields)))
			}
			return m, nil
		}
		if m.Logger != nil {
			m.Logger.Finfo("[editor] EditorFinishedMsg received but no active content form dialog, ignoring")
		}
		return m, nil
	}

	// Overlay intercepts key input even for Screen-based pages.
	// Must run before global keys so modals receive all keystrokes.
	if m.ActiveOverlay != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			overlay, cmd := m.ActiveOverlay.OverlayUpdate(keyMsg)
			m.ActiveOverlay = overlay
			return m, cmd
		}
		// Forward non-key messages (cursor blink, timer ticks) to overlays
		// that implement OverlayTicker. Without this, text input cursors
		// freeze and typed text may not render until focus changes.
		if ticker, ok := m.ActiveOverlay.(OverlayTicker); ok {
			overlay, cmd := ticker.OverlayTick(msg)
			m.ActiveOverlay = overlay
			return m, cmd
		}
	}

	// Global key handling for Screen-based pages (screen mode + accordion).
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
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
			m.ScreenMode = (m.ScreenMode + 1) % 2
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

	// File picker intercepts all input when active
	if m, cmd, handled := m.handleFilePicker(msg); handled {
		return m, cmd
	}

	// Delegate everything else to the screen
	if _, ok := msg.(HomeDashboardDataMsg); ok {
		utility.DefaultLogger.Fdebug(fmt.Sprintf("[home] root Update: delegating HomeDashboardDataMsg to ActiveScreen (page=%d)", m.ActiveScreen.PageIndex()))
	}
	ctx := m.AppCtx()
	screen, cmd := m.ActiveScreen.Update(ctx, msg)
	m.ActiveScreen = screen
	return m, cmd
}

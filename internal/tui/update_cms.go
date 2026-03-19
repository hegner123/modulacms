package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
)

// CmsUpdate signals a CMS-specific operation update.
type CmsUpdate struct{}

// NewCmsUpdate returns a command that creates a CmsUpdate message.
func NewCmsUpdate() tea.Cmd {
	return func() tea.Msg {
		return CmsUpdate{}
	}
}

// UpdateCms handles CMS-specific operations including content tree, datatypes, routes, media, and users.
func (m Model) UpdateCms(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case CreateDatatypeFromDialogRequestMsg:
		return m, m.HandleCreateDatatypeFromDialog(msg)
	case CreateFieldFromDialogRequestMsg:
		return m, m.HandleCreateFieldFromDialog(msg)
	case CreateRouteFromDialogRequestMsg:
		return m, m.HandleCreateRouteFromDialog(msg)
	case ReorderDatatypeRequestMsg:
		return m, m.HandleReorderDatatype(msg)
	case UpdateDatatypeFromDialogRequestMsg:
		return m, m.HandleUpdateDatatypeFromDialog(msg)
	case UpdateFieldFromDialogRequestMsg:
		return m, m.HandleUpdateFieldFromDialog(msg)
	case UpdateFieldUIConfigRequestMsg:
		return m, m.HandleUpdateFieldUIConfig(msg)
	case UpdateRouteFromDialogRequestMsg:
		return m, m.HandleUpdateRouteFromDialog(msg)
	case CreateRouteWithContentRequestMsg:
		return m, m.HandleCreateRouteWithContent(msg)
	case InitializeRouteContentRequestMsg:
		return m, m.HandleInitializeRouteContent(msg)
	case CmsEditDatatypeLoadMsg:
		return m, CmsEditDatatypeFormCmd(msg.Datatype)
	case DatatypeUpdateSaveMsg:
		d := m.DB
		cfg := m.Config
		userID := m.UserID
		datatypeID := msg.DatatypeID
		parentStr := msg.Parent
		name := msg.Name
		label := msg.Label
		dtType := msg.Type
		return m, func() tea.Msg {
			ctx := context.Background()
			ac := middleware.AuditContextFromCLI(*cfg, userID)

			// Fetch existing record to preserve unchanged values
			existing, err := d.GetDatatype(datatypeID)
			if err != nil {
				return DatatypeUpdateFailedMsg{Error: fmt.Errorf("failed to get datatype for update: %w", err)}
			}

			parentID := existing.ParentID
			if parentStr != "" {
				parentID = types.NullableDatatypeID{
					ID:    types.DatatypeID(parentStr),
					Valid: true,
				}
			}
			// Preserve existing Name when not provided (legacy huh.Form path has no Name input)
			updatedName := name
			if updatedName == "" {
				updatedName = existing.Name
			}
			params := db.UpdateDatatypeParams{
				DatatypeID:   datatypeID,
				ParentID:     parentID,
				SortOrder:    existing.SortOrder,
				Name:         updatedName,
				Label:        label,
				Type:         dtType,
				AuthorID:     existing.AuthorID,
				DateCreated:  existing.DateCreated,
				DateModified: types.TimestampNow(),
			}
			_, err = d.UpdateDatatype(ctx, ac, params)
			if err != nil {
				return DatatypeUpdateFailedMsg{Error: err}
			}
			return DatatypeUpdatedMsg{DatatypeID: datatypeID, Label: label}
		}
	case DatatypeUpdatedMsg:
		datatypesPage := m.PageMap[DATATYPES]
		return m, tea.Batch(
			LoadingStartCmd(),
			LogMessageCmd(fmt.Sprintf("Datatype updated: %s", msg.Label)),
			AllDatatypesFetchCmd(),
			FormCompletedCmd(&datatypesPage),
		)
	case DatatypeUpdateFailedMsg:
		return m, LogMessageCmd(fmt.Sprintf("Datatype update failed: %v", msg.Error))
	case CmsDefineDatatypeReadyMsg:
		return m, nil
	case BuildContentFormMsg:
		// Build dynamic form for content creation
		return m, m.BuildContentFieldsForm(msg.DatatypeID, msg.RouteID)
	case CreateContentFromDialogRequestMsg:
		// Create content from dialog values using authenticated user
		return m, m.HandleCreateContentFromDialog(msg, m.UserID)
	case FetchContentForEditMsg:
		// Fetch content fields for editing - this runs in background and shows edit dialog
		return m, m.HandleFetchContentForEdit(msg)
	case UpdateContentFromDialogRequestMsg:
		// Update content from dialog values using authenticated user
		return m, m.HandleUpdateContentFromDialog(msg, m.UserID)
	case MoveContentRequestMsg:
		return m, m.HandleMoveContent(msg)
	case ContentMovedMsg:
		if msg.AdminMode {
			return m, tea.Batch(
				LoadingStopCmd(),
				LogMessageCmd(fmt.Sprintf("Admin content moved: %s", msg.SourceContentID)),
				ReloadAdminContentTreeCmd(m.Config, types.AdminRouteID(msg.RouteID)),
			)
		}
		return m, tea.Batch(
			LoadingStopCmd(),
			ShowDialog("Success", "Content moved successfully", false),
			LogMessageCmd(fmt.Sprintf("Content moved: %s -> %s", msg.SourceContentID, msg.TargetContentID)),
			ReloadContentTreeCmd(m.Config, msg.RouteID),
		)
	case ReorderSiblingRequestMsg:
		return m, m.HandleReorderSibling(msg)
	case CopyContentRequestMsg:
		return m, m.HandleCopyContent(msg)
	case ContentCopiedMsg:
		if msg.AdminMode {
			return m, tea.Batch(
				LoadingStopCmd(),
				LogMessageCmd(fmt.Sprintf("Admin content copied: %s", msg.NewContentID)),
				ReloadAdminContentTreeCmd(m.Config, types.AdminRouteID(msg.RouteID)),
			)
		}
		return m, tea.Batch(
			LoadingStopCmd(),
			ShowDialog("Success", fmt.Sprintf("Content copied with %d fields", msg.FieldCount), false),
			ReloadContentTreeCmd(m.Config, msg.RouteID),
		)
	case TogglePublishRequestMsg:
		return m, m.HandleTogglePublish(msg)
	case ConfirmedPublishMsg:
		return m, m.HandleConfirmedPublish(msg)
	case ConfirmedUnpublishMsg:
		return m, m.HandleConfirmedUnpublish(msg)
	case PublishCompletedMsg:
		if msg.AdminMode {
			return m, tea.Batch(
				LoadingStopCmd(),
				LogMessageCmd(fmt.Sprintf("Admin content published: %s", msg.ContentID)),
				ReloadAdminContentTreeCmd(m.Config, types.AdminRouteID(msg.RouteID)),
			)
		}
		return m, tea.Batch(
			LoadingStopCmd(),
			ShowDialog("Published", "Content published via snapshot.", false),
			ReloadContentTreeCmd(m.Config, msg.RouteID),
		)
	case UnpublishCompletedMsg:
		if msg.AdminMode {
			return m, tea.Batch(
				LoadingStopCmd(),
				LogMessageCmd(fmt.Sprintf("Admin content unpublished: %s", msg.ContentID)),
				ReloadAdminContentTreeCmd(m.Config, types.AdminRouteID(msg.RouteID)),
			)
		}
		return m, tea.Batch(
			LoadingStopCmd(),
			ShowDialog("Unpublished", "Content is now draft.", false),
			ReloadContentTreeCmd(m.Config, msg.RouteID),
		)
	case ContentPublishToggledMsg:
		statusLabel := "Published"
		if msg.NewStatus == types.ContentStatusDraft {
			statusLabel = "Draft"
		}
		return m, tea.Batch(
			LoadingStopCmd(),
			ShowDialog("Status Changed", fmt.Sprintf("Content is now: %s", statusLabel), false),
			ReloadContentTreeCmd(m.Config, msg.RouteID),
		)
	case ListVersionsRequestMsg:
		return m, m.HandleListVersions(msg)
	case ConfirmedRestoreVersionMsg:
		return m, m.HandleConfirmedRestoreVersion(msg)
	case DeleteContentRequestMsg:
		// Delete content
		return m, m.HandleDeleteContent(msg)
	case ContentDeletedMsg:
		// Content deleted successfully - reload tree and show success
		if msg.AdminMode {
			newModel := m
			newModel.Cursor = 0
			return newModel, tea.Batch(
				LoadingStopCmd(),
				LogMessageCmd(fmt.Sprintf("Admin content deleted: %s", msg.ContentID)),
				ReloadAdminContentTreeCmd(m.Config, types.AdminRouteID(msg.RouteID)),
			)
		}
		return m, tea.Batch(
			LoadingStopCmd(),
			ShowDialog("Success", "Content deleted successfully", false),
			LogMessageCmd(fmt.Sprintf("Content deleted: ID=%s", msg.ContentID)),
			ReloadContentTreeCmd(m.Config, msg.RouteID),
			RootContentSummaryFetchCmd(),
		)
	case DeleteDatatypeRequestMsg:
		// Delete datatype and its junction records
		return m, m.HandleDeleteDatatype(msg)
	case DeleteFieldRequestMsg:
		// Delete field and its junction record
		return m, m.HandleDeleteField(msg)
	case DeleteRouteRequestMsg:
		// Delete route
		return m, m.HandleDeleteRoute(msg)
	case MediaUploadStartMsg:
		return m, m.HandleMediaUpload(msg)
	case MediaUploadProgressMsg:
		// Chain next progress read; the final message will be
		// MediaUploadedMsg or ActionResultMsg, not another progress.
		return m, waitForMsg(msg.ProgressCh)
	case MediaUploadedMsg:
		return m, tea.Batch(
			MediaFetchCmd(),
			ShowDialog("Upload Complete", fmt.Sprintf("'%s' uploaded successfully.", msg.Name), false),
		)
	case DeleteMediaRequestMsg:
		// Delete media item
		return m, m.HandleDeleteMedia(msg)
	case CreateUserFromDialogRequestMsg:
		// Create user from form dialog
		return m, m.HandleCreateUserFromDialog(msg)
	case UpdateUserFromDialogRequestMsg:
		// Update user from form dialog
		return m, m.HandleUpdateUserFromDialog(msg)
	case DeleteUserRequestMsg:
		// Delete user
		return m, m.HandleDeleteUser(msg)
	case CreateWebhookFromDialogRequestMsg:
		// Create webhook from form dialog
		return m, m.HandleCreateWebhook(msg)
	case UpdateWebhookFromDialogRequestMsg:
		// Update webhook from form dialog
		return m, m.HandleUpdateWebhook(msg)
	case DeleteWebhookRequestMsg:
		// Delete webhook
		return m, m.HandleDeleteWebhook(msg)
	case CreateTokenFromDialogRequestMsg:
		// Create token from form dialog
		return m, m.HandleCreateToken(msg)
	case DeleteTokenRequestMsg:
		// Delete token
		return m, m.HandleDeleteToken(msg)
	case DeleteSessionRequestMsg:
		// Revoke session
		return m, m.HandleDeleteSession(msg)
	case DeleteSshKeyRequestMsg:
		return m, m.HandleDeleteSshKey(msg)
	case CreateRoleFromDialogRequestMsg:
		return m, m.HandleCreateRole(msg)
	case UpdateRoleFromDialogRequestMsg:
		return m, m.HandleUpdateRole(msg)
	case DeleteRoleRequestMsg:
		return m, m.HandleDeleteRole(msg)
	case UnlinkOauthRequestMsg:
		return m, m.HandleUnlinkOauth(msg)
	case CreateMediaDimensionFromDialogRequestMsg:
		return m, m.HandleCreateMediaDimension(msg)
	case UpdateMediaDimensionFromDialogRequestMsg:
		return m, m.HandleUpdateMediaDimension(msg)
	case DeleteMediaDimensionRequestMsg:
		return m, m.HandleDeleteMediaDimension(msg)
	case ContentCreatedMsg:
		if msg.AdminMode {
			return m, tea.Batch(
				LoadingStopCmd(),
				LogMessageCmd(fmt.Sprintf("Admin content created: %s", msg.ContentID)),
				ReloadAdminContentTreeCmd(m.Config, types.AdminRouteID(msg.RouteID)),
			)
		}
		// Success path - reload tree and navigate back to content browser
		contentPage := m.PageMap[CONTENT]
		return m, tea.Batch(
			ShowDialog(
				"Success",
				fmt.Sprintf("Created content with %d fields", msg.FieldCount),
				false,
			),
			LogMessageCmd(fmt.Sprintf("ContentData created: ID=%s, RouteID=%s", msg.ContentID, msg.RouteID)),
			ReloadContentTreeCmd(m.Config, msg.RouteID),
			FormCompletedCmd(&contentPage), // Navigate back to content browser
		)

	case ContentCreatedWithErrorsMsg:
		// Partial success path - reload tree even with errors, navigate back
		contentPage := m.PageMap[CONTENT]
		return m, tea.Batch(
			ShowDialog(
				"Warning",
				fmt.Sprintf("Content created but %d/%d fields failed",
					len(msg.FailedFields),
					msg.CreatedFields+len(msg.FailedFields),
				),
				false,
			),
			LogMessageCmd(fmt.Sprintf("Failed field IDs: %v", msg.FailedFields)),
			ReloadContentTreeCmd(m.Config, msg.RouteID),
			FormCompletedCmd(&contentPage), // Navigate back to content browser
		)

	// Plugin management messages
	case PluginActionRequestMsg:
		return m, m.HandlePluginAction(msg)
	case PluginActionCompleteMsg:
		var title, verb string
		switch msg.Action {
		case PluginActionEnable:
			title, verb = "Plugin Enabled", "enabled"
		case PluginActionDisable:
			title, verb = "Plugin Disabled", "disabled"
		case PluginActionReload:
			title, verb = "Plugin Reloaded", "reloaded"
		}
		return m, tea.Batch(
			PluginsFetchCmd(),
			HomeDashboardFetchCmd(m.DB),
			func() tea.Msg {
				return ActionResultMsg{
					Title:   title,
					Message: fmt.Sprintf("Plugin '%s' has been %s.", msg.Name, verb),
				}
			},
		)
	case PluginRoutesApprovedMsg:
		return m, tea.Batch(
			PluginsFetchCmd(),
			HomeDashboardFetchCmd(m.DB),
			func() tea.Msg {
				return ActionResultMsg{
					Title:   "Routes Approved",
					Message: fmt.Sprintf("Approved %d routes for plugin '%s'.", msg.Count, msg.Name),
				}
			},
		)
	case PluginHooksApprovedMsg:
		return m, tea.Batch(
			PluginsFetchCmd(),
			HomeDashboardFetchCmd(m.DB),
			func() tea.Msg {
				return ActionResultMsg{
					Title:   "Hooks Approved",
					Message: fmt.Sprintf("Approved %d hooks for plugin '%s'.", msg.Count, msg.Name),
				}
			},
		)
	case PluginActionResultMsg:
		return m, func() tea.Msg {
			return ActionResultMsg{
				Title:   msg.Title,
				Message: msg.Message,
			}
		}
	// PluginSyncCapabilitiesRequestMsg: handler not yet implemented.
	// The screen emits this message but the sync logic is pending.

	default:
		return m, nil
	}
}

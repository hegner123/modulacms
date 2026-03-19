package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/hegner123/modulacms/internal/backup"
	"github.com/hegner123/modulacms/internal/db/types"
)

// =============================================================================
// DELETE DATATYPE
// =============================================================================

// DeleteDatatypeContext stores context for a datatype deletion operation.
type DeleteDatatypeContext struct {
	DatatypeID types.DatatypeID
	Label      string
}

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
// DELETE WEBHOOK
// =============================================================================

// DeleteWebhookContext stores context for a webhook deletion operation.
type DeleteWebhookContext struct {
	WebhookID types.WebhookID
	Name      string
}

// ShowDeleteWebhookDialogMsg triggers showing a delete webhook confirmation dialog.
type ShowDeleteWebhookDialogMsg struct {
	WebhookID types.WebhookID
	Name      string
}

// ShowDeleteWebhookDialogCmd creates a command to show a delete webhook confirmation dialog.
func ShowDeleteWebhookDialogCmd(webhookID types.WebhookID, name string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteWebhookDialogMsg{
			WebhookID: webhookID,
			Name:      name,
		}
	}
}

// DeleteWebhookRequestMsg triggers webhook deletion.
type DeleteWebhookRequestMsg struct {
	WebhookID types.WebhookID
}

// WebhookDeletedMsg is sent after a webhook is successfully deleted.
type WebhookDeletedMsg struct {
	WebhookID types.WebhookID
}

// DeleteWebhookCmd creates a command to delete a webhook.
func DeleteWebhookCmd(webhookID types.WebhookID) tea.Cmd {
	return func() tea.Msg {
		return DeleteWebhookRequestMsg{WebhookID: webhookID}
	}
}

// =============================================================================
// DELETE MEDIA DIMENSION
// =============================================================================

// DeleteMediaDimensionContext stores context for a media dimension deletion.
type DeleteMediaDimensionContext struct {
	MdID  string
	Label string
}

// ShowDeleteMediaDimensionDialogMsg triggers showing a delete media dimension dialog.
type ShowDeleteMediaDimensionDialogMsg struct {
	MdID  string
	Label string
}

// ShowDeleteMediaDimensionDialogCmd creates a command to show a delete media dimension dialog.
func ShowDeleteMediaDimensionDialogCmd(mdID string, label string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteMediaDimensionDialogMsg{MdID: mdID, Label: label}
	}
}

// DeleteMediaDimensionRequestMsg triggers media dimension deletion.
type DeleteMediaDimensionRequestMsg struct {
	MdID string
}

// MediaDimensionDeletedMsg is sent after a media dimension is deleted.
type MediaDimensionDeletedMsg struct {
	MdID string
}

// DeleteMediaDimensionCmd creates a command to delete a media dimension.
func DeleteMediaDimensionCmd(mdID string) tea.Cmd {
	return func() tea.Msg {
		return DeleteMediaDimensionRequestMsg{MdID: mdID}
	}
}

// =============================================================================
// DELETE TOKEN
// =============================================================================

// DeleteTokenContext stores context for a token deletion operation.
type DeleteTokenContext struct {
	TokenID string
	Label   string
}

// ShowDeleteTokenDialogMsg triggers showing a delete token confirmation dialog.
type ShowDeleteTokenDialogMsg struct {
	TokenID string
	Label   string
}

// ShowDeleteTokenDialogCmd creates a command to show a delete token confirmation dialog.
func ShowDeleteTokenDialogCmd(tokenID string, label string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteTokenDialogMsg{
			TokenID: tokenID,
			Label:   label,
		}
	}
}

// DeleteTokenRequestMsg triggers token deletion.
type DeleteTokenRequestMsg struct {
	TokenID string
}

// TokenDeletedMsg is sent after a token is successfully deleted.
type TokenDeletedMsg struct {
	TokenID string
}

// DeleteTokenCmd creates a command to delete a token.
func DeleteTokenCmd(tokenID string) tea.Cmd {
	return func() tea.Msg {
		return DeleteTokenRequestMsg{TokenID: tokenID}
	}
}

// =============================================================================
// DELETE SESSION
// =============================================================================

// DeleteSessionContext stores context for a session deletion operation.
type DeleteSessionContext struct {
	SessionID types.SessionID
	Label     string
}

// ShowDeleteSessionDialogMsg triggers showing a delete session confirmation dialog.
type ShowDeleteSessionDialogMsg struct {
	SessionID types.SessionID
	Label     string
}

// ShowDeleteSessionDialogCmd creates a command to show a delete session confirmation dialog.
func ShowDeleteSessionDialogCmd(sessionID types.SessionID, label string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteSessionDialogMsg{
			SessionID: sessionID,
			Label:     label,
		}
	}
}

// DeleteSessionRequestMsg triggers session deletion.
type DeleteSessionRequestMsg struct {
	SessionID types.SessionID
}

// SessionDeletedMsg is sent after a session is successfully deleted.
type SessionDeletedMsg struct {
	SessionID types.SessionID
}

// DeleteSessionCmd creates a command to delete a session.
func DeleteSessionCmd(sessionID types.SessionID) tea.Cmd {
	return func() tea.Msg {
		return DeleteSessionRequestMsg{SessionID: sessionID}
	}
}

// =============================================================================
// DIALOG ACCEPT DISPATCH
// =============================================================================

// handleDialogAccept processes DialogAcceptMsg by dispatching on the action type.
func (m Model) handleDialogAccept(msg DialogAcceptMsg) (Model, tea.Cmd) {
	switch msg.Action {
	case DIALOGQUITCONFIRM:
		// User confirmed quit
		return m, tea.Quit
	case DIALOGDELETECONTENT:
		// User confirmed content deletion (regular or admin)
		if ctx, ok := m.DCtx.Active.(*DeleteContentContext); ok {
			contentID := ctx.ContentID
			routeID := ctx.RouteID
			adminMode := ctx.AdminMode
			m.DCtx.Active = nil
			if adminMode {
				return m, tea.Batch(
					OverlayClearCmd(),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					func() tea.Msg {
						return ConfirmedDeleteAdminContentMsg{
							AdminContentID: types.AdminContentID(contentID),
							AdminRouteID:   types.AdminRouteID(routeID),
						}
					},
				)
			}
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				DeleteContentCmd(contentID, routeID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGPUBLISHCONTENT:
		// User confirmed publish (regular or admin)
		if ctx, ok := m.DCtx.Active.(*PublishContentContext); ok {
			m.DCtx.Active = nil
			if ctx.AdminMode {
				return m, tea.Batch(
					OverlayClearCmd(),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					func() tea.Msg {
						return ConfirmedPublishAdminContentMsg{
							AdminContentID: types.AdminContentID(ctx.ContentID),
							AdminRouteID:   types.AdminRouteID(ctx.RouteID),
						}
					},
				)
			}
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				func() tea.Msg {
					return ConfirmedPublishMsg{
						ContentID: ctx.ContentID,
						RouteID:   ctx.RouteID,
					}
				},
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGUNPUBLISHCONTENT:
		// User confirmed unpublish (regular or admin)
		if ctx, ok := m.DCtx.Active.(*PublishContentContext); ok {
			m.DCtx.Active = nil
			if ctx.AdminMode {
				return m, tea.Batch(
					OverlayClearCmd(),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					func() tea.Msg {
						return ConfirmedUnpublishAdminContentMsg{
							AdminContentID: types.AdminContentID(ctx.ContentID),
							AdminRouteID:   types.AdminRouteID(ctx.RouteID),
						}
					},
				)
			}
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				func() tea.Msg {
					return ConfirmedUnpublishMsg{
						ContentID: ctx.ContentID,
						RouteID:   ctx.RouteID,
					}
				},
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGRESTOREVERSION:
		// User confirmed version restore (regular or admin)
		if ctx, ok := m.DCtx.Active.(*RestoreVersionContext); ok {
			m.DCtx.Active = nil
			if ctx.AdminMode {
				return m, tea.Batch(
					OverlayClearCmd(),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					func() tea.Msg {
						return ConfirmedRestoreAdminVersionMsg{
							AdminContentID: types.AdminContentID(ctx.ContentID),
							VersionID:      types.AdminContentVersionID(ctx.VersionID),
							AdminRouteID:   types.AdminRouteID(ctx.RouteID),
						}
					},
				)
			}
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				func() tea.Msg {
					return ConfirmedRestoreVersionMsg{
						ContentID: ctx.ContentID,
						VersionID: ctx.VersionID,
						RouteID:   ctx.RouteID,
					}
				},
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGLOCALESELECT:
		// Extract the locale code from the selected locale option
		selectedIdx := 0
		if d, ok := m.ActiveOverlay.(*DialogModel); ok && d != nil {
			selectedIdx = d.ActionIndex
		}
		var selectedLocale string
		if d, ok := m.ActiveOverlay.(*DialogModel); ok && d != nil && selectedIdx >= 0 && selectedIdx < len(d.Locales) {
			// The locale string is "code - label", extract the code
			parts := strings.SplitN(d.Locales[selectedIdx], " - ", 2)
			if len(parts) > 0 {
				selectedLocale = parts[0]
			}
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			func() tea.Msg {
				return LocaleSwitchMsg{Locale: selectedLocale}
			},
		)
	case DIALOGDELETE:
		col, val := m.GetCurrentRowPK()
		return m, tea.Batch(
			DatabaseDeleteEntryCmd(col, val, m.TableState.Table),
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGACTIONCONFIRM:
		actionIndex := 0
		if d, ok := m.ActiveOverlay.(*DialogModel); ok && d != nil {
			actionIndex = d.ActionIndex
		}
		return m, tea.Batch(
			OverlayClearCmd(),
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
		if ctx, ok := m.DCtx.Active.(*InitializeRouteContentContext); ok {
			routeID := ctx.Route.RouteID
			datatypeID := ctx.DatatypeID
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				InitializeRouteContentCmd(routeID, datatypeID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETEDATATYPE:
		if ctx, ok := m.DCtx.Active.(*DeleteDatatypeContext); ok {
			datatypeID := ctx.DatatypeID
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				DeleteDatatypeCmd(datatypeID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETEFIELD:
		if ctx, ok := m.DCtx.Active.(*DeleteFieldContext); ok {
			fieldID := ctx.FieldID
			datatypeID := ctx.DatatypeID
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				DeleteFieldCmd(fieldID, datatypeID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETEROUTE:
		if ctx, ok := m.DCtx.Active.(*DeleteRouteContext); ok {
			routeID := ctx.RouteID
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				DeleteRouteCmd(routeID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETEMEDIA:
		if ctx, ok := m.DCtx.Active.(*DeleteMediaContext); ok {
			mediaID := ctx.MediaID
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				DeleteMediaCmd(mediaID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETEUSER:
		if ctx, ok := m.DCtx.Active.(*DeleteUserContext); ok {
			userID := ctx.UserID
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				DeleteUserCmd(userID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETEADMINROUTE:
		if ctx, ok := m.DCtx.Active.(*DeleteAdminRouteContext); ok {
			adminRouteID := ctx.AdminRouteID
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				DeleteAdminRouteCmd(adminRouteID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETEADMINDATATYPE:
		if ctx, ok := m.DCtx.Active.(*DeleteAdminDatatypeContext); ok {
			adminDatatypeID := ctx.AdminDatatypeID
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				DeleteAdminDatatypeCmd(adminDatatypeID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETEADMINFIELD:
		if ctx, ok := m.DCtx.Active.(*DeleteAdminFieldContext); ok {
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				DeleteAdminFieldCmd(ctx.AdminFieldID, ctx.AdminDatatypeID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETEFIELDTYPE:
		if ctx, ok := m.DCtx.Active.(*DeleteFieldTypeContext); ok {
			fieldTypeID := ctx.FieldTypeID
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				DeleteFieldTypeCmd(fieldTypeID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETEADMINFIELDTYPE:
		if ctx, ok := m.DCtx.Active.(*DeleteAdminFieldTypeContext); ok {
			adminFieldTypeID := ctx.AdminFieldTypeID
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				DeleteAdminFieldTypeCmd(adminFieldTypeID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETEVALIDATION:
		if ctx, ok := m.DCtx.Active.(*DeleteValidationContext); ok {
			validationID := ctx.ValidationID
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				DeleteValidationCmd(validationID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETEADMINVALIDATION:
		if ctx, ok := m.DCtx.Active.(*DeleteAdminValidationContext); ok {
			adminValidationID := ctx.AdminValidationID
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				DeleteAdminValidationCmd(adminValidationID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGBACKUPRESTORE:
		if ctx, ok := m.DCtx.Active.(*RestoreBackupContext); ok {
			backupPath := ctx.Path
			m.DCtx.Active = nil
			cfg := m.Config
			return m, tea.Batch(
				OverlayClearCmd(),
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
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGAPPROVEPLUGINROUTES:
		if ctx, ok := m.DCtx.Active.(*ApprovePluginContext); ok {
			pluginName := ctx.PluginName
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				func() tea.Msg {
					return PluginActionRequestMsg{Name: pluginName, Action: PluginActionApproveRoutes}
				},
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGAPPROVEPLUGINSHOOKS:
		if ctx, ok := m.DCtx.Active.(*ApprovePluginContext); ok {
			pluginName := ctx.PluginName
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				func() tea.Msg {
					return PluginActionRequestMsg{Name: pluginName, Action: PluginActionApproveHooks}
				},
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETECONTENTFIELD:
		if ctx, ok := m.DCtx.Active.(*DeleteContentFieldContext); ok {
			m.DCtx.Active = nil
			if ctx.AdminMode {
				return m, tea.Batch(
					OverlayClearCmd(),
					FocusSetCmd(PAGEFOCUS),
					LoadingStartCmd(),
					func() tea.Msg {
						return ConfirmedDeleteAdminContentFieldMsg{
							AdminContentFieldID: types.AdminContentFieldID(ctx.ContentFieldID),
							AdminContentID:      types.AdminContentID(ctx.ContentID),
							AdminRouteID:        types.AdminRouteID(ctx.RouteID),
							AdminDatatypeID:     types.NullableAdminDatatypeID{ID: types.AdminDatatypeID(ctx.DatatypeID.ID), Valid: ctx.DatatypeID.Valid},
						}
					},
				)
			}
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				m.HandleDeleteContentField(ctx.ContentFieldID, ctx.ContentID, ctx.RouteID, ctx.DatatypeID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGQUICKSTART:
		schemaIndex := 0
		if d, ok := m.ActiveOverlay.(*DialogModel); ok && d != nil {
			schemaIndex = d.ActionIndex
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			RunQuickstartInstallCmd(m.Config, m.UserID, schemaIndex),
		)
	case DIALOGDEPLOYPULL:
		if ctx, ok := m.DCtx.Active.(*DeployPullContext); ok {
			envName := ctx.EnvName
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				DeployPullCmd(envName, false),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDEPLOYPUSH:
		if ctx, ok := m.DCtx.Active.(*DeployPushContext); ok {
			envName := ctx.EnvName
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				DeployPushCmd(envName, false),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETEWEBHOOK:
		if ctx, ok := m.DCtx.Active.(*DeleteWebhookContext); ok {
			webhookID := ctx.WebhookID
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				DeleteWebhookCmd(webhookID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETEMEDIAFOLDER:
		if ctx, ok := m.DCtx.Active.(*DeleteMediaFolderContext); ok {
			folderID := ctx.FolderID
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				func() tea.Msg {
					return DeleteMediaFolderRequestMsg{FolderID: folderID}
				},
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETETOKEN:
		if ctx, ok := m.DCtx.Active.(*DeleteTokenContext); ok {
			tokenID := ctx.TokenID
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				DeleteTokenCmd(tokenID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETEMEDIADIMENSION:
		if ctx, ok := m.DCtx.Active.(*DeleteMediaDimensionContext); ok {
			mdID := ctx.MdID
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				DeleteMediaDimensionCmd(mdID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGDELETESESSION:
		if ctx, ok := m.DCtx.Active.(*DeleteSessionContext); ok {
			sessionID := ctx.SessionID
			m.DCtx.Active = nil
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				LoadingStartCmd(),
				DeleteSessionCmd(sessionID),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	case DIALOGPLUGINCONFIRM:
		m.DCtx.Active = nil
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			func() tea.Msg { return PluginDialogResponseMsg{Accepted: true} },
		)
	default:
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	}
}

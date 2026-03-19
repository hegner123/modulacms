package tui

import (
	"context"
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/utility"
)

// =============================================================================
// ADMIN CONTENT CREATE/UPDATE HANDLERS
// =============================================================================

// HandleCreateAdminContentFromDialog creates admin content from dialog values with tree support.
func (m Model) HandleCreateAdminContentFromDialog(
	datatypeID types.AdminDatatypeID,
	routeID types.AdminRouteID,
	parentID types.NullableAdminContentID,
	fieldValues map[types.AdminFieldID]string,
) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	locale := m.ActiveLocale
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Cannot create admin content: configuration not loaded"}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		contentData, err := d.CreateAdminContentData(ctx, ac, db.CreateAdminContentDataParams{
			AdminDatatypeID: types.NullableAdminDatatypeID{ID: datatypeID, Valid: true},
			AdminRouteID:    types.NullableAdminRouteID{ID: routeID, Valid: true},
			AuthorID:        userID,
			Status:          types.ContentStatusDraft,
			DateCreated:     types.TimestampNow(),
			DateModified:    types.TimestampNow(),
			ParentID:        parentID,
			FirstChildID:    types.NullableAdminContentID{},
			NextSiblingID:   types.NullableAdminContentID{},
			PrevSiblingID:   types.NullableAdminContentID{},
		})
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to create admin content: %v", err)}
		}

		dtFilter := types.NullableAdminDatatypeID{ID: datatypeID, Valid: true}
		allFields, fieldListErr := d.ListAdminFieldsByDatatypeID(dtFilter)
		if fieldListErr != nil {
			logger.Ferror("Failed to list admin datatype fields", fieldListErr)
		}

		createdFields := 0
		if allFields != nil {
			for _, field := range *allFields {
				value := fieldValues[field.AdminFieldID]
				_, fieldErr := d.CreateAdminContentField(ctx, ac, db.CreateAdminContentFieldParams{
					AdminContentDataID: types.NullableAdminContentID{ID: contentData.AdminContentDataID, Valid: true},
					AdminFieldID:       types.NullableAdminFieldID{ID: field.AdminFieldID, Valid: true},
					AdminFieldValue:    value,
					AdminRouteID:       types.NullableAdminRouteID{ID: routeID, Valid: true},
					Locale:             locale,
					AuthorID:           userID,
					DateCreated:        types.TimestampNow(),
					DateModified:       types.TimestampNow(),
				})
				if fieldErr != nil {
					logger.Ferror(fmt.Sprintf("Failed to create admin content field: %v", fieldErr), fieldErr)
				} else {
					createdFields++
				}
			}
		}

		logger.Finfo(fmt.Sprintf("Admin content created: %s with %d fields", contentData.AdminContentDataID, createdFields))
		return ContentCreatedMsg{
			ContentID: types.ContentID(contentData.AdminContentDataID),
			RouteID:   types.RouteID(routeID),
			AdminMode: true,
		}
	}
}

// HandleUpdateAdminContentFromDialog updates existing admin content fields from dialog values.
func (m Model) HandleUpdateAdminContentFromDialog(
	adminContentID types.AdminContentID,
	adminRouteID types.AdminRouteID,
	fieldValues map[types.AdminFieldID]string,
) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	locale := m.ActiveLocale
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Cannot update admin content: configuration not loaded"}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		contentFilter := types.NullableAdminContentID{ID: adminContentID, Valid: true}
		existingFields, err := d.ListAdminContentFieldsByContentDataAndLocale(contentFilter, locale)
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to fetch existing fields: %v", err)}
		}

		existingMap := make(map[types.AdminFieldID]db.AdminContentFields)
		if existingFields != nil {
			for _, cf := range *existingFields {
				if cf.AdminFieldID.Valid {
					existingMap[cf.AdminFieldID.ID] = cf
				}
			}
		}

		updatedCount := 0
		var updateErrors []string

		for fieldID, value := range fieldValues {
			if existing, ok := existingMap[fieldID]; ok {
				_, updateErr := d.UpdateAdminContentField(ctx, ac, db.UpdateAdminContentFieldParams{
					AdminContentFieldID: existing.AdminContentFieldID,
					AdminRouteID:        existing.AdminRouteID,
					AdminContentDataID:  contentFilter,
					AdminFieldID:        types.NullableAdminFieldID{ID: fieldID, Valid: true},
					AdminFieldValue:     value,
					Locale:              locale,
					AuthorID:            userID,
					DateCreated:         existing.DateCreated,
					DateModified:        types.TimestampNow(),
				})
				if updateErr != nil {
					logger.Ferror(fmt.Sprintf("Failed to update admin field %s", fieldID), updateErr)
					updateErrors = append(updateErrors, string(fieldID))
				} else {
					updatedCount++
				}
			} else {
				_, createErr := d.CreateAdminContentField(ctx, ac, db.CreateAdminContentFieldParams{
					AdminContentDataID: contentFilter,
					AdminFieldID:       types.NullableAdminFieldID{ID: fieldID, Valid: true},
					AdminFieldValue:    value,
					AdminRouteID:       types.NullableAdminRouteID{ID: adminRouteID, Valid: true},
					Locale:             locale,
					AuthorID:           userID,
					DateCreated:        types.TimestampNow(),
					DateModified:       types.TimestampNow(),
				})
				if createErr != nil {
					logger.Ferror(fmt.Sprintf("Failed to create admin field %s", fieldID), createErr)
					updateErrors = append(updateErrors, string(fieldID))
				} else {
					updatedCount++
				}
			}
		}

		if len(updateErrors) > 0 {
			return ActionResultMsg{
				Title:   "Partial Update",
				Message: fmt.Sprintf("Updated %d fields, but %d failed", updatedCount, len(updateErrors)),
			}
		}

		return AdminContentUpdatedFromDialogMsg{
			AdminContentID: adminContentID,
			AdminRouteID:   adminRouteID,
		}
	}
}

// =============================================================================
// ADMIN CONTENT DELETE HANDLER (tree-aware)
// =============================================================================

// HandleDeleteAdminContent processes tree-aware admin content deletion.
func (m Model) HandleDeleteAdminContent(msg ConfirmedDeleteAdminContentMsg) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Cannot delete admin content: configuration not loaded"}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		ops := newAdminTreeOps(d)
		node, err := ops.getNode(string(msg.AdminContentID))
		if err != nil || node == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Admin content not found: %v", err)}
		}

		if !node.firstChildID.isEmpty() {
			return ActionResultMsg{Title: "Cannot Delete", Message: "This content has children. Delete child nodes first."}
		}

		// Update sibling pointers before deletion
		if treeErrs := detachFromSiblings(ctx, ac, ops, node); len(treeErrs) > 0 {
			return ActionResultMsg{
				Title:   "Partial Failure",
				Message: fmt.Sprintf("Tree operation incomplete: %s", errors.Join(treeErrs...)),
				IsError: true,
			}
		}

		if deleteErr := d.DeleteAdminContentData(ctx, ac, msg.AdminContentID); deleteErr != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to delete admin content: %v", deleteErr)}
		}

		logger.Finfo(fmt.Sprintf("Admin content deleted: %s", msg.AdminContentID))
		return ContentDeletedMsg{ContentID: types.ContentID(msg.AdminContentID), RouteID: types.RouteID(msg.AdminRouteID), AdminMode: true}
	}
}

// =============================================================================
// ADMIN CONTENT MOVE HANDLER
// =============================================================================

// HandleMoveAdminContent detaches source from current position and attaches as last child of target.
func (m Model) HandleMoveAdminContent(msg AdminMoveContentRequestMsg) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Cannot move admin content: configuration not loaded"}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		ops := newAdminTreeOps(d)

		sourceNode, err := ops.getNode(string(msg.SourceID))
		if err != nil || sourceNode == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Source admin content not found: %v", err)}
		}

		targetNode, err := ops.getNode(string(msg.TargetID))
		if err != nil || targetNode == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Target admin content not found: %v", err)}
		}

		// STEP 1: Detach source from old position
		if treeErrs := detachFromSiblings(ctx, ac, ops, sourceNode); len(treeErrs) > 0 {
			return ActionResultMsg{
				Title:   "Partial Failure",
				Message: fmt.Sprintf("Tree operation incomplete: %s", errors.Join(treeErrs...)),
				IsError: true,
			}
		}

		// STEP 2: Attach source as last child of target
		newPrevID, attachErr := attachAsLastChild(ctx, ac, ops, sourceNode.id, string(msg.TargetID))
		if attachErr != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to attach content: %v", attachErr)}
		}

		// STEP 3: Update source with new parent
		source := sourceNode.source.(*db.AdminContentData)
		sourceParams := db.UpdateAdminContentDataParams{
			AdminContentDataID: source.AdminContentDataID,
			AdminRouteID:       source.AdminRouteID,
			AdminDatatypeID:    source.AdminDatatypeID,
			ParentID:           types.NullableAdminContentID{ID: msg.TargetID, Valid: true},
			FirstChildID:       source.FirstChildID,
			NextSiblingID:      types.NullableAdminContentID{},
			PrevSiblingID:      types.NullableAdminContentID{ID: types.AdminContentID(newPrevID.ID), Valid: newPrevID.Valid},
			AuthorID:           source.AuthorID,
			Status:             source.Status,
			DateCreated:        source.DateCreated,
			DateModified:       types.TimestampNow(),
		}
		if _, updateErr := d.UpdateAdminContentData(ctx, ac, sourceParams); updateErr != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to update source: %v", updateErr)}
		}

		logger.Finfo(fmt.Sprintf("Admin content moved: %s -> %s", msg.SourceID, msg.TargetID))
		return ContentMovedMsg{SourceContentID: types.ContentID(msg.SourceID), RouteID: types.RouteID(msg.AdminRouteID), AdminMode: true}
	}
}

// =============================================================================
// ADMIN CONTENT REORDER HANDLER
// =============================================================================

// HandleAdminReorderSibling swaps an admin node with its prev or next sibling.
func (m Model) HandleAdminReorderSibling(msg AdminReorderSiblingRequestMsg) tea.Cmd {
	cfg := m.Config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Configuration not loaded"}
		}
	}

	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)
		logger := utility.DefaultLogger

		ops := newAdminTreeOps(d)

		a, err := ops.getNode(string(msg.AdminContentID))
		if err != nil || a == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Admin content not found: %v", err)}
		}

		if msg.Direction == "up" {
			if a.prevSiblingID.isEmpty() {
				return ActionResultMsg{Title: "Info", Message: "Already at top"}
			}
			b, bErr := ops.getNode(a.prevSiblingID.ID)
			if bErr != nil || b == nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Previous sibling not found: %v", bErr)}
			}

			if treeErrs := swapSiblings(ctx, ac, ops, a, b, "up"); len(treeErrs) > 0 {
				return ActionResultMsg{
					Title:   "Partial Failure",
					Message: fmt.Sprintf("Tree operation incomplete: %s", errors.Join(treeErrs...)),
					IsError: true,
				}
			}

		} else {
			// Move down: swap A with its next sibling B
			if a.nextSiblingID.isEmpty() {
				return ActionResultMsg{Title: "Info", Message: "Already at bottom"}
			}
			b, bErr := ops.getNode(a.nextSiblingID.ID)
			if bErr != nil || b == nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Next sibling not found: %v", bErr)}
			}

			if treeErrs := swapSiblings(ctx, ac, ops, a, b, "down"); len(treeErrs) > 0 {
				return ActionResultMsg{
					Title:   "Partial Failure",
					Message: fmt.Sprintf("Tree operation incomplete: %s", errors.Join(treeErrs...)),
					IsError: true,
				}
			}
		}

		logger.Finfo(fmt.Sprintf("Admin content reordered %s: %s", msg.Direction, msg.AdminContentID))
		return ContentReorderedMsg{
			ContentID: types.ContentID(msg.AdminContentID),
			RouteID:   types.RouteID(msg.AdminRouteID),
			Direction: msg.Direction,
			AdminMode: true,
		}
	}
}

// =============================================================================
// ADMIN CONTENT COPY HANDLER
// =============================================================================

// HandleCopyAdminContent duplicates an admin content node and its fields as a new sibling.
func (m Model) HandleCopyAdminContent(msg AdminCopyContentRequestMsg) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	locale := m.ActiveLocale

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Configuration not loaded"}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)
		logger := utility.DefaultLogger

		source, err := d.GetAdminContentData(msg.SourceID)
		if err != nil || source == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Source admin content not found: %v", err)}
		}

		// Fetch source fields
		sourceFields, sfErr := d.ListAdminContentFieldsByContentDataIDs(ctx, []types.AdminContentID{msg.SourceID}, locale)
		if sfErr != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to read source fields: %v", sfErr)}
		}

		now := types.TimestampNow()
		newContent, createErr := d.CreateAdminContentData(ctx, ac, db.CreateAdminContentDataParams{
			AdminRouteID:    source.AdminRouteID,
			ParentID:        source.ParentID,
			FirstChildID:    types.NullableAdminContentID{},
			NextSiblingID:   source.NextSiblingID,
			PrevSiblingID:   types.NullableAdminContentID{ID: source.AdminContentDataID, Valid: true},
			AdminDatatypeID: source.AdminDatatypeID,
			AuthorID:        userID,
			Status:          types.ContentStatusDraft,
			DateCreated:     now,
			DateModified:    now,
		})
		if createErr != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to create admin content copy: %v", createErr)}
		}

		// Update sibling pointers: source.next -> new, old-next.prev -> new
		ops := newAdminTreeOps(d)
		sourceNode := adminContentToTreeNode(source)
		if spliceErr := spliceAfter(ctx, ac, ops, sourceNode, string(newContent.AdminContentDataID)); spliceErr != nil {
			logger.Ferror("Failed to update sibling pointers after admin copy", spliceErr)
		}

		// Copy fields
		fieldCount := 0
		if sourceFields != nil {
			for _, sf := range *sourceFields {
				_, fieldErr := d.CreateAdminContentField(ctx, ac, db.CreateAdminContentFieldParams{
					AdminRouteID:       sf.AdminRouteID,
					AdminContentDataID: types.NullableAdminContentID{ID: newContent.AdminContentDataID, Valid: true},
					AdminFieldID:       sf.AdminFieldID,
					AdminFieldValue:    sf.AdminFieldValue,
					Locale:             locale,
					AuthorID:           userID,
					DateCreated:        now,
					DateModified:       now,
				})
				if fieldErr != nil {
					logger.Ferror(fmt.Sprintf("Failed to copy admin field: %v", fieldErr), fieldErr)
				} else {
					fieldCount++
				}
			}
		}

		logger.Finfo(fmt.Sprintf("Admin content copied: %s -> %s with %d fields", msg.SourceID, newContent.AdminContentDataID, fieldCount))
		return ContentCopiedMsg{NewContentID: types.ContentID(newContent.AdminContentDataID), RouteID: types.RouteID(msg.AdminRouteID), AdminMode: true}
	}
}

// =============================================================================
// ADMIN PUBLISHING HANDLERS
// =============================================================================

// HandleAdminConfirmedPublish publishes admin content via snapshot.
func (m Model) HandleAdminConfirmedPublish(msg ConfirmedPublishAdminContentMsg) tea.Cmd {
	cfg := m.Config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Configuration not loaded"}
		}
	}

	userID := m.UserID
	locale := m.ActiveLocale
	dispatcher := m.Dispatcher
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)
		logger := utility.DefaultLogger

		retentionCap := cfg.VersionMaxPerContent()
		publishAll := !cfg.Node_Level_Publish
		pubErr := publishing.PublishAdminContent(ctx, d, msg.AdminContentID, locale, userID, ac, retentionCap, publishAll, dispatcher)
		if pubErr != nil {
			logger.Ferror(fmt.Sprintf("Failed to publish admin content %s", msg.AdminContentID), pubErr)
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Publish failed: %v", pubErr)}
		}

		logger.Finfo(fmt.Sprintf("Admin content %s published", msg.AdminContentID))
		return PublishCompletedMsg{ContentID: types.ContentID(msg.AdminContentID), RouteID: types.RouteID(msg.AdminRouteID), AdminMode: true}
	}
}

// HandleAdminConfirmedUnpublish unpublishes admin content.
func (m Model) HandleAdminConfirmedUnpublish(msg ConfirmedUnpublishAdminContentMsg) tea.Cmd {
	cfg := m.Config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Configuration not loaded"}
		}
	}

	userID := m.UserID
	locale := m.ActiveLocale
	dispatcher := m.Dispatcher
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)
		logger := utility.DefaultLogger

		unpubErr := publishing.UnpublishAdminContent(ctx, d, msg.AdminContentID, locale, userID, ac, dispatcher)
		if unpubErr != nil {
			logger.Ferror(fmt.Sprintf("Failed to unpublish admin content %s", msg.AdminContentID), unpubErr)
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Unpublish failed: %v", unpubErr)}
		}

		logger.Finfo(fmt.Sprintf("Admin content %s unpublished", msg.AdminContentID))
		return UnpublishCompletedMsg{ContentID: types.ContentID(msg.AdminContentID), RouteID: types.RouteID(msg.AdminRouteID), AdminMode: true}
	}
}

// =============================================================================
// ADMIN VERSIONING HANDLERS
// =============================================================================

// HandleAdminListVersions fetches versions for an admin content item.
func (m Model) HandleAdminListVersions(msg AdminListVersionsRequestMsg) tea.Cmd {
	cfg := m.Config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Configuration not loaded"}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		versions, err := d.ListAdminContentVersionsByContent(msg.AdminContentID)
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to list admin versions: %v", err)}
		}
		var versionList []db.AdminContentVersion
		if versions != nil {
			versionList = *versions
		}
		return AdminVersionsListedMsg{
			Versions:       versionList,
			AdminContentID: msg.AdminContentID,
			AdminRouteID:   msg.AdminRouteID,
		}
	}
}

// HandleAdminConfirmedRestoreVersion restores admin content from a saved version.
func (m Model) HandleAdminConfirmedRestoreVersion(msg ConfirmedRestoreAdminVersionMsg) tea.Cmd {
	cfg := m.Config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Configuration not loaded"}
		}
	}

	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)
		logger := utility.DefaultLogger

		result, err := publishing.RestoreAdminContent(ctx, d, msg.AdminContentID, msg.VersionID, userID, ac)
		if err != nil {
			logger.Ferror(fmt.Sprintf("Failed to restore admin version for %s", msg.AdminContentID), err)
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Restore failed: %v", err)}
		}

		logger.Finfo(fmt.Sprintf("Admin content %s restored from version %s (%d fields)", msg.AdminContentID, msg.VersionID, result.FieldsRestored))
		return VersionRestoredMsg{
			ContentID:      types.ContentID(msg.AdminContentID),
			RouteID:        types.RouteID(msg.AdminRouteID),
			FieldsRestored: result.FieldsRestored,
			AdminMode:      true,
		}
	}
}

// =============================================================================
// ADMIN CONTENT FIELD HANDLERS
// =============================================================================

// HandleDeleteAdminContentField deletes an admin content field record.
func (m Model) HandleDeleteAdminContentField(msg ConfirmedDeleteAdminContentFieldMsg) tea.Cmd {
	cfg := m.Config
	userID := m.UserID

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Configuration not loaded"}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		err := d.DeleteAdminContentField(ctx, ac, msg.AdminContentFieldID)
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to delete admin field: %v", err)}
		}

		return ContentFieldDeletedMsg{
			ContentID: types.ContentID(msg.AdminContentID),
			RouteID:   types.RouteID(msg.AdminRouteID),
			AdminMode: true,
		}
	}
}

// HandleAddAdminContentField creates a new admin content field record.
func (m Model) HandleAddAdminContentField(
	adminContentID types.AdminContentID,
	adminFieldID types.AdminFieldID,
	adminRouteID types.AdminRouteID,
) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	locale := m.ActiveLocale

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Configuration not loaded"}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		_, err := d.CreateAdminContentField(ctx, ac, db.CreateAdminContentFieldParams{
			AdminContentDataID: types.NullableAdminContentID{ID: adminContentID, Valid: true},
			AdminFieldID:       types.NullableAdminFieldID{ID: adminFieldID, Valid: true},
			AdminFieldValue:    "",
			AdminRouteID:       types.NullableAdminRouteID{ID: adminRouteID, Valid: true},
			Locale:             locale,
			AuthorID:           userID,
			DateCreated:        types.TimestampNow(),
			DateModified:       types.TimestampNow(),
		})
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to add admin content field: %v", err)}
		}

		return ContentFieldAddedMsg{
			ContentID: types.ContentID(adminContentID),
			RouteID:   types.RouteID(adminRouteID),
			AdminMode: true,
		}
	}
}

// HandleEditAdminSingleField updates a single admin content field value.
func (m Model) HandleEditAdminSingleField(
	adminContentFieldID types.AdminContentFieldID,
	adminContentID types.AdminContentID,
	adminFieldID types.AdminFieldID,
	newValue string,
	adminRouteID types.AdminRouteID,
	adminDatatypeID types.NullableAdminDatatypeID,
) tea.Cmd {
	cfg := m.Config
	userID := m.UserID

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Configuration not loaded"}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		cf, err := d.GetAdminContentField(adminContentFieldID)
		if err != nil || cf == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Admin content field not found: %v", err)}
		}

		_, err = d.UpdateAdminContentField(ctx, ac, db.UpdateAdminContentFieldParams{
			AdminContentFieldID: adminContentFieldID,
			AdminRouteID:        cf.AdminRouteID,
			AdminContentDataID:  types.NullableAdminContentID{ID: adminContentID, Valid: true},
			AdminFieldID:        types.NullableAdminFieldID{ID: adminFieldID, Valid: true},
			AdminFieldValue:     newValue,
			Locale:              cf.Locale,
			AuthorID:            userID,
			DateCreated:         cf.DateCreated,
			DateModified:        types.TimestampNow(),
		})
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to update admin field: %v", err)}
		}

		return ContentFieldUpdatedMsg{
			ContentID: types.ContentID(adminContentID),
			RouteID:   types.RouteID(adminRouteID),
			AdminMode: true,
		}
	}
}

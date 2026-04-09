package tui

import (
	"context"
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/tree"
	"github.com/hegner123/modulacms/internal/utility"
)

// HandleDeleteContent deletes content and updates tree structure
func (m Model) HandleDeleteContent(msg DeleteContentRequestMsg) tea.Cmd {
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}
	cfg := m.Config
	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		contentID := types.ContentID(msg.ContentID)
		logger.Finfo(fmt.Sprintf("Deleting content: %s", contentID))

		// Get the content data to check structure
		ops := newContentTreeOps(d)
		node, err := ops.getNode(string(contentID))
		if err != nil || node == nil {
			logger.Ferror("failed to get content for deletion", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Content not found: %v", err),
			}
		}

		// Check if it has children (should have been prevented by UI, but double-check)
		if !node.firstChildID.isEmpty() {
			return ActionResultMsg{
				Title:   "Cannot Delete",
				Message: "This content has children. Delete child nodes first.",
			}
		}

		// Update sibling pointers before deletion
		if treeErrs := detachFromSiblings(ctx, ac, ops, node); len(treeErrs) > 0 {
			return ActionResultMsg{
				Title:   "Partial Failure",
				Message: fmt.Sprintf("Tree operation incomplete: %s", errors.Join(treeErrs...)),
				IsError: true,
			}
		}

		// Delete the content data (content_fields will cascade delete)
		if err := d.DeleteContentData(ctx, ac, contentID); err != nil {
			logger.Ferror("failed to delete content", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("failed to delete content: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Content deleted successfully: %s", contentID))
		return ContentDeletedMsg{
			ContentID: types.ContentID(msg.ContentID),
			RouteID:   types.RouteID(msg.RouteID),
		}
	}
}

// HandleMoveContent detaches source from its current position and attaches it
// as the last child of the target node. All affected sibling/parent pointers
// are updated in the database.
func (m Model) HandleMoveContent(msg MoveContentRequestMsg) tea.Cmd {
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}
	cfg := m.Config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot move content: configuration not loaded",
			}
		}
	}

	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		logger.Finfo(fmt.Sprintf("Moving content: %s -> %s", msg.SourceContentID, msg.TargetContentID))

		ops := newContentTreeOps(d)

		// Read source node
		sourceNode, err := ops.getNode(string(msg.SourceContentID))
		if err != nil || sourceNode == nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Source content not found: %v", err),
			}
		}

		// Verify target exists
		targetNode, err := ops.getNode(string(msg.TargetContentID))
		if err != nil || targetNode == nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Target content not found: %v", err),
			}
		}

		// === STEP 1: Detach source from old position ===
		if treeErrs := detachFromSiblings(ctx, ac, ops, sourceNode); len(treeErrs) > 0 {
			return ActionResultMsg{
				Title:   "Partial Failure",
				Message: fmt.Sprintf("Tree detach incomplete: %s", errors.Join(treeErrs...)),
				IsError: true,
			}
		}

		// === STEP 2: Attach source as last child of target ===
		newPrevID, attachErr := attachAsLastChild(ctx, ac, ops, sourceNode.id, string(msg.TargetContentID))
		if attachErr != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("failed to attach content: %v", attachErr),
			}
		}

		// === STEP 3: Update source node with new parent and cleared sibling pointers ===
		source := sourceNode.source.(*db.ContentData)
		sourceParams := db.UpdateContentDataParams{
			ContentDataID: source.ContentDataID,
			RouteID:       source.RouteID,
			ParentID:      types.NullableContentID{ID: msg.TargetContentID, Valid: true},
			FirstChildID:  source.FirstChildID,       // preserve children
			NextSiblingID: types.NullableContentID{}, // last child, no next
			PrevSiblingID: types.NullableContentID{ID: types.ContentID(newPrevID.ID), Valid: newPrevID.Valid},
			DatatypeID:    source.DatatypeID,
			AuthorID:      source.AuthorID,
			Status:        source.Status,
			DateCreated:   source.DateCreated,
			DateModified:  types.TimestampNow(),
		}
		if _, updateErr := d.UpdateContentData(ctx, ac, sourceParams); updateErr != nil {
			logger.Ferror("failed to update source node", updateErr)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("failed to update source: %v", updateErr),
			}
		}

		logger.Finfo(fmt.Sprintf("Content moved successfully: %s -> %s", msg.SourceContentID, msg.TargetContentID))
		return ContentMovedMsg{
			SourceContentID: msg.SourceContentID,
			TargetContentID: msg.TargetContentID,
			RouteID:         msg.RouteID,
		}
	}
}

// ReloadContentTree fetches tree data from database and loads it into the Root
func (m Model) ReloadContentTree(c *config.Config, routeID types.RouteID) tea.Cmd {
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}
	return func() tea.Msg {
		d := db.ConfigDB(*c)

		// Fetch tree data from database
		nullableRouteID := types.NullableRouteID{ID: routeID, Valid: !routeID.IsZero()}
		rows, err := d.GetContentTreeByRoute(nullableRouteID)
		if err != nil {
			logger.Ferror(fmt.Sprintf("GetContentTreeByRoute error for route %s", routeID), err)
			return FetchErrMsg{Error: fmt.Errorf("failed to fetch content tree: %w", err)}
		}

		if rows == nil {
			logger.Finfo(fmt.Sprintf("GetContentTreeByRoute returned nil rows for route %s", routeID))
			return TreeLoadedMsg{
				RouteID:  routeID,
				Stats:    &tree.LoadStats{},
				RootNode: nil,
			}
		}

		logger.Finfo(fmt.Sprintf("GetContentTreeByRoute returned %d rows for route %s", len(*rows), routeID))

		if len(*rows) == 0 {
			logger.Finfo(fmt.Sprintf("no rows returned for route %s", routeID))
			return TreeLoadedMsg{
				RouteID:  routeID,
				Stats:    &tree.LoadStats{},
				RootNode: nil,
			}
		}

		// Create new tree root and load from rows
		newRoot := tree.NewRoot()
		stats, err := newRoot.LoadFromRows(rows)
		if err != nil {
			return FetchErrMsg{Error: fmt.Errorf("failed to load tree from rows: %w", err)}
		}

		return TreeLoadedMsg{
			RouteID:  routeID,
			Stats:    stats,
			RootNode: newRoot,
		}
	}
}

// ReloadContentTreeByRootID loads a content tree by root_id for standalone/global content.
func (m Model) ReloadContentTreeByRootID(c *config.Config, rootID types.ContentID) tea.Cmd {
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}
	return func() tea.Msg {
		d := db.ConfigDB(*c)

		nullableRootID := types.NullableContentID{ID: rootID, Valid: !rootID.IsZero()}
		rows, err := d.GetContentTreeByRootID(nullableRootID)
		if err != nil {
			logger.Ferror(fmt.Sprintf("GetContentTreeByRootID error for root %s", rootID), err)
			return FetchErrMsg{Error: fmt.Errorf("failed to fetch content tree by root_id: %w", err)}
		}

		if rows == nil || len(*rows) == 0 {
			logger.Finfo(fmt.Sprintf("GetContentTreeByRootID returned no rows for root %s", rootID))
			return TreeLoadedMsg{
				Stats:    &tree.LoadStats{},
				RootNode: nil,
			}
		}

		logger.Finfo(fmt.Sprintf("GetContentTreeByRootID returned %d rows for root %s", len(*rows), rootID))

		newRoot := tree.NewRoot()
		stats, loadErr := newRoot.LoadFromRows(rows)
		if loadErr != nil {
			return FetchErrMsg{Error: fmt.Errorf("failed to load tree from rows: %w", loadErr)}
		}

		return TreeLoadedMsg{
			Stats:    stats,
			RootNode: newRoot,
		}
	}
}

// ReorderSiblingCmd creates a command to reorder content among siblings.
func ReorderSiblingCmd(contentID types.ContentID, routeID types.RouteID, direction string) tea.Cmd {
	return func() tea.Msg {
		return ReorderSiblingRequestMsg{
			ContentID: contentID,
			RouteID:   routeID,
			Direction: direction,
		}
	}
}

// HandleReorderSibling swaps a node with its prev or next sibling in the linked list.
func (m Model) HandleReorderSibling(msg ReorderSiblingRequestMsg) tea.Cmd {
	cfg := m.Config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "configuration not loaded"}
		}
	}

	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)
		logger := utility.DefaultLogger

		ops := newContentTreeOps(d)

		// Read the node to move
		a, err := ops.getNode(string(msg.ContentID))
		if err != nil || a == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Content not found: %v", err)}
		}

		if msg.Direction == "up" {
			// Move up: swap A with its prev sibling B
			if a.prevSiblingID.isEmpty() {
				return ActionResultMsg{Title: "Info", Message: "already at top"}
			}
			b, bErr := ops.getNode(a.prevSiblingID.ID)
			if bErr != nil || b == nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Previous sibling not found: %v", bErr)}
			}

			if treeErrs := swapSiblings(ctx, ac, ops, a, b, "up"); len(treeErrs) > 0 {
				return ActionResultMsg{
					Title:   "Partial Failure",
					Message: fmt.Sprintf("Reorder incomplete: %s", errors.Join(treeErrs...)),
					IsError: true,
				}
			}

		} else {
			// Move down: swap A with its next sibling B
			if a.nextSiblingID.isEmpty() {
				return ActionResultMsg{Title: "Info", Message: "already at bottom"}
			}
			b, bErr := ops.getNode(a.nextSiblingID.ID)
			if bErr != nil || b == nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Next sibling not found: %v", bErr)}
			}

			if treeErrs := swapSiblings(ctx, ac, ops, a, b, "down"); len(treeErrs) > 0 {
				return ActionResultMsg{
					Title:   "Partial Failure",
					Message: fmt.Sprintf("Reorder incomplete: %s", errors.Join(treeErrs...)),
					IsError: true,
				}
			}
		}

		logger.Finfo(fmt.Sprintf("Content reordered %s: %s", msg.Direction, msg.ContentID))
		return ContentReorderedMsg{
			ContentID: msg.ContentID,
			RouteID:   msg.RouteID,
			Direction: msg.Direction,
		}
	}
}

// CopyContentCmd creates a command to copy a content node as a new sibling.
func CopyContentCmd(sourceID types.ContentID, routeID types.RouteID) tea.Cmd {
	return func() tea.Msg {
		return CopyContentRequestMsg{
			SourceContentID: sourceID,
			RouteID:         routeID,
		}
	}
}

// HandleCopyContent duplicates a content node and its fields as a new sibling.
func (m Model) HandleCopyContent(msg CopyContentRequestMsg) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "configuration not loaded"}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)
		logger := utility.DefaultLogger

		// Read source
		source, err := d.GetContentData(msg.SourceContentID)
		if err != nil || source == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Source content not found: %v", err)}
		}

		// Read source fields
		sourceFields, err := d.ListContentFieldsByContentData(types.NullableContentID{ID: msg.SourceContentID, Valid: true})
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("failed to read source fields: %v", err)}
		}

		// Create new ContentData as sibling after source
		now := types.TimestampNow()
		newContent, createErr := d.CreateContentData(ctx, ac, db.CreateContentDataParams{
			RouteID:       source.RouteID,
			RootID:        source.RootID,
			ParentID:      source.ParentID,
			FirstChildID:  types.NullableContentID{},                                      // no children (flat copy)
			NextSiblingID: source.NextSiblingID,                                           // take source's next
			PrevSiblingID: types.NullableContentID{ID: source.ContentDataID, Valid: true}, // prev = source
			DatatypeID:    source.DatatypeID,
			AuthorID:      userID,
			Status:        types.ContentStatusDraft,
			DateCreated:   now,
			DateModified:  now,
		})
		if createErr != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("failed to create content copy: %v", createErr)}
		}

		if newContent.ContentDataID.IsZero() {
			return ActionResultMsg{Title: "Error", Message: "failed to create content copy"}
		}

		// Update sibling pointers: source.next -> new, old-next.prev -> new
		ops := newContentTreeOps(d)
		sourceNode := contentToTreeNode(source)
		if spliceErr := spliceAfter(ctx, ac, ops, sourceNode, string(newContent.ContentDataID)); spliceErr != nil {
			logger.Ferror("failed to update sibling pointers after copy", spliceErr)
		}

		// Copy fields — iterate the canonical datatype field list so all fields
		// get a content_field row, even if the source was created via the old
		// sparse TUI path.
		fieldCount := 0

		// Build lookup map from source content fields
		sourceFieldMap := make(map[types.FieldID]string)
		if sourceFields != nil {
			for _, sf := range *sourceFields {
				if sf.FieldID.Valid {
					sourceFieldMap[sf.FieldID.ID] = sf.FieldValue
				}
			}
		}

		if source.DatatypeID.Valid {
			allFields, fieldListErr := d.ListFieldsByDatatypeID(source.DatatypeID)
			if fieldListErr != nil {
				logger.Ferror("failed to list datatype fields for copy, falling back to source fields only", fieldListErr)
			}

			if allFields != nil && len(*allFields) > 0 {
				for _, field := range *allFields {
					value := sourceFieldMap[field.FieldID] // "" if not in source

					_, fieldErr := d.CreateContentField(ctx, ac, db.CreateContentFieldParams{
						RouteID:       source.RouteID,
						ContentDataID: types.NullableContentID{ID: newContent.ContentDataID, Valid: true},
						FieldID:       types.NullableFieldID{ID: field.FieldID, Valid: true},
						FieldValue:    value,
						AuthorID:      userID,
						DateCreated:   now,
						DateModified:  now,
					})
					if fieldErr != nil {
						logger.Ferror(fmt.Sprintf("failed to copy field: %v", fieldErr), fieldErr)
					}
					fieldCount++
				}
			} else if fieldListErr != nil {
				// Fallback: copy only source fields when canonical list unavailable
				if sourceFields != nil {
					for _, field := range *sourceFields {
						_, fieldErr := d.CreateContentField(ctx, ac, db.CreateContentFieldParams{
							RouteID:       field.RouteID,
							ContentDataID: types.NullableContentID{ID: newContent.ContentDataID, Valid: true},
							FieldID:       field.FieldID,
							FieldValue:    field.FieldValue,
							AuthorID:      userID,
							DateCreated:   now,
							DateModified:  now,
						})
						if fieldErr != nil {
							logger.Ferror(fmt.Sprintf("failed to copy field: %v", fieldErr), fieldErr)
						}
						fieldCount++
					}
				}
			}
		} else {
			// No datatype — just copy whatever source fields exist
			if sourceFields != nil {
				for _, field := range *sourceFields {
					_, fieldErr := d.CreateContentField(ctx, ac, db.CreateContentFieldParams{
						RouteID:       field.RouteID,
						ContentDataID: types.NullableContentID{ID: newContent.ContentDataID, Valid: true},
						FieldID:       field.FieldID,
						FieldValue:    field.FieldValue,
						AuthorID:      userID,
						DateCreated:   now,
						DateModified:  now,
					})
					if fieldErr != nil {
						logger.Ferror(fmt.Sprintf("failed to copy field: %v", fieldErr), fieldErr)
					}
					fieldCount++
				}
			}
		}

		logger.Finfo(fmt.Sprintf("Content copied: %s -> %s with %d fields", msg.SourceContentID, newContent.ContentDataID, fieldCount))
		return ContentCopiedMsg{
			SourceContentID: msg.SourceContentID,
			NewContentID:    newContent.ContentDataID,
			RouteID:         msg.RouteID,
			FieldCount:      fieldCount,
		}
	}
}

// TogglePublishCmd creates a command to show the publish/unpublish confirmation dialog.
func TogglePublishCmd(contentID types.ContentID, routeID types.RouteID) tea.Cmd {
	return func() tea.Msg {
		return TogglePublishRequestMsg{
			ContentID: contentID,
			RouteID:   routeID,
		}
	}
}

// HandleTogglePublish shows a confirmation dialog before publishing or unpublishing.
// The actual operation runs after the user confirms via ConfirmedPublishMsg / ConfirmedUnpublishMsg.
func (m Model) HandleTogglePublish(msg TogglePublishRequestMsg) tea.Cmd {
	cfg := m.Config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "configuration not loaded"}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		content, err := d.GetContentData(msg.ContentID)
		if err != nil || content == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Content not found: %v", err)}
		}

		contentName := msg.ContentID.String()
		if len(contentName) > 12 {
			contentName = contentName[:8] + "..."
		}
		isPublished := content.Status == types.ContentStatusPublished

		return ShowPublishDialogMsg{
			ContentID:   msg.ContentID,
			RouteID:     msg.RouteID,
			ContentName: contentName,
			IsPublished: isPublished,
		}
	}
}

// HandleConfirmedPublish creates a snapshot and publishes content.
func (m Model) HandleConfirmedPublish(msg ConfirmedPublishMsg) tea.Cmd {
	cfg := m.Config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "configuration not loaded"}
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
		_, pubErr := publishing.PublishContent(ctx, d, msg.ContentID, locale, userID, ac, retentionCap, publishAll, dispatcher, nil)
		if pubErr != nil {
			logger.Ferror(fmt.Sprintf("failed to publish content %s", msg.ContentID), pubErr)
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Publish failed: %v", pubErr)}
		}

		logger.Finfo(fmt.Sprintf("Content %s published via snapshot", msg.ContentID))
		return PublishCompletedMsg{
			ContentID: msg.ContentID,
			RouteID:   msg.RouteID,
		}
	}
}

// HandleConfirmedUnpublish unpublishes content (sets status to draft, clears published metadata).
func (m Model) HandleConfirmedUnpublish(msg ConfirmedUnpublishMsg) tea.Cmd {
	cfg := m.Config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "configuration not loaded"}
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

		unpubErr := publishing.UnpublishContent(ctx, d, msg.ContentID, locale, userID, ac, dispatcher, nil)
		if unpubErr != nil {
			logger.Ferror(fmt.Sprintf("failed to unpublish content %s", msg.ContentID), unpubErr)
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Unpublish failed: %v", unpubErr)}
		}

		logger.Finfo(fmt.Sprintf("Content %s unpublished", msg.ContentID))
		return UnpublishCompletedMsg{
			ContentID: msg.ContentID,
			RouteID:   msg.RouteID,
		}
	}
}

// ListVersionsCmd creates a command to list versions for a content item.
func ListVersionsCmd(contentID types.ContentID, routeID types.RouteID) tea.Cmd {
	return func() tea.Msg {
		return ListVersionsRequestMsg{
			ContentID: contentID,
			RouteID:   routeID,
		}
	}
}

// HandleListVersions fetches versions for a content item.
func (m Model) HandleListVersions(msg ListVersionsRequestMsg) tea.Cmd {
	cfg := m.Config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "configuration not loaded"}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		versions, err := d.ListContentVersionsByContent(msg.ContentID)
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("failed to list versions: %v", err)}
		}
		var versionList []db.ContentVersion
		if versions != nil {
			versionList = *versions
		}
		return VersionsListedMsg{
			ContentID: msg.ContentID,
			RouteID:   msg.RouteID,
			Versions:  versionList,
		}
	}
}

// ShowRestoreVersionDialogMsg requests showing the restore version confirmation dialog.
type ShowRestoreVersionDialogMsg struct {
	ContentID     types.ContentID
	VersionID     types.ContentVersionID
	RouteID       types.RouteID
	VersionNumber int64
}

// ShowRestoreVersionDialogCmd creates a command to show the restore version dialog.
func ShowRestoreVersionDialogCmd(contentID types.ContentID, versionID types.ContentVersionID, routeID types.RouteID, versionNumber int64) tea.Cmd {
	return func() tea.Msg {
		return ShowRestoreVersionDialogMsg{
			ContentID:     contentID,
			VersionID:     versionID,
			RouteID:       routeID,
			VersionNumber: versionNumber,
		}
	}
}

// HandleConfirmedRestoreVersion restores content from a saved version.
func (m Model) HandleConfirmedRestoreVersion(msg ConfirmedRestoreVersionMsg) tea.Cmd {
	cfg := m.Config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "configuration not loaded"}
		}
	}

	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)
		logger := utility.DefaultLogger

		result, err := publishing.RestoreContent(ctx, d, msg.ContentID, msg.VersionID, userID, ac)
		if err != nil {
			logger.Ferror(fmt.Sprintf("failed to restore version for %s", msg.ContentID), err)
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("restore failed: %v", err)}
		}

		logger.Finfo(fmt.Sprintf("Content %s restored from version %s (%d fields)", msg.ContentID, msg.VersionID, result.FieldsRestored))
		return VersionRestoredMsg{
			ContentID:      msg.ContentID,
			RouteID:        msg.RouteID,
			FieldsRestored: result.FieldsRestored,
		}
	}
}

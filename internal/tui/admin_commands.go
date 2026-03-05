package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
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
		return AdminContentCreatedMsg{
			AdminContentID: contentData.AdminContentDataID,
			AdminRouteID:   routeID,
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

		content, err := d.GetAdminContentData(msg.AdminContentID)
		if err != nil || content == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Admin content not found: %v", err)}
		}

		if content.FirstChildID.Valid && !content.FirstChildID.ID.IsZero() {
			return ActionResultMsg{Title: "Cannot Delete", Message: "This content has children. Delete child nodes first."}
		}

		// Update sibling pointers before deletion
		if content.PrevSiblingID.Valid && !content.PrevSiblingID.ID.IsZero() {
			prevSibling, prevErr := d.GetAdminContentData(content.PrevSiblingID.ID)
			if prevErr == nil && prevSibling != nil {
				params := db.UpdateAdminContentDataParams{
					AdminContentDataID: prevSibling.AdminContentDataID,
					AdminRouteID:       prevSibling.AdminRouteID,
					AdminDatatypeID:    prevSibling.AdminDatatypeID,
					ParentID:           prevSibling.ParentID,
					FirstChildID:       prevSibling.FirstChildID,
					NextSiblingID:      content.NextSiblingID,
					PrevSiblingID:      prevSibling.PrevSiblingID,
					AuthorID:           prevSibling.AuthorID,
					Status:             prevSibling.Status,
					DateCreated:        prevSibling.DateCreated,
					DateModified:       types.TimestampNow(),
				}
				if _, updateErr := d.UpdateAdminContentData(ctx, ac, params); updateErr != nil {
					logger.Ferror("Failed to update prev sibling during admin delete", updateErr)
				}
			}
		}

		if content.NextSiblingID.Valid && !content.NextSiblingID.ID.IsZero() {
			nextSibling, nextErr := d.GetAdminContentData(content.NextSiblingID.ID)
			if nextErr == nil && nextSibling != nil {
				params := db.UpdateAdminContentDataParams{
					AdminContentDataID: nextSibling.AdminContentDataID,
					AdminRouteID:       nextSibling.AdminRouteID,
					AdminDatatypeID:    nextSibling.AdminDatatypeID,
					ParentID:           nextSibling.ParentID,
					FirstChildID:       nextSibling.FirstChildID,
					NextSiblingID:      nextSibling.NextSiblingID,
					PrevSiblingID:      content.PrevSiblingID,
					AuthorID:           nextSibling.AuthorID,
					Status:             nextSibling.Status,
					DateCreated:        nextSibling.DateCreated,
					DateModified:       types.TimestampNow(),
				}
				if _, updateErr := d.UpdateAdminContentData(ctx, ac, params); updateErr != nil {
					logger.Ferror("Failed to update next sibling during admin delete", updateErr)
				}
			}
		}

		if content.ParentID.Valid && !content.ParentID.ID.IsZero() {
			parent, parentErr := d.GetAdminContentData(content.ParentID.ID)
			if parentErr == nil && parent != nil {
				if parent.FirstChildID.Valid && parent.FirstChildID.ID == msg.AdminContentID {
					params := db.UpdateAdminContentDataParams{
						AdminContentDataID: parent.AdminContentDataID,
						AdminRouteID:       parent.AdminRouteID,
						AdminDatatypeID:    parent.AdminDatatypeID,
						ParentID:           parent.ParentID,
						FirstChildID:       content.NextSiblingID,
						NextSiblingID:      parent.NextSiblingID,
						PrevSiblingID:      parent.PrevSiblingID,
						AuthorID:           parent.AuthorID,
						Status:             parent.Status,
						DateCreated:        parent.DateCreated,
						DateModified:       types.TimestampNow(),
					}
					if _, updateErr := d.UpdateAdminContentData(ctx, ac, params); updateErr != nil {
						logger.Ferror("Failed to update parent first_child during admin delete", updateErr)
					}
				}
			}
		}

		if deleteErr := d.DeleteAdminContentData(ctx, ac, msg.AdminContentID); deleteErr != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to delete admin content: %v", deleteErr)}
		}

		logger.Finfo(fmt.Sprintf("Admin content deleted: %s", msg.AdminContentID))
		return AdminContentDeletedMsg{AdminContentID: msg.AdminContentID, AdminRouteID: msg.AdminRouteID}
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

		source, err := d.GetAdminContentData(msg.SourceID)
		if err != nil || source == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Source admin content not found: %v", err)}
		}

		target, err := d.GetAdminContentData(msg.TargetID)
		if err != nil || target == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Target admin content not found: %v", err)}
		}

		// STEP 1: Detach source from old position
		if source.PrevSiblingID.Valid && !source.PrevSiblingID.ID.IsZero() {
			prev, prevErr := d.GetAdminContentData(source.PrevSiblingID.ID)
			if prevErr == nil && prev != nil {
				params := db.UpdateAdminContentDataParams{
					AdminContentDataID: prev.AdminContentDataID,
					AdminRouteID:       prev.AdminRouteID,
					AdminDatatypeID:    prev.AdminDatatypeID,
					ParentID:           prev.ParentID,
					FirstChildID:       prev.FirstChildID,
					NextSiblingID:      source.NextSiblingID,
					PrevSiblingID:      prev.PrevSiblingID,
					AuthorID:           prev.AuthorID,
					Status:             prev.Status,
					DateCreated:        prev.DateCreated,
					DateModified:       types.TimestampNow(),
				}
				if _, updateErr := d.UpdateAdminContentData(ctx, ac, params); updateErr != nil {
					logger.Ferror("Failed to update prev sibling during admin move", updateErr)
				}
			}
		}

		if source.NextSiblingID.Valid && !source.NextSiblingID.ID.IsZero() {
			next, nextErr := d.GetAdminContentData(source.NextSiblingID.ID)
			if nextErr == nil && next != nil {
				params := db.UpdateAdminContentDataParams{
					AdminContentDataID: next.AdminContentDataID,
					AdminRouteID:       next.AdminRouteID,
					AdminDatatypeID:    next.AdminDatatypeID,
					ParentID:           next.ParentID,
					FirstChildID:       next.FirstChildID,
					NextSiblingID:      next.NextSiblingID,
					PrevSiblingID:      source.PrevSiblingID,
					AuthorID:           next.AuthorID,
					Status:             next.Status,
					DateCreated:        next.DateCreated,
					DateModified:       types.TimestampNow(),
				}
				if _, updateErr := d.UpdateAdminContentData(ctx, ac, params); updateErr != nil {
					logger.Ferror("Failed to update next sibling during admin move", updateErr)
				}
			}
		}

		if source.ParentID.Valid && !source.ParentID.ID.IsZero() {
			oldParent, parentErr := d.GetAdminContentData(source.ParentID.ID)
			if parentErr == nil && oldParent != nil {
				if oldParent.FirstChildID.Valid && oldParent.FirstChildID.ID == source.AdminContentDataID {
					params := db.UpdateAdminContentDataParams{
						AdminContentDataID: oldParent.AdminContentDataID,
						AdminRouteID:       oldParent.AdminRouteID,
						AdminDatatypeID:    oldParent.AdminDatatypeID,
						ParentID:           oldParent.ParentID,
						FirstChildID:       source.NextSiblingID,
						NextSiblingID:      oldParent.NextSiblingID,
						PrevSiblingID:      oldParent.PrevSiblingID,
						AuthorID:           oldParent.AuthorID,
						Status:             oldParent.Status,
						DateCreated:        oldParent.DateCreated,
						DateModified:       types.TimestampNow(),
					}
					if _, updateErr := d.UpdateAdminContentData(ctx, ac, params); updateErr != nil {
						logger.Ferror("Failed to update old parent during admin move", updateErr)
					}
				}
			}
		}

		// STEP 2: Attach source as last child of target
		target, err = d.GetAdminContentData(msg.TargetID)
		if err != nil || target == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Target not found after detach: %v", err)}
		}

		newPrevSiblingID := types.NullableAdminContentID{}

		if !target.FirstChildID.Valid || target.FirstChildID.ID.IsZero() {
			targetParams := db.UpdateAdminContentDataParams{
				AdminContentDataID: target.AdminContentDataID,
				AdminRouteID:       target.AdminRouteID,
				AdminDatatypeID:    target.AdminDatatypeID,
				ParentID:           target.ParentID,
				FirstChildID:       types.NullableAdminContentID{ID: source.AdminContentDataID, Valid: true},
				NextSiblingID:      target.NextSiblingID,
				PrevSiblingID:      target.PrevSiblingID,
				AuthorID:           target.AuthorID,
				Status:             target.Status,
				DateCreated:        target.DateCreated,
				DateModified:       types.TimestampNow(),
			}
			if _, updateErr := d.UpdateAdminContentData(ctx, ac, targetParams); updateErr != nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to update target: %v", updateErr)}
			}
		} else {
			currentID := target.FirstChildID.ID
			for {
				current, walkErr := d.GetAdminContentData(currentID)
				if walkErr != nil || current == nil {
					break
				}
				if !current.NextSiblingID.Valid || current.NextSiblingID.ID.IsZero() {
					lastParams := db.UpdateAdminContentDataParams{
						AdminContentDataID: current.AdminContentDataID,
						AdminRouteID:       current.AdminRouteID,
						AdminDatatypeID:    current.AdminDatatypeID,
						ParentID:           current.ParentID,
						FirstChildID:       current.FirstChildID,
						NextSiblingID:      types.NullableAdminContentID{ID: source.AdminContentDataID, Valid: true},
						PrevSiblingID:      current.PrevSiblingID,
						AuthorID:           current.AuthorID,
						Status:             current.Status,
						DateCreated:        current.DateCreated,
						DateModified:       types.TimestampNow(),
					}
					if _, updateErr := d.UpdateAdminContentData(ctx, ac, lastParams); updateErr != nil {
						logger.Ferror("Failed to update last sibling during admin move", updateErr)
					}
					newPrevSiblingID = types.NullableAdminContentID{ID: current.AdminContentDataID, Valid: true}
					break
				}
				currentID = current.NextSiblingID.ID
			}
		}

		// STEP 3: Update source with new parent
		sourceParams := db.UpdateAdminContentDataParams{
			AdminContentDataID: source.AdminContentDataID,
			AdminRouteID:       source.AdminRouteID,
			AdminDatatypeID:    source.AdminDatatypeID,
			ParentID:           types.NullableAdminContentID{ID: msg.TargetID, Valid: true},
			FirstChildID:       source.FirstChildID,
			NextSiblingID:      types.NullableAdminContentID{},
			PrevSiblingID:      newPrevSiblingID,
			AuthorID:           source.AuthorID,
			Status:             source.Status,
			DateCreated:        source.DateCreated,
			DateModified:       types.TimestampNow(),
		}
		if _, updateErr := d.UpdateAdminContentData(ctx, ac, sourceParams); updateErr != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to update source: %v", updateErr)}
		}

		logger.Finfo(fmt.Sprintf("Admin content moved: %s -> %s", msg.SourceID, msg.TargetID))
		return AdminContentMovedMsg{AdminContentID: msg.SourceID, AdminRouteID: msg.AdminRouteID}
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

		a, err := d.GetAdminContentData(msg.AdminContentID)
		if err != nil || a == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Admin content not found: %v", err)}
		}

		if msg.Direction == "up" {
			if !a.PrevSiblingID.Valid || a.PrevSiblingID.ID.IsZero() {
				return ActionResultMsg{Title: "Info", Message: "Already at top"}
			}
			b, bErr := d.GetAdminContentData(a.PrevSiblingID.ID)
			if bErr != nil || b == nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Previous sibling not found: %v", bErr)}
			}

			// If B has a prev (C), update C.Next -> A
			if b.PrevSiblingID.Valid && !b.PrevSiblingID.ID.IsZero() {
				c, cErr := d.GetAdminContentData(b.PrevSiblingID.ID)
				if cErr == nil && c != nil {
					_, updateErr := d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
						AdminContentDataID: c.AdminContentDataID, AdminRouteID: c.AdminRouteID, AdminDatatypeID: c.AdminDatatypeID,
						ParentID: c.ParentID, FirstChildID: c.FirstChildID,
						NextSiblingID: types.NullableAdminContentID{ID: a.AdminContentDataID, Valid: true},
						PrevSiblingID: c.PrevSiblingID,
						AuthorID:      c.AuthorID, Status: c.Status, DateCreated: c.DateCreated, DateModified: types.TimestampNow(),
					})
					if updateErr != nil {
						logger.Ferror("Failed to update C during admin reorder up", updateErr)
					}
				}
			}

			// If A has a next (D), update D.Prev -> B
			if a.NextSiblingID.Valid && !a.NextSiblingID.ID.IsZero() {
				dNode, dErr := d.GetAdminContentData(a.NextSiblingID.ID)
				if dErr == nil && dNode != nil {
					_, updateErr := d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
						AdminContentDataID: dNode.AdminContentDataID, AdminRouteID: dNode.AdminRouteID, AdminDatatypeID: dNode.AdminDatatypeID,
						ParentID: dNode.ParentID, FirstChildID: dNode.FirstChildID,
						NextSiblingID: dNode.NextSiblingID,
						PrevSiblingID: types.NullableAdminContentID{ID: b.AdminContentDataID, Valid: true},
						AuthorID:      dNode.AuthorID, Status: dNode.Status, DateCreated: dNode.DateCreated, DateModified: types.TimestampNow(),
					})
					if updateErr != nil {
						logger.Ferror("Failed to update D during admin reorder up", updateErr)
					}
				}
			}

			// If parent.FirstChildID == B, update to A
			if a.ParentID.Valid && !a.ParentID.ID.IsZero() {
				parent, pErr := d.GetAdminContentData(a.ParentID.ID)
				if pErr == nil && parent != nil && parent.FirstChildID.Valid && parent.FirstChildID.ID == b.AdminContentDataID {
					_, updateErr := d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
						AdminContentDataID: parent.AdminContentDataID, AdminRouteID: parent.AdminRouteID, AdminDatatypeID: parent.AdminDatatypeID,
						ParentID:      parent.ParentID,
						FirstChildID:  types.NullableAdminContentID{ID: a.AdminContentDataID, Valid: true},
						NextSiblingID: parent.NextSiblingID, PrevSiblingID: parent.PrevSiblingID,
						AuthorID: parent.AuthorID, Status: parent.Status, DateCreated: parent.DateCreated, DateModified: types.TimestampNow(),
					})
					if updateErr != nil {
						logger.Ferror("Failed to update parent during admin reorder up", updateErr)
					}
				}
			}

			// Update A: prev = B.prev, next = B
			_, aErr := d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
				AdminContentDataID: a.AdminContentDataID, AdminRouteID: a.AdminRouteID, AdminDatatypeID: a.AdminDatatypeID,
				ParentID: a.ParentID, FirstChildID: a.FirstChildID,
				NextSiblingID: types.NullableAdminContentID{ID: b.AdminContentDataID, Valid: true},
				PrevSiblingID: b.PrevSiblingID,
				AuthorID:      a.AuthorID, Status: a.Status, DateCreated: a.DateCreated, DateModified: types.TimestampNow(),
			})
			if aErr != nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to update node: %v", aErr)}
			}

			// Update B: prev = A, next = A's original next
			_, bUpdateErr := d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
				AdminContentDataID: b.AdminContentDataID, AdminRouteID: b.AdminRouteID, AdminDatatypeID: b.AdminDatatypeID,
				ParentID: b.ParentID, FirstChildID: b.FirstChildID,
				NextSiblingID: a.NextSiblingID,
				PrevSiblingID: types.NullableAdminContentID{ID: a.AdminContentDataID, Valid: true},
				AuthorID:      b.AuthorID, Status: b.Status, DateCreated: b.DateCreated, DateModified: types.TimestampNow(),
			})
			if bUpdateErr != nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to update sibling: %v", bUpdateErr)}
			}

		} else {
			// Move down: swap A with its next sibling B
			if !a.NextSiblingID.Valid || a.NextSiblingID.ID.IsZero() {
				return ActionResultMsg{Title: "Info", Message: "Already at bottom"}
			}
			b, bErr := d.GetAdminContentData(a.NextSiblingID.ID)
			if bErr != nil || b == nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Next sibling not found: %v", bErr)}
			}

			// If A has a prev (C), update C.Next -> B
			if a.PrevSiblingID.Valid && !a.PrevSiblingID.ID.IsZero() {
				c, cErr := d.GetAdminContentData(a.PrevSiblingID.ID)
				if cErr == nil && c != nil {
					_, updateErr := d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
						AdminContentDataID: c.AdminContentDataID, AdminRouteID: c.AdminRouteID, AdminDatatypeID: c.AdminDatatypeID,
						ParentID: c.ParentID, FirstChildID: c.FirstChildID,
						NextSiblingID: types.NullableAdminContentID{ID: b.AdminContentDataID, Valid: true},
						PrevSiblingID: c.PrevSiblingID,
						AuthorID:      c.AuthorID, Status: c.Status, DateCreated: c.DateCreated, DateModified: types.TimestampNow(),
					})
					if updateErr != nil {
						logger.Ferror("Failed to update C during admin reorder down", updateErr)
					}
				}
			}

			// If B has a next (D), update D.Prev -> A
			if b.NextSiblingID.Valid && !b.NextSiblingID.ID.IsZero() {
				dNode, dErr := d.GetAdminContentData(b.NextSiblingID.ID)
				if dErr == nil && dNode != nil {
					_, updateErr := d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
						AdminContentDataID: dNode.AdminContentDataID, AdminRouteID: dNode.AdminRouteID, AdminDatatypeID: dNode.AdminDatatypeID,
						ParentID: dNode.ParentID, FirstChildID: dNode.FirstChildID,
						NextSiblingID: dNode.NextSiblingID,
						PrevSiblingID: types.NullableAdminContentID{ID: a.AdminContentDataID, Valid: true},
						AuthorID:      dNode.AuthorID, Status: dNode.Status, DateCreated: dNode.DateCreated, DateModified: types.TimestampNow(),
					})
					if updateErr != nil {
						logger.Ferror("Failed to update D during admin reorder down", updateErr)
					}
				}
			}

			// If parent.FirstChildID == A, update to B
			if a.ParentID.Valid && !a.ParentID.ID.IsZero() {
				parent, pErr := d.GetAdminContentData(a.ParentID.ID)
				if pErr == nil && parent != nil && parent.FirstChildID.Valid && parent.FirstChildID.ID == a.AdminContentDataID {
					_, updateErr := d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
						AdminContentDataID: parent.AdminContentDataID, AdminRouteID: parent.AdminRouteID, AdminDatatypeID: parent.AdminDatatypeID,
						ParentID:      parent.ParentID,
						FirstChildID:  types.NullableAdminContentID{ID: b.AdminContentDataID, Valid: true},
						NextSiblingID: parent.NextSiblingID, PrevSiblingID: parent.PrevSiblingID,
						AuthorID: parent.AuthorID, Status: parent.Status, DateCreated: parent.DateCreated, DateModified: types.TimestampNow(),
					})
					if updateErr != nil {
						logger.Ferror("Failed to update parent during admin reorder down", updateErr)
					}
				}
			}

			// Update B: prev = A.prev, next = A
			_, bUpdateErr := d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
				AdminContentDataID: b.AdminContentDataID, AdminRouteID: b.AdminRouteID, AdminDatatypeID: b.AdminDatatypeID,
				ParentID: b.ParentID, FirstChildID: b.FirstChildID,
				NextSiblingID: types.NullableAdminContentID{ID: a.AdminContentDataID, Valid: true},
				PrevSiblingID: a.PrevSiblingID,
				AuthorID:      b.AuthorID, Status: b.Status, DateCreated: b.DateCreated, DateModified: types.TimestampNow(),
			})
			if bUpdateErr != nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to update sibling: %v", bUpdateErr)}
			}

			// Update A: prev = B, next = B's original next
			_, aErr := d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
				AdminContentDataID: a.AdminContentDataID, AdminRouteID: a.AdminRouteID, AdminDatatypeID: a.AdminDatatypeID,
				ParentID: a.ParentID, FirstChildID: a.FirstChildID,
				NextSiblingID: b.NextSiblingID,
				PrevSiblingID: types.NullableAdminContentID{ID: b.AdminContentDataID, Valid: true},
				AuthorID:      a.AuthorID, Status: a.Status, DateCreated: a.DateCreated, DateModified: types.TimestampNow(),
			})
			if aErr != nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to update node: %v", aErr)}
			}
		}

		logger.Finfo(fmt.Sprintf("Admin content reordered %s: %s", msg.Direction, msg.AdminContentID))
		return AdminContentReorderedMsg{
			AdminContentID: msg.AdminContentID,
			AdminRouteID:   msg.AdminRouteID,
			Direction:      msg.Direction,
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

		// Update source.Next -> new node
		_, sErr := d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
			AdminContentDataID: source.AdminContentDataID,
			AdminRouteID:       source.AdminRouteID,
			AdminDatatypeID:    source.AdminDatatypeID,
			ParentID:           source.ParentID,
			FirstChildID:       source.FirstChildID,
			NextSiblingID:      types.NullableAdminContentID{ID: newContent.AdminContentDataID, Valid: true},
			PrevSiblingID:      source.PrevSiblingID,
			AuthorID:           source.AuthorID,
			Status:             source.Status,
			DateCreated:        source.DateCreated,
			DateModified:       types.TimestampNow(),
		})
		if sErr != nil {
			logger.Ferror("Failed to update source next pointer after admin copy", sErr)
		}

		// If source had a next sibling (D), update D.Prev -> new node
		if source.NextSiblingID.Valid && !source.NextSiblingID.ID.IsZero() {
			dNode, dErr := d.GetAdminContentData(source.NextSiblingID.ID)
			if dErr == nil && dNode != nil {
				_, updateErr := d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
					AdminContentDataID: dNode.AdminContentDataID,
					AdminRouteID:       dNode.AdminRouteID,
					AdminDatatypeID:    dNode.AdminDatatypeID,
					ParentID:           dNode.ParentID,
					FirstChildID:       dNode.FirstChildID,
					NextSiblingID:      dNode.NextSiblingID,
					PrevSiblingID:      types.NullableAdminContentID{ID: newContent.AdminContentDataID, Valid: true},
					AuthorID:           dNode.AuthorID,
					Status:             dNode.Status,
					DateCreated:        dNode.DateCreated,
					DateModified:       types.TimestampNow(),
				})
				if updateErr != nil {
					logger.Ferror("Failed to update next sibling prev after admin copy", updateErr)
				}
			}
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
		return AdminContentCopiedMsg{NewID: newContent.AdminContentDataID, AdminRouteID: msg.AdminRouteID}
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
		pubErr := publishing.PublishAdminContent(ctx, d, msg.AdminContentID, locale, userID, ac, retentionCap, dispatcher)
		if pubErr != nil {
			logger.Ferror(fmt.Sprintf("Failed to publish admin content %s", msg.AdminContentID), pubErr)
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Publish failed: %v", pubErr)}
		}

		logger.Finfo(fmt.Sprintf("Admin content %s published", msg.AdminContentID))
		return AdminPublishCompletedMsg{AdminContentID: msg.AdminContentID, AdminRouteID: msg.AdminRouteID}
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
		return AdminUnpublishCompletedMsg{AdminContentID: msg.AdminContentID, AdminRouteID: msg.AdminRouteID}
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
		return AdminVersionRestoredMsg{
			AdminContentID: msg.AdminContentID,
			AdminRouteID:   msg.AdminRouteID,
			FieldsRestored: result.FieldsRestored,
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

		return AdminContentFieldDeletedMsg{
			AdminContentID: msg.AdminContentID,
			AdminRouteID:   msg.AdminRouteID,
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

		return AdminContentFieldAddedMsg{
			AdminContentID: adminContentID,
			AdminRouteID:   adminRouteID,
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

		return AdminContentFieldUpdatedMsg{
			AdminContentID: adminContentID,
			AdminRouteID:   adminRouteID,
		}
	}
}

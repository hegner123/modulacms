package cli

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
)

// =============================================================================
// DELETE CONTEXTS (global vars, same pattern as regular CMS)
// =============================================================================

// DeleteAdminRouteContext stores context for deleting an admin route.
type DeleteAdminRouteContext struct {
	AdminRouteID types.AdminRouteID
	Title        string
}

var deleteAdminRouteContext *DeleteAdminRouteContext

// DeleteAdminDatatypeContext stores context for deleting an admin datatype.
type DeleteAdminDatatypeContext struct {
	AdminDatatypeID types.AdminDatatypeID
	Label           string
}

var deleteAdminDatatypeContext *DeleteAdminDatatypeContext

// DeleteAdminFieldContext stores context for deleting an admin field.
type DeleteAdminFieldContext struct {
	AdminFieldID    types.AdminFieldID
	AdminDatatypeID types.AdminDatatypeID
	Label           string
}

var deleteAdminFieldContext *DeleteAdminFieldContext

// =============================================================================
// ADMIN ROUTE HANDLERS
// =============================================================================

// HandleCreateAdminRouteFromDialog processes the admin route creation request.
func (m Model) HandleCreateAdminRouteFromDialog(msg CreateAdminRouteFromDialogRequestMsg) tea.Cmd {
	authorID := m.UserID
	cfg := m.Config

	if authorID.IsZero() {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create admin route: no user is logged in",
			}
		}
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create admin route: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		// Prepare slug
		slug := msg.Slug
		if slug == "" {
			slug = msg.Title
		}
		validSlug := types.Slugify(slug)
		if err := validSlug.Validate(); err != nil {
			return ActionResultMsg{
				Title:   "Invalid Slug",
				Message: fmt.Sprintf("Could not create admin route: %v", err),
			}
		}

		params := db.CreateAdminRouteParams{
			Slug:  validSlug,
			Title: msg.Title,
			Status: 1,
			AuthorID: types.NullableUserID{
				ID:    authorID,
				Valid: true,
			},
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		route, err := d.CreateAdminRoute(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to create admin route: %v", err),
			}
		}

		return AdminRouteCreatedFromDialogMsg{
			AdminRouteID: route.AdminRouteID,
			Title:        route.Title,
			Slug:         string(route.Slug),
		}
	}
}

// HandleUpdateAdminRouteFromDialog processes the admin route update request.
func (m Model) HandleUpdateAdminRouteFromDialog(msg UpdateAdminRouteFromDialogRequestMsg) tea.Cmd {
	authorID := m.UserID
	cfg := m.Config

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot update admin route: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		// Fetch existing route to get original slug for WHERE clause
		originalSlug := types.Slug(msg.OriginalSlug)
		existingRoute, err := d.GetAdminRoute(originalSlug)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Admin route not found: %v", err),
			}
		}

		// Prepare slug
		slug := msg.Slug
		if slug == "" {
			slug = msg.Title
		}
		validSlug := types.Slugify(slug)
		if err := validSlug.Validate(); err != nil {
			return ActionResultMsg{
				Title:   "Invalid Slug",
				Message: fmt.Sprintf("Could not update admin route: %v", err),
			}
		}

		params := db.UpdateAdminRouteParams{
			Slug:  validSlug,
			Title: msg.Title,
			Status: existingRoute.Status,
			AuthorID: types.NullableUserID{
				ID:    authorID,
				Valid: !authorID.IsZero(),
			},
			DateCreated:  existingRoute.DateCreated,
			DateModified: types.TimestampNow(),
			Slug_2:       existingRoute.Slug, // Original slug for WHERE clause
		}

		_, err = d.UpdateAdminRoute(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update admin route: %v", err),
			}
		}

		return AdminRouteUpdatedFromDialogMsg{
			AdminRouteID: existingRoute.AdminRouteID,
			Title:        msg.Title,
			Slug:         string(validSlug),
		}
	}
}

// HandleDeleteAdminRoute processes the admin route deletion request.
func (m Model) HandleDeleteAdminRoute(msg DeleteAdminRouteRequestMsg) tea.Cmd {
	cfg := m.Config
	authorID := m.UserID

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot delete admin route: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		err := d.DeleteAdminRoute(ctx, ac, msg.AdminRouteID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to delete admin route: %v", err),
			}
		}

		return AdminRouteDeletedMsg{AdminRouteID: msg.AdminRouteID}
	}
}

// =============================================================================
// ADMIN DATATYPE HANDLERS
// =============================================================================

// HandleCreateAdminDatatypeFromDialog processes the admin datatype creation request.
func (m Model) HandleCreateAdminDatatypeFromDialog(msg CreateAdminDatatypeFromDialogRequestMsg) tea.Cmd {
	authorID := m.UserID
	cfg := m.Config

	if authorID.IsZero() {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create admin datatype: no user is logged in",
			}
		}
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create admin datatype: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		dtype := msg.Type
		if dtype == "" {
			dtype = "ROOT"
		}

		var parentID types.NullableAdminDatatypeID
		if msg.ParentID != "" {
			parentID = types.NullableAdminDatatypeID{
				ID:    types.AdminDatatypeID(msg.ParentID),
				Valid: true,
			}
		}

		params := db.CreateAdminDatatypeParams{
			ParentID: parentID,
			Label:    msg.Label,
			Type:     dtype,
			AuthorID: types.NullableUserID{
				ID:    authorID,
				Valid: true,
			},
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		dt, err := d.CreateAdminDatatype(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to create admin datatype: %v", err),
			}
		}

		return AdminDatatypeCreatedFromDialogMsg{
			AdminDatatypeID: dt.AdminDatatypeID,
			Label:           dt.Label,
		}
	}
}

// HandleUpdateAdminDatatypeFromDialog processes the admin datatype update request.
func (m Model) HandleUpdateAdminDatatypeFromDialog(msg UpdateAdminDatatypeFromDialogRequestMsg) tea.Cmd {
	cfg := m.Config
	authorID := m.UserID

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot update admin datatype: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		datatypeID := types.AdminDatatypeID(msg.AdminDatatypeID)
		existing, err := d.GetAdminDatatypeById(datatypeID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to get admin datatype for update: %v", err),
			}
		}

		dtype := msg.Type
		if dtype == "" {
			dtype = existing.Type
		}

		parentID := existing.ParentID
		if msg.ParentID != "" {
			parentID = types.NullableAdminDatatypeID{
				ID:    types.AdminDatatypeID(msg.ParentID),
				Valid: true,
			}
		}

		params := db.UpdateAdminDatatypeParams{
			ParentID:        parentID,
			Label:           msg.Label,
			Type:            dtype,
			AuthorID:        existing.AuthorID,
			DateCreated:     existing.DateCreated,
			DateModified:    types.TimestampNow(),
			AdminDatatypeID: datatypeID,
		}

		_, err = d.UpdateAdminDatatype(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update admin datatype: %v", err),
			}
		}

		return AdminDatatypeUpdatedFromDialogMsg{
			AdminDatatypeID: datatypeID,
			Label:           msg.Label,
		}
	}
}

// HandleDeleteAdminDatatype processes the admin datatype deletion request.
// It deletes junction records first, then the entity.
func (m Model) HandleDeleteAdminDatatype(msg DeleteAdminDatatypeRequestMsg) tea.Cmd {
	cfg := m.Config
	authorID := m.UserID

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot delete admin datatype: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		// Delete junction records first
		dtFields, err := d.ListAdminDatatypeFieldByDatatypeID(msg.AdminDatatypeID)
		if err == nil && dtFields != nil {
			for _, dtf := range *dtFields {
				deleteErr := d.DeleteAdminDatatypeField(ctx, ac, dtf.ID)
				if deleteErr != nil {
					return ActionResultMsg{
						Title:   "Error",
						Message: fmt.Sprintf("Failed to delete admin datatype field junction: %v", deleteErr),
					}
				}
			}
		}

		// Delete the datatype
		err = d.DeleteAdminDatatype(ctx, ac, msg.AdminDatatypeID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to delete admin datatype: %v", err),
			}
		}

		return AdminDatatypeDeletedMsg{AdminDatatypeID: msg.AdminDatatypeID}
	}
}

// =============================================================================
// ADMIN FIELD HANDLERS
// =============================================================================

// HandleCreateAdminFieldFromDialog processes the admin field creation request.
func (m Model) HandleCreateAdminFieldFromDialog(msg CreateAdminFieldFromDialogRequestMsg) tea.Cmd {
	authorID := m.UserID
	cfg := m.Config

	if authorID.IsZero() {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create admin field: no user is logged in",
			}
		}
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create admin field: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		// Pre-flight user validation
		authorUser, userErr := d.GetUser(authorID)
		if userErr != nil || authorUser == nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Cannot create admin field: author user %s not found in database", authorID),
			}
		}

		fieldTypeStr := msg.Type
		if fieldTypeStr == "" {
			fieldTypeStr = "text"
		}
		fieldType := types.FieldType(fieldTypeStr)

		fieldParams := db.CreateAdminFieldParams{
			Label:      msg.Label,
			Data:       "",
			Validation: types.EmptyJSON,
			UIConfig:   types.EmptyJSON,
			Type:       fieldType,
			AuthorID: types.NullableUserID{
				ID:    authorID,
				Valid: true,
			},
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		field, err := d.CreateAdminField(ctx, ac, fieldParams)
		if err != nil || field == nil {
			errMsg := "Failed to create admin field in database"
			if err != nil {
				errMsg = fmt.Sprintf("Failed to create admin field: %v", err)
			}
			return ActionResultMsg{
				Title:   "Error",
				Message: errMsg,
			}
		}

		// Link field to admin datatype via junction table
		dtFieldParams := db.CreateAdminDatatypeFieldParams{
			AdminDatatypeID: msg.AdminDatatypeID,
			AdminFieldID:    field.AdminFieldID,
		}

		_, dtfErr := d.CreateAdminDatatypeField(ctx, ac, dtFieldParams)
		if dtfErr != nil {
			return ActionResultMsg{
				Title:   "Warning",
				Message: fmt.Sprintf("Admin field created but failed to link to datatype: %v", dtfErr),
			}
		}

		return AdminFieldCreatedFromDialogMsg{
			AdminFieldID:    field.AdminFieldID,
			AdminDatatypeID: msg.AdminDatatypeID,
			Label:           field.Label,
		}
	}
}

// HandleUpdateAdminFieldFromDialog processes the admin field update request.
func (m Model) HandleUpdateAdminFieldFromDialog(msg UpdateAdminFieldFromDialogRequestMsg) tea.Cmd {
	cfg := m.Config
	authorID := m.UserID

	// Capture current admin datatype ID for field refresh
	var adminDatatypeID types.AdminDatatypeID
	if len(m.AdminAllDatatypes) > 0 && m.Cursor < len(m.AdminAllDatatypes) {
		adminDatatypeID = m.AdminAllDatatypes[m.Cursor].AdminDatatypeID
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot update admin field: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		fieldID := types.AdminFieldID(msg.AdminFieldID)
		existing, err := d.GetAdminField(fieldID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to get admin field for update: %v", err),
			}
		}

		fieldTypeStr := msg.Type
		if fieldTypeStr == "" {
			fieldTypeStr = string(existing.Type)
		}
		fieldType := types.FieldType(fieldTypeStr)

		params := db.UpdateAdminFieldParams{
			ParentID:     existing.ParentID,
			Label:        msg.Label,
			Data:         existing.Data,
			Validation:   existing.Validation,
			UIConfig:     existing.UIConfig,
			Type:         fieldType,
			AuthorID:     existing.AuthorID,
			DateCreated:  existing.DateCreated,
			DateModified: types.TimestampNow(),
			AdminFieldID: fieldID,
		}

		_, err = d.UpdateAdminField(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update admin field: %v", err),
			}
		}

		return AdminFieldUpdatedFromDialogMsg{
			AdminFieldID:    fieldID,
			AdminDatatypeID: adminDatatypeID,
			Label:           msg.Label,
		}
	}
}

// HandleDeleteAdminField processes the admin field deletion request.
// It deletes junction records first, then the field.
func (m Model) HandleDeleteAdminField(msg DeleteAdminFieldRequestMsg) tea.Cmd {
	cfg := m.Config
	authorID := m.UserID

	// Capture current admin datatype ID for field refresh
	var adminDatatypeID types.AdminDatatypeID
	if len(m.AdminAllDatatypes) > 0 && m.Cursor < len(m.AdminAllDatatypes) {
		adminDatatypeID = m.AdminAllDatatypes[m.Cursor].AdminDatatypeID
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot delete admin field: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		// Delete junction records first
		dtFields, err := d.ListAdminDatatypeFieldByFieldID(msg.AdminFieldID)
		if err == nil && dtFields != nil {
			for _, dtf := range *dtFields {
				deleteErr := d.DeleteAdminDatatypeField(ctx, ac, dtf.ID)
				if deleteErr != nil {
					return ActionResultMsg{
						Title:   "Error",
						Message: fmt.Sprintf("Failed to delete admin datatype field junction: %v", deleteErr),
					}
				}
			}
		}

		// Delete the field
		err = d.DeleteAdminField(ctx, ac, msg.AdminFieldID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to delete admin field: %v", err),
			}
		}

		return AdminFieldDeletedMsg{
			AdminFieldID:    msg.AdminFieldID,
			AdminDatatypeID: adminDatatypeID,
		}
	}
}

// =============================================================================
// ADMIN CONTENT HANDLERS
// =============================================================================

// HandleDeleteAdminContent processes the admin content deletion request.
func (m Model) HandleDeleteAdminContent(msg DeleteAdminContentRequestMsg) tea.Cmd {
	cfg := m.Config
	authorID := m.UserID

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot delete admin content: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		err := d.DeleteAdminContentData(ctx, ac, msg.AdminContentID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to delete admin content: %v", err),
			}
		}

		return AdminContentDeletedMsg{AdminContentID: msg.AdminContentID}
	}
}

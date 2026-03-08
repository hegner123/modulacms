package tui

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

// DeleteAdminDatatypeContext stores context for deleting an admin datatype.
type DeleteAdminDatatypeContext struct {
	AdminDatatypeID types.AdminDatatypeID
	Label           string
}

// DeleteAdminFieldContext stores context for deleting an admin field.
type DeleteAdminFieldContext struct {
	AdminFieldID    types.AdminFieldID
	AdminDatatypeID types.AdminDatatypeID
	Label           string
}

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
			Slug:   validSlug,
			Title:  msg.Title,
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
			Slug:   validSlug,
			Title:  msg.Title,
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

		var parentID types.NullableAdminDatatypeID
		if msg.ParentID != "" {
			parentID = types.NullableAdminDatatypeID{
				ID:    types.AdminDatatypeID(msg.ParentID),
				Valid: true,
			}
		}

		// Determine next sort order
		maxSort, sortErr := d.GetMaxAdminDatatypeSortOrder(parentID)
		if sortErr != nil {
			maxSort = -1
		}

		params := db.CreateAdminDatatypeParams{
			ParentID:     parentID,
			SortOrder:    maxSort + 1,
			Name:         msg.Name,
			Label:        msg.Label,
			Type:         dtype,
			AuthorID:     authorID,
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
			SortOrder:       existing.SortOrder,
			Name:            msg.Name,
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

		// Delete the datatype (fields with parent_id referencing it are set to NULL by FK cascade)
		err := d.DeleteAdminDatatype(ctx, ac, msg.AdminDatatypeID)
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
			ParentID:   types.NullableAdminDatatypeID{ID: msg.AdminDatatypeID, Valid: true},
			Name:       msg.Name,
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

		// Derive parent datatype from the fetched field record
		var adminDatatypeID types.AdminDatatypeID
		if existing.ParentID.Valid {
			adminDatatypeID = existing.ParentID.ID
		}

		fieldTypeStr := msg.Type
		if fieldTypeStr == "" {
			fieldTypeStr = string(existing.Type)
		}
		fieldType := types.FieldType(fieldTypeStr)

		params := db.UpdateAdminFieldParams{
			ParentID:     existing.ParentID,
			Name:         msg.Name,
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

		// Fetch field first to determine parent datatype ID for refresh
		existing, err := d.GetAdminField(msg.AdminFieldID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to get admin field for delete: %v", err),
			}
		}
		adminDatatypeID := existing.ParentID.ID

		// Delete the field (parent_id FK on admin_datatypes is handled by cascade)
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
// FIELD TYPE HANDLERS
// =============================================================================

// DeleteFieldTypeContext stores context for deleting a field type.
type DeleteFieldTypeContext struct {
	FieldTypeID types.FieldTypeID
	Label       string
}

// HandleCreateFieldTypeFromDialog processes the field type creation request.
func (m Model) HandleCreateFieldTypeFromDialog(msg CreateFieldTypeFromDialogRequestMsg) tea.Cmd {
	authorID := m.UserID
	cfg := m.Config

	if authorID.IsZero() {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create field type: no user is logged in",
			}
		}
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create field type: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		params := db.CreateFieldTypeParams{
			Type:  msg.Type,
			Label: msg.Label,
		}

		ft, err := d.CreateFieldType(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to create field type: %v", err),
			}
		}

		return FieldTypeCreatedFromDialogMsg{
			FieldTypeID: ft.FieldTypeID,
			Type:        ft.Type,
			Label:       ft.Label,
		}
	}
}

// HandleUpdateFieldTypeFromDialog processes the field type update request.
func (m Model) HandleUpdateFieldTypeFromDialog(msg UpdateFieldTypeFromDialogRequestMsg) tea.Cmd {
	cfg := m.Config
	authorID := m.UserID

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot update field type: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		// STEP 1: Fetch existing record (Golden Rule)
		fieldTypeID := types.FieldTypeID(msg.FieldTypeID)
		existing, err := d.GetFieldType(fieldTypeID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to get field type for update: %v", err),
			}
		}

		// STEP 2: Build params -- set EVERY field
		// For field_types, there are only Type and Label (no timestamps, no FKs).
		// Changed fields use new values; unchanged fields use existing values.
		updateType := msg.Type
		if updateType == "" {
			updateType = existing.Type
		}
		updateLabel := msg.Label
		if updateLabel == "" {
			updateLabel = existing.Label
		}

		params := db.UpdateFieldTypeParams{
			Type:        updateType,
			Label:       updateLabel,
			FieldTypeID: fieldTypeID,
		}

		// STEP 3: Execute update
		_, err = d.UpdateFieldType(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update field type: %v", err),
			}
		}

		return FieldTypeUpdatedFromDialogMsg{
			FieldTypeID: fieldTypeID,
			Type:        updateType,
			Label:       updateLabel,
		}
	}
}

// HandleDeleteFieldType processes the field type deletion request.
func (m Model) HandleDeleteFieldType(msg DeleteFieldTypeRequestMsg) tea.Cmd {
	cfg := m.Config
	authorID := m.UserID

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot delete field type: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		err := d.DeleteFieldType(ctx, ac, msg.FieldTypeID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to delete field type: %v", err),
			}
		}

		return FieldTypeDeletedMsg{FieldTypeID: msg.FieldTypeID}
	}
}

// =============================================================================
// ADMIN FIELD TYPE HANDLERS
// =============================================================================

// DeleteAdminFieldTypeContext stores context for deleting an admin field type.
type DeleteAdminFieldTypeContext struct {
	AdminFieldTypeID types.AdminFieldTypeID
	Label            string
}

// HandleCreateAdminFieldTypeFromDialog processes the admin field type creation request.
func (m Model) HandleCreateAdminFieldTypeFromDialog(msg CreateAdminFieldTypeFromDialogRequestMsg) tea.Cmd {
	authorID := m.UserID
	cfg := m.Config

	if authorID.IsZero() {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create admin field type: no user is logged in",
			}
		}
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create admin field type: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		params := db.CreateAdminFieldTypeParams{
			Type:  msg.Type,
			Label: msg.Label,
		}

		ft, err := d.CreateAdminFieldType(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to create admin field type: %v", err),
			}
		}

		return AdminFieldTypeCreatedFromDialogMsg{
			AdminFieldTypeID: ft.AdminFieldTypeID,
			Type:             ft.Type,
			Label:            ft.Label,
		}
	}
}

// HandleUpdateAdminFieldTypeFromDialog processes the admin field type update request.
func (m Model) HandleUpdateAdminFieldTypeFromDialog(msg UpdateAdminFieldTypeFromDialogRequestMsg) tea.Cmd {
	cfg := m.Config
	authorID := m.UserID

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot update admin field type: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		// STEP 1: Fetch existing record (Golden Rule)
		adminFieldTypeID := types.AdminFieldTypeID(msg.AdminFieldTypeID)
		existing, err := d.GetAdminFieldType(adminFieldTypeID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to get admin field type for update: %v", err),
			}
		}

		// STEP 2: Build params -- set EVERY field
		// For admin_field_types, there are only Type and Label (no timestamps, no FKs).
		updateType := msg.Type
		if updateType == "" {
			updateType = existing.Type
		}
		updateLabel := msg.Label
		if updateLabel == "" {
			updateLabel = existing.Label
		}

		params := db.UpdateAdminFieldTypeParams{
			Type:             updateType,
			Label:            updateLabel,
			AdminFieldTypeID: adminFieldTypeID,
		}

		// STEP 3: Execute update
		_, err = d.UpdateAdminFieldType(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update admin field type: %v", err),
			}
		}

		return AdminFieldTypeUpdatedFromDialogMsg{
			AdminFieldTypeID: adminFieldTypeID,
			Type:             updateType,
			Label:            updateLabel,
		}
	}
}

// HandleDeleteAdminFieldType processes the admin field type deletion request.
func (m Model) HandleDeleteAdminFieldType(msg DeleteAdminFieldTypeRequestMsg) tea.Cmd {
	cfg := m.Config
	authorID := m.UserID

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot delete admin field type: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		err := d.DeleteAdminFieldType(ctx, ac, msg.AdminFieldTypeID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to delete admin field type: %v", err),
			}
		}

		return AdminFieldTypeDeletedMsg{AdminFieldTypeID: msg.AdminFieldTypeID}
	}
}

// HandleCreateAdminRouteWithContent processes the admin route with content creation request.
func (m Model) HandleCreateAdminRouteWithContent(msg CreateAdminRouteWithContentRequestMsg) tea.Cmd {
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
		adminDatatypeID := types.AdminDatatypeID(msg.AdminDatatypeID)

		var adminRouteID types.AdminRouteID
		var routeSlug string

		if msg.Slug != "" {
			validSlug := types.Slugify(msg.Slug)
			if err := validSlug.Validate(); err != nil {
				return ActionResultMsg{
					Title:   "Invalid Slug",
					Message: fmt.Sprintf("Could not create admin route: %v", err),
				}
			}

			existing, _ := d.GetAdminRoute(validSlug)
			if existing != nil {
				return ActionResultMsg{
					Title:   "Duplicate Slug",
					Message: fmt.Sprintf("An admin route with slug %q already exists", validSlug),
				}
			}

			route, routeErr := d.CreateAdminRoute(ctx, ac, db.CreateAdminRouteParams{
				Slug:   validSlug,
				Title:  msg.Title,
				Status: 1,
				AuthorID: types.NullableUserID{
					ID:    authorID,
					Valid: true,
				},
				DateCreated:  types.TimestampNow(),
				DateModified: types.TimestampNow(),
			})
			if routeErr != nil {
				return ActionResultMsg{
					Title:   "Error",
					Message: fmt.Sprintf("Failed to create admin route: %v", routeErr),
				}
			}
			if route.AdminRouteID.IsZero() {
				return ActionResultMsg{
					Title:   "Error",
					Message: "Failed to create admin route in database",
				}
			}
			adminRouteID = route.AdminRouteID
			routeSlug = string(route.Slug)
		}

		contentParams := db.CreateAdminContentDataParams{
			AdminRouteID: types.NullableAdminRouteID{
				ID:    adminRouteID,
				Valid: !adminRouteID.IsZero(),
			},
			AdminDatatypeID: types.NullableAdminDatatypeID{
				ID:    adminDatatypeID,
				Valid: true,
			},
			AuthorID:     authorID,
			Status:       types.ContentStatusDraft,
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		contentData, contentErr := d.CreateAdminContentData(ctx, ac, contentParams)
		if contentErr != nil || contentData.AdminContentDataID.IsZero() {
			if !adminRouteID.IsZero() {
				return ActionResultMsg{
					Title:   "Warning",
					Message: fmt.Sprintf("Admin route created but failed to create initial content. Route: %s", msg.Title),
				}
			}
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to create admin content: %v", contentErr),
			}
		}

		return AdminRouteWithContentCreatedMsg{
			AdminRouteID:       adminRouteID,
			AdminContentDataID: contentData.AdminContentDataID,
			AdminDatatypeID:    adminDatatypeID,
			Title:              msg.Title,
			Slug:               routeSlug,
		}
	}
}

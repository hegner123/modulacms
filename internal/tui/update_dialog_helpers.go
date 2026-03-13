package tui

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"

	tea "charm.land/bubbletea/v2"
)

// =============================================================================
// CONTEXT STRUCTS
// =============================================================================

// InitializeRouteContentContext stores context for route content initialization.
type InitializeRouteContentContext struct {
	Route      db.Routes
	DatatypeID string
}

// DeleteContentContext stores context for a content deletion operation.
// AdminMode selects admin vs regular delete flow.
type DeleteContentContext struct {
	ContentID string
	RouteID   string
	AdminMode bool
}

// RestoreBackupContext stores context for a backup restore operation.
type RestoreBackupContext struct {
	Path string
}

// PublishContentContext stores context for a publish/unpublish confirmation dialog.
// AdminMode selects admin vs regular publish flow.
type PublishContentContext struct {
	ContentID types.ContentID
	RouteID   types.RouteID
	AdminMode bool
}

// RestoreVersionContext stores context for a version restore confirmation dialog.
// AdminMode selects admin vs regular restore flow.
type RestoreVersionContext struct {
	ContentID types.ContentID
	VersionID types.ContentVersionID
	RouteID   types.RouteID
	AdminMode bool
}

// ApprovePluginContext stores context for a plugin approval confirmation dialog.
type ApprovePluginContext struct {
	PluginName string
}

// editSingleFieldCtx stores context for the single-field edit dialog.
type editSingleFieldCtx struct {
	ContentFieldID types.ContentFieldID
	ContentID      types.ContentID
	FieldID        types.FieldID
	RouteID        types.RouteID
	DatatypeID     types.NullableDatatypeID
}

// addContentFieldCtx stores context for the add content field operation.
type addContentFieldCtx struct {
	ContentID  types.ContentID
	RouteID    types.RouteID
	DatatypeID types.NullableDatatypeID
}

// =============================================================================
// CREATE DATATYPE HANDLER
// =============================================================================

// HandleCreateDatatypeFromDialog processes the datatype creation request.
func (m Model) HandleCreateDatatypeFromDialog(msg CreateDatatypeFromDialogRequestMsg) tea.Cmd {
	// Capture values from model for use in closure
	authorID := m.UserID
	cfg := m.Config

	// Validate that we have a user ID (required by database constraint)
	if authorID.IsZero() {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create datatype: no user is logged in",
			}
		}
	}

	// Validate config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create datatype: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		dtype := msg.Type

		// Prepare parent ID
		var parentID types.NullableDatatypeID
		if msg.ParentID != "" {
			parentID = types.NullableDatatypeID{
				ID:    types.DatatypeID(msg.ParentID),
				Valid: true,
			}
		}

		// Determine next sort order
		maxSort, sortErr := d.GetMaxDatatypeSortOrder(parentID)
		if sortErr != nil {
			maxSort = -1
		}

		// Create the datatype
		params := db.CreateDatatypeParams{
			DatatypeID:   types.NewDatatypeID(),
			ParentID:     parentID,
			SortOrder:    maxSort + 1,
			Name:         msg.Name,
			Label:        msg.Label,
			Type:         dtype,
			AuthorID:     authorID,
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		dt, err := d.CreateDatatype(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to create datatype: %v", err),
			}
		}
		return DatatypeCreatedFromDialogMsg{
			DatatypeID: dt.DatatypeID,
			Label:      dt.Label,
		}
	}
}

// =============================================================================
// CREATE FIELD HANDLER
// =============================================================================

// HandleCreateFieldFromDialog processes the field creation request.
func (m Model) HandleCreateFieldFromDialog(msg CreateFieldFromDialogRequestMsg) tea.Cmd {
	// Capture values from model for use in closure
	authorID := m.UserID
	cfg := m.Config

	// Validate that we have a user ID
	if authorID.IsZero() {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create field: no user is logged in",
			}
		}
	}

	// Validate config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create field: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		// Verify the author user exists in the database before creating the field
		authorUser, userErr := d.GetUser(authorID)
		if userErr != nil || authorUser == nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Cannot create field: author user %s not found in database (run --install to bootstrap)", authorID),
			}
		}

		// Prepare the field type - default to "text" if empty
		fieldTypeStr := msg.Type
		if fieldTypeStr == "" {
			fieldTypeStr = "text"
		}
		fieldType := types.FieldType(fieldTypeStr)

		// Set author ID
		nullableAuthorID := types.NullableUserID{
			ID:    authorID,
			Valid: true,
		}

		// Determine sort order for new field
		parentID := types.NullableDatatypeID{ID: msg.DatatypeID, Valid: true}
		maxSort, sortErr := d.GetMaxSortOrderByParentID(parentID)
		if sortErr != nil {
			maxSort = -1
		}

		// Create the field with parent_id and sort_order set directly
		fieldID := types.NewFieldID()
		fieldParams := db.CreateFieldParams{
			FieldID:      fieldID,
			ParentID:     parentID,
			SortOrder:    maxSort + 1,
			Name:         msg.Name,
			Label:        msg.Label,
			Data:         "",
			Validation:   types.EmptyJSON,
			UIConfig:     types.EmptyJSON,
			Type:         fieldType,
			AuthorID:     nullableAuthorID,
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		field, err := d.CreateField(ctx, ac, fieldParams)
		if err != nil || field.FieldID.IsZero() {
			errMsg := "Failed to create field in database"
			if err != nil {
				errMsg = fmt.Sprintf("Failed to create field: %v", err)
			}
			return ActionResultMsg{
				Title:   "Error",
				Message: errMsg,
			}
		}

		return FieldCreatedFromDialogMsg{
			FieldID:    field.FieldID,
			DatatypeID: msg.DatatypeID,
			Label:      field.Label,
		}
	}
}

// =============================================================================
// CREATE ROUTE HANDLER
// =============================================================================

// HandleCreateRouteFromDialog processes the route creation request.
func (m Model) HandleCreateRouteFromDialog(msg CreateRouteFromDialogRequestMsg) tea.Cmd {
	// Capture values from model for use in closure
	authorID := m.UserID
	cfg := m.Config

	// Validate that we have a user ID (required by database constraint)
	if authorID.IsZero() {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create route: no user is logged in",
			}
		}
	}

	// Validate config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create route: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

		// Prepare the slug - use Slugify to ensure valid format
		slug := msg.Slug
		if slug == "" {
			slug = msg.Title
		}
		validSlug := types.Slugify(slug)

		// Validate the slug
		if err := validSlug.Validate(); err != nil {
			return ActionResultMsg{
				Title:   "Invalid Slug",
				Message: fmt.Sprintf("Could not create route: %v", err),
			}
		}

		// Check if slug already exists
		existingID, _ := d.GetRouteID(string(validSlug))
		if existingID != nil {
			return ActionResultMsg{
				Title:   "Duplicate Slug",
				Message: fmt.Sprintf("A route with slug %q already exists", validSlug),
			}
		}

		// Create the route
		params := db.CreateRouteParams{
			Slug:   validSlug,
			Title:  msg.Title,
			Status: 1, // Active by default
			AuthorID: types.NullableUserID{
				ID:    authorID,
				Valid: true,
			},
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		route, err := d.CreateRoute(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to create route: %v", err),
			}
		}
		if route.RouteID.IsZero() {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Failed to create route in database",
			}
		}

		return RouteCreatedFromDialogMsg{
			RouteID: route.RouteID,
			Title:   route.Title,
			Slug:    string(route.Slug),
		}
	}
}

// =============================================================================
// UPDATE DATATYPE HANDLER
// =============================================================================

// HandleUpdateDatatypeFromDialog processes the datatype update request.
func (m Model) HandleUpdateDatatypeFromDialog(msg UpdateDatatypeFromDialogRequestMsg) tea.Cmd {
	cfg := m.Config

	// Validate config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot update datatype: configuration not loaded",
			}
		}
	}

	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		// Fetch existing datatype to preserve unchanged values
		datatypeID := types.DatatypeID(msg.DatatypeID)
		existing, err := d.GetDatatype(datatypeID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to get datatype for update: %v", err),
			}
		}

		// Prepare the type - default to existing type if empty
		dtype := msg.Type
		if dtype == "" {
			dtype = existing.Type
		}

		// Prepare parent ID - use provided value or preserve existing
		parentID := existing.ParentID
		if msg.ParentID != "" {
			parentID = types.NullableDatatypeID{
				ID:    types.DatatypeID(msg.ParentID),
				Valid: true,
			}
		}

		// Update only changed fields; preserve author and date_created
		params := db.UpdateDatatypeParams{
			DatatypeID:   datatypeID,
			ParentID:     parentID,
			SortOrder:    existing.SortOrder,
			Name:         msg.Name,
			Label:        msg.Label,
			Type:         dtype,
			AuthorID:     existing.AuthorID,
			DateCreated:  existing.DateCreated,
			DateModified: types.TimestampNow(),
		}

		_, err = d.UpdateDatatype(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update datatype: %v", err),
			}
		}

		return DatatypeUpdatedFromDialogMsg{
			DatatypeID: datatypeID,
			Label:      msg.Label,
		}
	}
}

// =============================================================================
// UPDATE FIELD HANDLER
// =============================================================================

// HandleUpdateFieldFromDialog processes the field update request.
func (m Model) HandleUpdateFieldFromDialog(msg UpdateFieldFromDialogRequestMsg) tea.Cmd {
	cfg := m.Config

	// Validate config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot update field: configuration not loaded",
			}
		}
	}

	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		// Fetch existing field to preserve unchanged values
		fieldID := types.FieldID(msg.FieldID)
		existing, err := d.GetField(fieldID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to get field for update: %v", err),
			}
		}

		// Derive parent datatype ID from the fetched field record
		datatypeID := existing.ParentID.ID

		// Prepare the field type - default to existing type if empty
		fieldTypeStr := msg.Type
		if fieldTypeStr == "" {
			fieldTypeStr = string(existing.Type)
		}
		fieldType := types.FieldType(fieldTypeStr)

		// Update only name, label, type, and date_modified; preserve everything else
		params := db.UpdateFieldParams{
			FieldID:      fieldID,
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
		}

		_, err = d.UpdateField(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update field: %v", err),
			}
		}

		return FieldUpdatedFromDialogMsg{
			FieldID:    fieldID,
			DatatypeID: datatypeID,
			Label:      msg.Label,
		}
	}
}

// =============================================================================
// UPDATE FIELD UICONFIG HANDLER
// =============================================================================

// HandleUpdateFieldUIConfig processes a field UIConfig update request.
func (m Model) HandleUpdateFieldUIConfig(msg UpdateFieldUIConfigRequestMsg) tea.Cmd {
	cfg := m.Config

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot update field UIConfig: configuration not loaded",
			}
		}
	}

	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		fieldID := types.FieldID(msg.FieldID)
		existing, err := d.GetField(fieldID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to get field for UIConfig update: %v", err),
			}
		}

		// Derive parent datatype ID from the fetched field record
		datatypeID := existing.ParentID.ID

		params := db.UpdateFieldParams{
			FieldID:      fieldID,
			ParentID:     existing.ParentID,
			Label:        existing.Label,
			Data:         existing.Data,
			Validation:   existing.Validation,
			UIConfig:     msg.UIConfigJSON,
			Type:         existing.Type,
			AuthorID:     existing.AuthorID,
			DateCreated:  existing.DateCreated,
			DateModified: types.TimestampNow(),
		}

		_, err = d.UpdateField(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update field UIConfig: %v", err),
			}
		}

		return FieldUIConfigUpdatedMsg{
			FieldID:    fieldID,
			DatatypeID: datatypeID,
		}
	}
}

// =============================================================================
// UPDATE ROUTE HANDLER
// =============================================================================

// HandleUpdateRouteFromDialog processes the route update request.
func (m Model) HandleUpdateRouteFromDialog(msg UpdateRouteFromDialogRequestMsg) tea.Cmd {
	// Capture values from model for use in closure
	authorID := m.UserID
	cfg := m.Config

	// Validate config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot update route: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

		// Get the existing route to preserve its original slug for the WHERE clause
		existingRoute, err := d.GetRoute(types.RouteID(msg.RouteID))
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Route not found: %v", err),
			}
		}

		// Prepare the slug - use Slugify to ensure valid format
		slug := msg.Slug
		if slug == "" {
			slug = msg.Title
		}
		validSlug := types.Slugify(slug)

		// Validate the slug
		if err := validSlug.Validate(); err != nil {
			return ActionResultMsg{
				Title:   "Invalid Slug",
				Message: fmt.Sprintf("Could not update route: %v", err),
			}
		}

		// Check if new slug already exists (unless it's the same route)
		if validSlug != existingRoute.Slug {
			existingID, _ := d.GetRouteID(string(validSlug))
			if existingID != nil {
				return ActionResultMsg{
					Title:   "Duplicate Slug",
					Message: fmt.Sprintf("A route with slug %q already exists", validSlug),
				}
			}
		}

		// Set author ID
		nullableAuthorID := types.NullableUserID{
			ID:    authorID,
			Valid: !authorID.IsZero(),
		}

		// Update the route
		// Note: UpdateRouteParams uses Slug_2 for the WHERE clause (original slug)
		params := db.UpdateRouteParams{
			Slug:         validSlug,
			Title:        msg.Title,
			Status:       existingRoute.Status,
			AuthorID:     nullableAuthorID,
			DateCreated:  existingRoute.DateCreated,
			DateModified: types.TimestampNow(),
			Slug_2:       existingRoute.Slug, // Original slug for WHERE clause
		}

		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		_, err = d.UpdateRoute(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update route: %v", err),
			}
		}

		return RouteUpdatedFromDialogMsg{
			RouteID: types.RouteID(msg.RouteID),
			Title:   msg.Title,
			Slug:    string(validSlug),
		}
	}
}

// =============================================================================
// CREATE ROUTE WITH CONTENT HANDLER
// =============================================================================

// HandleCreateRouteWithContent processes the route with content creation request.
func (m Model) HandleCreateRouteWithContent(msg CreateRouteWithContentRequestMsg) tea.Cmd {
	// Capture values from model for use in closure
	authorID := m.UserID
	cfg := m.Config

	// Validate that we have a user ID (required by database constraint)
	if authorID.IsZero() {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create route: no user is logged in",
			}
		}
	}

	// Validate config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create route: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)
		datatypeID := types.DatatypeID(msg.DatatypeID)

		var routeID types.RouteID
		var routeSlug string

		// If slug is provided, create a route; otherwise create content without a route
		if msg.Slug != "" {
			validSlug := types.Slugify(msg.Slug)

			if err := validSlug.Validate(); err != nil {
				return ActionResultMsg{
					Title:   "Invalid Slug",
					Message: fmt.Sprintf("Could not create route: %v", err),
				}
			}

			existingID, _ := d.GetRouteID(string(validSlug))
			if existingID != nil {
				return ActionResultMsg{
					Title:   "Duplicate Slug",
					Message: fmt.Sprintf("A route with slug %q already exists", validSlug),
				}
			}

			route, routeErr := d.CreateRoute(ctx, ac, db.CreateRouteParams{
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
					Message: fmt.Sprintf("Failed to create route: %v", routeErr),
				}
			}
			if route.RouteID.IsZero() {
				return ActionResultMsg{
					Title:   "Error",
					Message: "Failed to create route in database",
				}
			}
			routeID = route.RouteID
			routeSlug = string(route.Slug)
		}

		contentParams := db.CreateContentDataParams{
			RouteID: types.NullableRouteID{
				ID:    routeID,
				Valid: !routeID.IsZero(),
			},
			DatatypeID: types.NullableDatatypeID{
				ID:    datatypeID,
				Valid: true,
			},
			AuthorID:     authorID,
			Status:       types.ContentStatusDraft,
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		contentData, contentErr := d.CreateContentData(ctx, ac, contentParams)
		if contentErr != nil || contentData.ContentDataID.IsZero() {
			if !routeID.IsZero() {
				return ActionResultMsg{
					Title:   "Warning",
					Message: fmt.Sprintf("Route created but failed to create initial content. Route: %s", msg.Title),
				}
			}
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to create content: %v", contentErr),
			}
		}

		// Set root_id to self for the new root content node
		_, rootUpdateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: contentData.ContentDataID,
			RootID:        types.NullableContentID{ID: contentData.ContentDataID, Valid: true},
			ParentID:      contentData.ParentID,
			FirstChildID:  contentData.FirstChildID,
			NextSiblingID: contentData.NextSiblingID,
			PrevSiblingID: contentData.PrevSiblingID,
			RouteID:       contentData.RouteID,
			DatatypeID:    contentData.DatatypeID,
			AuthorID:      contentData.AuthorID,
			Status:        contentData.Status,
			DateCreated:   contentData.DateCreated,
			DateModified:  types.TimestampNow(),
		})
		if rootUpdateErr != nil {
			utility.DefaultLogger.Ferror("Failed to set root_id on new root content", rootUpdateErr)
		}

		return RouteWithContentCreatedMsg{
			RouteID:       routeID,
			ContentDataID: contentData.ContentDataID,
			DatatypeID:    datatypeID,
			Title:         msg.Title,
			Slug:          routeSlug,
		}
	}
}

// =============================================================================
// INITIALIZE ROUTE CONTENT HANDLER
// =============================================================================

// HandleInitializeRouteContent processes the route content initialization request.
func (m Model) HandleInitializeRouteContent(msg InitializeRouteContentRequestMsg) tea.Cmd {
	// Capture values from model for use in closure
	authorID := m.UserID
	cfg := m.Config

	// Validate config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot initialize content: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

		// Get the route to include its title in the response
		route, err := d.GetRoute(msg.RouteID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Route not found: %v", err),
			}
		}

		// Create initial content data for this route
		datatypeID := types.DatatypeID(msg.DatatypeID)
		contentParams := db.CreateContentDataParams{
			RouteID: types.NullableRouteID{
				ID:    msg.RouteID,
				Valid: true,
			},
			DatatypeID: types.NullableDatatypeID{
				ID:    datatypeID,
				Valid: true,
			},
			AuthorID:     authorID,
			Status:       types.ContentStatusDraft,
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		contentData, contentErr := d.CreateContentData(ctx, ac, contentParams)
		if contentErr != nil || contentData.ContentDataID.IsZero() {
			errMsg := "Failed to create content in database"
			if contentErr != nil {
				errMsg = fmt.Sprintf("Failed to create content: %v", contentErr)
			}
			return ActionResultMsg{
				Title:   "Error",
				Message: errMsg,
			}
		}

		// Set root_id to self for the new root content node
		_, rootUpdateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: contentData.ContentDataID,
			RootID:        types.NullableContentID{ID: contentData.ContentDataID, Valid: true},
			ParentID:      contentData.ParentID,
			FirstChildID:  contentData.FirstChildID,
			NextSiblingID: contentData.NextSiblingID,
			PrevSiblingID: contentData.PrevSiblingID,
			RouteID:       contentData.RouteID,
			DatatypeID:    contentData.DatatypeID,
			AuthorID:      contentData.AuthorID,
			Status:        contentData.Status,
			DateCreated:   contentData.DateCreated,
			DateModified:  types.TimestampNow(),
		})
		if rootUpdateErr != nil {
			utility.DefaultLogger.Ferror("Failed to set root_id on new root content", rootUpdateErr)
		}

		return RouteContentInitializedMsg{
			RouteID:       msg.RouteID,
			ContentDataID: contentData.ContentDataID,
			DatatypeID:    datatypeID,
			Title:         route.Title,
		}
	}
}

// handleCrudResultMsg handles CRUD result messages that trigger data refreshes
// after successful create/update/delete operations.
// Returns (model, cmd, handled). If handled is false, the caller should
// continue dispatching through the main switch.
func (m Model) handleCrudResultMsg(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case UserCreatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("User created: %s", msg.Username)),
			UsersFetchCmd(),
		), true
	case UserUpdatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("User updated: %s", msg.Username)),
			UsersFetchCmd(),
		), true
	case UserDeletedMsg:
		newModel := m
		newModel.Cursor = 0
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("User deleted: %s", msg.UserID)),
			UsersFetchCmd(),
		), true
	case ContentCreatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			ShowDialog(
				"Success",
				fmt.Sprintf("✓ Content created with %d fields", msg.FieldCount),
				false,
			),
			LogMessageCmd(fmt.Sprintf("Content created: ID=%s, DatatypeID=%s", msg.ContentID, msg.DatatypeID)),
			ReloadContentTreeCmd(m.Config, msg.RouteID),
		), true
	case ContentUpdatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			ShowDialog(
				"Success",
				fmt.Sprintf("✓ Content updated (%d fields)", msg.UpdatedCount),
				false,
			),
			LogMessageCmd(fmt.Sprintf("Content updated: ID=%s, DatatypeID=%s", msg.ContentID, msg.DatatypeID)),
			ReloadContentTreeCmd(m.Config, msg.RouteID),
		), true
	case DatatypeCreatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Datatype created: %s", msg.Label)),
			AllDatatypesFetchCmd(),
		), true
	case FieldCreatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Field created: %s", msg.Label)),
			DatatypeFieldsFetchCmd(msg.DatatypeID),
		), true
	case RouteCreatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Route created: %s", msg.Title)),
			RoutesFetchCmd(),
		), true
	case DatatypeUpdatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Datatype updated: %s", msg.Label)),
			AllDatatypesFetchCmd(),
		), true
	case FieldUpdatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Field updated: %s", msg.Label)),
			DatatypeFieldsFetchCmd(msg.DatatypeID),
		), true
	case RouteUpdatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Route updated: %s", msg.Title)),
			RoutesFetchCmd(),
		), true
	case RouteWithContentCreatedMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Route created with content: %s (ContentID: %s)", msg.Title, msg.ContentDataID)),
			RoutesByDatatypeFetchCmd(msg.DatatypeID),
		), true
	case RouteContentInitializedMsg:
		newModel := m
		newModel.PageRouteId = msg.RouteID
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Content initialized for route: %s", msg.Title)),
			ReloadContentTreeCmd(m.Config, msg.RouteID),
		), true
	case DatatypeDeletedMsg:
		newModel := m
		newModel.Cursor = 0
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Datatype deleted: %s", msg.DatatypeID)),
			AllDatatypesFetchCmd(),
		), true
	case FieldDeletedMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Field deleted: %s", msg.FieldID)),
			DatatypeFieldsFetchCmd(msg.DatatypeID),
		), true
	case RouteDeletedMsg:
		newModel := m
		newModel.Cursor = 0
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Route deleted: %s", msg.RouteID)),
			RoutesFetchCmd(),
		), true
	case MediaDeletedMsg:
		newModel := m
		newModel.Cursor = 0
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Media deleted: %s", msg.MediaID)),
			MediaFetchCmd(),
		), true
	default:
		return m, nil, false
	}
}

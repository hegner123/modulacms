package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// CreateContentWithFields performs atomic content creation using typed DbDriver methods.
// Creates ContentData first, then uses the returned ID to create associated ContentFields.
// This solves the ID-passing problem that the generic query builder pattern couldn't handle.
func (m Model) CreateContentWithFields(
	c *config.Config,
	datatypeID types.DatatypeID,
	routeID types.RouteID,
	authorID types.UserID,
	fieldValues map[types.FieldID]string,
) tea.Cmd {
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}
	cfg := c
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		// Debug logging
		logger.Finfo(fmt.Sprintf("Creating ContentData: DatatypeID=%s, RouteID=%s, AuthorID=%s", datatypeID, routeID, authorID))

		// Step 1: Create ContentData using typed DbDriver method
		// RootID is not set here because root nodes need their own ID, which
		// is not known until after creation. It is set in Step 1.5 below.
		contentData, err := d.CreateContentData(ctx, ac, db.CreateContentDataParams{
			DatatypeID:    types.NullableDatatypeID{ID: datatypeID, Valid: true},
			RouteID:       types.NullableRouteID{ID: routeID, Valid: true},
			AuthorID:      authorID,
			Status:        types.ContentStatusDraft,
			DateCreated:   types.TimestampNow(),
			DateModified:  types.TimestampNow(),
			ParentID:      types.NullableContentID{}, // NULL - no parent initially
			FirstChildID:  types.NullableContentID{}, // NULL - no children initially
			NextSiblingID: types.NullableContentID{}, // NULL - no siblings initially
			PrevSiblingID: types.NullableContentID{}, // NULL - no siblings initially
		})
		if err != nil {
			return DbErrMsg{
				Error: fmt.Errorf("failed to create content data: %w", err),
			}
		}

		// Check if creation succeeded
		if contentData.ContentDataID.IsZero() {
			return DbErrMsg{
				Error: fmt.Errorf("failed to create content data"),
			}
		}

		// Step 1.5: Set root_id on the new content.
		// This function creates root-level content (no parent), so root_id = self.
		rootID := types.NullableContentID{ID: contentData.ContentDataID, Valid: true}
		_, rootUpdateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: contentData.ContentDataID,
			RootID:        rootID,
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
			logger.Ferror("Failed to set root_id on new root content", rootUpdateErr)
		}

		// Step 2: Create ContentFields for every field defined on the datatype.
		// Uses the canonical field list so all fields get a content_field row,
		// matching the API/admin panel behavior.
		var failedFields []types.FieldID
		createdFields := 0

		allFields, fieldListErr := d.ListFieldsByDatatypeID(types.NullableDatatypeID{ID: datatypeID, Valid: true})
		if fieldListErr != nil {
			logger.Ferror("Failed to list datatype fields, falling back to user-provided fields only", fieldListErr)
		}

		if allFields != nil && len(*allFields) > 0 {
			for _, field := range *allFields {
				value := fieldValues[field.FieldID] // "" if not in map

				fieldResult, fieldErr := d.CreateContentField(ctx, ac, db.CreateContentFieldParams{
					ContentDataID: types.NullableContentID{ID: contentData.ContentDataID, Valid: true},
					FieldID:       types.NullableFieldID{ID: field.FieldID, Valid: true},
					FieldValue:    value,
					RootID:        rootID,
					RouteID:       types.NullableRouteID{ID: routeID, Valid: true},
					AuthorID:      authorID,
					DateCreated:   types.TimestampNow(),
					DateModified:  types.TimestampNow(),
				})

				if fieldErr != nil || fieldResult.ContentFieldID.IsZero() {
					failedFields = append(failedFields, field.FieldID)
				} else {
					createdFields++
				}
			}
		} else if fieldListErr != nil {
			// Fallback: use only user-provided values when canonical list unavailable
			for fieldID, value := range fieldValues {
				fieldResult, fieldErr := d.CreateContentField(ctx, ac, db.CreateContentFieldParams{
					ContentDataID: types.NullableContentID{ID: contentData.ContentDataID, Valid: true},
					FieldID:       types.NullableFieldID{ID: fieldID, Valid: true},
					FieldValue:    value,
					RootID:        rootID,
					RouteID:       types.NullableRouteID{ID: routeID, Valid: true},
					AuthorID:      authorID,
					DateCreated:   types.TimestampNow(),
					DateModified:  types.TimestampNow(),
				})

				if fieldErr != nil || fieldResult.ContentFieldID.IsZero() {
					failedFields = append(failedFields, fieldID)
				} else {
					createdFields++
				}
			}
		}

		// Step 3: Return appropriate message based on results
		if len(failedFields) > 0 {
			return ContentCreatedWithErrorsMsg{
				ContentDataID: contentData.ContentDataID,
				RouteID:       routeID,
				CreatedFields: createdFields,
				FailedFields:  failedFields,
			}
		}

		return ContentCreatedMsg{
			ContentID:  contentData.ContentDataID,
			RouteID:    routeID,
			FieldCount: createdFields,
		}
	}
}

// HandleCreateContentFromDialog creates content from dialog values with parent support
func (m Model) HandleCreateContentFromDialog(
	msg CreateContentFromDialogRequestMsg,
	authorID types.UserID,
) tea.Cmd {
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}
	cfg := m.Config
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		// Debug logging
		logger.Finfo(fmt.Sprintf("Creating ContentData from dialog: DatatypeID=%s, RouteID=%s, AuthorID=%s, HasParent=%v",
			msg.DatatypeID, msg.RouteID, authorID, msg.ParentID.Valid))

		// Step 0: Determine root_id before creation.
		// Child nodes inherit root_id from their parent; root nodes set it to
		// self after creation (since the ID is not known until INSERT).
		var rootID types.NullableContentID
		if msg.ParentID.Valid {
			parentData, lookupErr := d.GetContentData(msg.ParentID.ID)
			if lookupErr != nil {
				logger.Ferror("Failed to look up parent content data for root_id", lookupErr)
				// Proceed without root_id rather than aborting the entire creation
			} else if parentData != nil {
				rootID = parentData.RootID
			}
		}

		// Step 1: Create ContentData using typed DbDriver method
		nRouteID := types.NullableRouteID{ID: msg.RouteID, Valid: !msg.RouteID.IsZero()}
		contentData, err := d.CreateContentData(ctx, ac, db.CreateContentDataParams{
			DatatypeID:    types.NullableDatatypeID{ID: msg.DatatypeID, Valid: true},
			RouteID:       nRouteID,
			AuthorID:      authorID,
			Status:        types.ContentStatusDraft,
			DateCreated:   types.TimestampNow(),
			DateModified:  types.TimestampNow(),
			ParentID:      msg.ParentID,
			RootID:        rootID, // Set for child nodes; NULL for root nodes (updated in Step 1.5)
			FirstChildID:  types.NullableContentID{}, // NULL - no children initially
			NextSiblingID: types.NullableContentID{}, // NULL - no siblings initially
			PrevSiblingID: types.NullableContentID{}, // NULL - no siblings initially
		})
		if err != nil {
			return DbErrMsg{
				Error: fmt.Errorf("failed to create content data: %w", err),
			}
		}

		// Check if creation succeeded
		if contentData.ContentDataID.IsZero() {
			return DbErrMsg{
				Error: fmt.Errorf("failed to create content data"),
			}
		}

		// Step 1.5: For root nodes (no parent), set root_id to self.
		// The ID was not available at creation time, so a follow-up update is required.
		if !msg.ParentID.Valid {
			rootID = types.NullableContentID{ID: contentData.ContentDataID, Valid: true}
			_, rootUpdateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: contentData.ContentDataID,
				RootID:        rootID,
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
				logger.Ferror("Failed to set root_id on new root content", rootUpdateErr)
			}
		}

		// Step 2: Create ContentFields for every field defined on the datatype.
		var failedFields []types.FieldID
		createdFields := 0

		allFields, fieldListErr := d.ListFieldsByDatatypeID(types.NullableDatatypeID{ID: msg.DatatypeID, Valid: true})
		if fieldListErr != nil {
			logger.Ferror("Failed to list datatype fields, falling back to user-provided fields only", fieldListErr)
		}

		if allFields != nil && len(*allFields) > 0 {
			for _, field := range *allFields {
				value := msg.FieldValues[field.FieldID] // "" if not in map

				fieldResult, fieldErr := d.CreateContentField(ctx, ac, db.CreateContentFieldParams{
					ContentDataID: types.NullableContentID{ID: contentData.ContentDataID, Valid: true},
					FieldID:       types.NullableFieldID{ID: field.FieldID, Valid: true},
					FieldValue:    value,
					RootID:        rootID,
					RouteID:       nRouteID,
					AuthorID:      authorID,
					DateCreated:   types.TimestampNow(),
					DateModified:  types.TimestampNow(),
				})

				if fieldErr != nil || fieldResult.ContentFieldID.IsZero() {
					failedFields = append(failedFields, field.FieldID)
				} else {
					createdFields++
				}
			}
		} else if fieldListErr != nil {
			// Fallback: use only user-provided values when canonical list unavailable
			for fieldID, value := range msg.FieldValues {
				fieldResult, fieldErr := d.CreateContentField(ctx, ac, db.CreateContentFieldParams{
					ContentDataID: types.NullableContentID{ID: contentData.ContentDataID, Valid: true},
					FieldID:       types.NullableFieldID{ID: fieldID, Valid: true},
					FieldValue:    value,
					RootID:        rootID,
					RouteID:       nRouteID,
					AuthorID:      authorID,
					DateCreated:   types.TimestampNow(),
					DateModified:  types.TimestampNow(),
				})

				if fieldErr != nil || fieldResult.ContentFieldID.IsZero() {
					failedFields = append(failedFields, fieldID)
				} else {
					createdFields++
				}
			}
		}

		// Step 3: Return appropriate message based on results
		if len(failedFields) > 0 {
			return ContentCreatedWithErrorsMsg{
				ContentDataID: contentData.ContentDataID,
				RouteID:       msg.RouteID,
				CreatedFields: createdFields,
				FailedFields:  failedFields,
			}
		}

		return ContentCreatedFromDialogMsg{
			ContentID:  contentData.ContentDataID,
			DatatypeID: msg.DatatypeID,
			RouteID:    msg.RouteID,
			FieldCount: createdFields,
		}
	}
}

// HandleFetchContentForEdit fetches existing content fields and shows edit dialog
func (m Model) HandleFetchContentForEdit(msg FetchContentForEditMsg) tea.Cmd {
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}
	cfg := m.Config
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

		logger.Finfo(fmt.Sprintf("Fetching content fields for edit: ContentID=%s, DatatypeID=%s", msg.ContentID, msg.DatatypeID))

		// Get existing content fields for this content
		contentDataID := types.NullableContentID{ID: msg.ContentID, Valid: true}
		contentFields, err := d.ListContentFieldsByContentData(contentDataID)
		if err != nil {
			logger.Ferror("Failed to fetch content fields for edit", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to fetch content fields: %v", err),
			}
		}

		// Get field definitions for this datatype by parent_id
		fieldList, err := d.ListFieldsByDatatypeID(types.NullableDatatypeID{ID: msg.DatatypeID, Valid: true})
		if err != nil {
			logger.Ferror("Failed to fetch datatype fields", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to fetch field definitions: %v", err),
			}
		}

		// Build a map of existing content field values for quick lookup
		contentFieldMap := make(map[string]db.ContentFields)
		if contentFields != nil {
			for _, cf := range *contentFields {
				if cf.FieldID.Valid {
					contentFieldMap[string(cf.FieldID.ID)] = cf
				}
			}
		}

		// Build the existing fields list ordered by sort_order
		existingFields := make([]ExistingContentField, 0)
		if fieldList != nil {
			for _, field := range *fieldList {
				uc, _ := types.ParseUIConfig(field.UIConfig)

				ef := ExistingContentField{
					FieldID:        field.FieldID,
					Label:          field.Label,
					Type:           string(field.Type),
					Widget:         uc.Widget,
					Placeholder:    uc.Placeholder,
					Value:          "",
					ValidationJSON: field.Validation,
					DataJSON:       field.Data,
					HelpText:       uc.HelpText,
					Hidden:         uc.Hidden,
				}
				// Check if there's an existing value for this field
				if cf, ok := contentFieldMap[string(field.FieldID)]; ok {
					ef.ContentFieldID = cf.ContentFieldID
					ef.Value = cf.FieldValue
				}
				existingFields = append(existingFields, ef)
			}
		}

		logger.Finfo(fmt.Sprintf("Found %d field definitions, %d existing values", len(existingFields), len(contentFieldMap)))

		return ShowEditContentFormDialogMsg{
			Title:          msg.Title,
			ContentID:      msg.ContentID,
			DatatypeID:     msg.DatatypeID,
			RouteID:        msg.RouteID,
			ExistingFields: existingFields,
		}
	}
}

// HandleUpdateContentFromDialog updates existing content fields from dialog values
func (m Model) HandleUpdateContentFromDialog(
	msg UpdateContentFromDialogRequestMsg,
	authorID types.UserID,
) tea.Cmd {
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}
	cfg := m.Config
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		logger.Finfo(fmt.Sprintf("Updating content fields: ContentID=%s, AuthorID=%s, %d fields",
			msg.ContentID, authorID, len(msg.FieldValues)))

		// Get existing content fields to determine if we need to update or create
		contentDataID := types.NullableContentID{ID: msg.ContentID, Valid: true}
		existingFields, err := d.ListContentFieldsByContentData(contentDataID)
		if err != nil {
			logger.Ferror("Failed to fetch existing content fields", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to fetch existing fields: %v", err),
			}
		}

		// Build a map of existing content fields by field_id
		nRouteID := types.NullableRouteID{ID: msg.RouteID, Valid: !msg.RouteID.IsZero()}
		existingMap := make(map[string]db.ContentFields)
		if existingFields != nil {
			for _, cf := range *existingFields {
				if cf.FieldID.Valid {
					existingMap[string(cf.FieldID.ID)] = cf
				}
			}
		}

		updatedCount := 0
		var updateErrors []string

		for fieldID, value := range msg.FieldValues {
			// Check if this field already exists
			if existing, ok := existingMap[string(fieldID)]; ok {
				// Update existing field
				_, err := d.UpdateContentField(ctx, ac, db.UpdateContentFieldParams{
					ContentFieldID: existing.ContentFieldID,
					RouteID:        existing.RouteID,
					ContentDataID:  contentDataID,
					FieldID:        types.NullableFieldID{ID: fieldID, Valid: true},
					FieldValue:     value,
					AuthorID:       authorID,
					DateCreated:    existing.DateCreated,
					DateModified:   types.TimestampNow(),
				})
				if err != nil {
					logger.Ferror(fmt.Sprintf("Failed to update field %s", fieldID), err)
					updateErrors = append(updateErrors, string(fieldID))
				} else {
					updatedCount++
				}
			} else {
				// Create new field (field was added to datatype after content was created)
				fieldResult, fieldErr := d.CreateContentField(ctx, ac, db.CreateContentFieldParams{
					ContentDataID: contentDataID,
					FieldID:       types.NullableFieldID{ID: fieldID, Valid: true},
					FieldValue:    value,
					RouteID:       nRouteID,
					AuthorID:      authorID,
					DateCreated:   types.TimestampNow(),
					DateModified:  types.TimestampNow(),
				})
				if fieldErr != nil || fieldResult.ContentFieldID.IsZero() {
					logger.Ferror(fmt.Sprintf("Failed to create field %s", fieldID), fieldErr)
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

		return ContentUpdatedFromDialogMsg{
			ContentID:    msg.ContentID,
			DatatypeID:   msg.DatatypeID,
			RouteID:      msg.RouteID,
			UpdatedCount: updatedCount,
		}
	}
}

// LoadContentFieldsMsg carries cached content fields for the right panel.
type LoadContentFieldsMsg struct {
	Fields []ContentFieldDisplay
}

// LoadContentFieldsCmd fetches content fields for a specific content node
// and resolves field labels from the datatype field definitions.
func LoadContentFieldsCmd(cfg *config.Config, contentDataID types.ContentID, datatypeID types.NullableDatatypeID) tea.Cmd {
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

		contentID := types.NullableContentID{ID: contentDataID, Valid: true}
		contentFields, err := d.ListContentFieldsByContentData(contentID)
		if err != nil {
			return LoadContentFieldsMsg{Fields: nil}
		}

		// Fetch field definitions by parent datatype ID
		var fieldDefs *[]db.Fields
		if datatypeID.Valid {
			fieldDefs, err = d.ListFieldsByDatatypeID(datatypeID)
			if err != nil {
				fieldDefs = nil
			}
		}

		// Build content field value map: field_id -> ContentFields
		cfMap := make(map[string]db.ContentFields)
		if contentFields != nil {
			for _, cf := range *contentFields {
				if cf.FieldID.Valid {
					cfMap[string(cf.FieldID.ID)] = cf
				}
			}
		}

		// Build result ordered by sort_order from field definitions
		var result []ContentFieldDisplay
		if fieldDefs != nil {
			result = make([]ContentFieldDisplay, 0, len(*fieldDefs))
			for _, field := range *fieldDefs {
				display := ContentFieldDisplay{
					FieldID:        field.FieldID,
					Label:          field.Label,
					Type:           string(field.Type),
					ValidationJSON: field.Validation,
					DataJSON:       field.Data,
				}
				if cf, ok := cfMap[string(field.FieldID)]; ok {
					display.ContentFieldID = cf.ContentFieldID
					display.Value = cf.FieldValue
				}
				result = append(result, display)
			}
		}

		return LoadContentFieldsMsg{Fields: result}
	}
}

// LoadContentFieldsForLocaleCmd fetches content fields for a specific content
// node and locale. When locale is non-empty, it uses the locale-filtered query;
// otherwise it falls back to the default (all-locale) query.
func LoadContentFieldsForLocaleCmd(cfg *config.Config, contentDataID types.ContentID, datatypeID types.NullableDatatypeID, locale string) tea.Cmd {
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

		contentID := types.NullableContentID{ID: contentDataID, Valid: true}

		var contentFields *[]db.ContentFields
		var err error
		if locale != "" {
			contentFields, err = d.ListContentFieldsByContentDataAndLocale(contentID, locale)
		} else {
			contentFields, err = d.ListContentFieldsByContentData(contentID)
		}
		if err != nil {
			return LoadContentFieldsMsg{Fields: nil}
		}

		// Fetch field definitions by parent datatype ID
		var fieldDefs *[]db.Fields
		if datatypeID.Valid {
			fieldDefs, err = d.ListFieldsByDatatypeID(datatypeID)
			if err != nil {
				fieldDefs = nil
			}
		}

		// Build content field value map: field_id -> ContentFields
		cfMap := make(map[string]db.ContentFields)
		if contentFields != nil {
			for _, cf := range *contentFields {
				if cf.FieldID.Valid {
					cfMap[string(cf.FieldID.ID)] = cf
				}
			}
		}

		// Build result ordered by sort_order from field definitions
		var result []ContentFieldDisplay
		if fieldDefs != nil {
			result = make([]ContentFieldDisplay, 0, len(*fieldDefs))
			for _, field := range *fieldDefs {
				display := ContentFieldDisplay{
					FieldID:        field.FieldID,
					Label:          field.Label,
					Type:           string(field.Type),
					ValidationJSON: field.Validation,
					DataJSON:       field.Data,
				}
				if cf, ok := cfMap[string(field.FieldID)]; ok {
					display.ContentFieldID = cf.ContentFieldID
					display.Value = cf.FieldValue
				}
				result = append(result, display)
			}
		}

		return LoadContentFieldsMsg{Fields: result}
	}
}

// =============================================================================
// BATCH FIELD LOADING (for grid layout preview)
// =============================================================================

// BatchContentFieldsLoadedMsg carries batch-loaded fields for all tree nodes.
type BatchContentFieldsLoadedMsg struct {
	Fields map[types.ContentID][]ContentFieldDisplay
}

// BatchAdminContentFieldsLoadedMsg carries batch-loaded admin fields for all tree nodes.
type BatchAdminContentFieldsLoadedMsg struct {
	Fields map[types.AdminContentID][]AdminContentFieldDisplay
}

// BatchLoadContentFieldsCmd loads ALL content fields for a route in one pass,
// resolves field definitions, and returns them grouped by ContentDataID.
func BatchLoadContentFieldsCmd(cfg *config.Config, routeID types.RouteID, datatypeIDs []types.DatatypeID, locale string) tea.Cmd {
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

		nRouteID := types.NullableRouteID{ID: routeID, Valid: true}
		var contentFields *[]db.ContentFields
		var err error
		if locale != "" {
			contentFields, err = d.ListContentFieldsByRouteAndLocale(nRouteID, locale)
		} else {
			contentFields, err = d.ListContentFieldsByRoute(nRouteID)
		}
		if err != nil {
			return FetchErrMsg{Error: fmt.Errorf("batch load content fields: %w", err)}
		}

		var allDefs []db.Fields
		for _, dtID := range datatypeIDs {
			nDtID := types.NullableDatatypeID{ID: dtID, Valid: true}
			defs, defErr := d.ListFieldsByDatatypeID(nDtID)
			if defErr != nil {
				continue
			}
			if defs != nil {
				allDefs = append(allDefs, *defs...)
			}
		}

		var cfs []db.ContentFields
		if contentFields != nil {
			cfs = *contentFields
		}

		return BatchContentFieldsLoadedMsg{
			Fields: MapContentFieldsToDisplay(cfs, allDefs),
		}
	}
}

// BatchLoadAdminContentFieldsCmd loads ALL admin content fields for a route in one pass.
func BatchLoadAdminContentFieldsCmd(cfg *config.Config, adminRouteID types.AdminRouteID, datatypeIDs []types.AdminDatatypeID, locale string) tea.Cmd {
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

		nRouteID := types.NullableAdminRouteID{ID: adminRouteID, Valid: true}
		var contentFields *[]db.AdminContentFields
		var err error
		if locale != "" {
			contentFields, err = d.ListAdminContentFieldsByRouteAndLocale(nRouteID, locale)
		} else {
			contentFields, err = d.ListAdminContentFieldsByRoute(nRouteID)
		}
		if err != nil {
			return FetchErrMsg{Error: fmt.Errorf("batch load admin content fields: %w", err)}
		}

		var allDefs []db.AdminFields
		for _, dtID := range datatypeIDs {
			nDtID := types.NullableAdminDatatypeID{ID: dtID, Valid: true}
			defs, defErr := d.ListAdminFieldsByDatatypeID(nDtID)
			if defErr != nil {
				continue
			}
			if defs != nil {
				allDefs = append(allDefs, *defs...)
			}
		}

		var cfs []db.AdminContentFields
		if contentFields != nil {
			cfs = *contentFields
		}

		return BatchAdminContentFieldsLoadedMsg{
			Fields: MapAdminContentFieldsToDisplay(cfs, allDefs),
		}
	}
}

// BatchLoadContentFieldsByRootIDCmd loads ALL content fields for a root_id in one pass.
func BatchLoadContentFieldsByRootIDCmd(cfg *config.Config, rootID types.ContentID, datatypeIDs []types.DatatypeID, locale string) tea.Cmd {
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

		nRootID := types.NullableContentID{ID: rootID, Valid: true}
		var contentFields *[]db.ContentFields
		var err error
		if locale != "" {
			contentFields, err = d.ListContentFieldsByRootIDAndLocale(nRootID, locale)
		} else {
			contentFields, err = d.ListContentFieldsByRootID(nRootID)
		}
		if err != nil {
			return FetchErrMsg{Error: fmt.Errorf("batch load content fields by root_id: %w", err)}
		}

		var allDefs []db.Fields
		for _, dtID := range datatypeIDs {
			nDtID := types.NullableDatatypeID{ID: dtID, Valid: true}
			defs, defErr := d.ListFieldsByDatatypeID(nDtID)
			if defErr != nil {
				continue
			}
			if defs != nil {
				allDefs = append(allDefs, *defs...)
			}
		}

		var cfs []db.ContentFields
		if contentFields != nil {
			cfs = *contentFields
		}

		return BatchContentFieldsLoadedMsg{
			Fields: MapContentFieldsToDisplay(cfs, allDefs),
		}
	}
}

// BatchLoadAdminContentFieldsByRootIDCmd loads ALL admin content fields for a root_id in one pass.
func BatchLoadAdminContentFieldsByRootIDCmd(cfg *config.Config, rootID types.AdminContentID, datatypeIDs []types.AdminDatatypeID, locale string) tea.Cmd {
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

		nRootID := types.NullableAdminContentID{ID: rootID, Valid: true}
		var contentFields *[]db.AdminContentFields
		var err error
		if locale != "" {
			contentFields, err = d.ListAdminContentFieldsByRootIDAndLocale(nRootID, locale)
		} else {
			contentFields, err = d.ListAdminContentFieldsByRootID(nRootID)
		}
		if err != nil {
			return FetchErrMsg{Error: fmt.Errorf("batch load admin content fields by root_id: %w", err)}
		}

		var allDefs []db.AdminFields
		for _, dtID := range datatypeIDs {
			nDtID := types.NullableAdminDatatypeID{ID: dtID, Valid: true}
			defs, defErr := d.ListAdminFieldsByDatatypeID(nDtID)
			if defErr != nil {
				continue
			}
			if defs != nil {
				allDefs = append(allDefs, *defs...)
			}
		}

		var cfs []db.AdminContentFields
		if contentFields != nil {
			cfs = *contentFields
		}

		return BatchAdminContentFieldsLoadedMsg{
			Fields: MapAdminContentFieldsToDisplay(cfs, allDefs),
		}
	}
}

// ContentFieldUpdatedMsg signals that a single content field was updated.
type ContentFieldUpdatedMsg struct {
	ContentID  types.ContentID
	DatatypeID types.NullableDatatypeID
	RouteID    types.RouteID
	AdminMode  bool
}

// ContentFieldDeletedMsg signals that a content field was deleted.
type ContentFieldDeletedMsg struct {
	ContentID  types.ContentID
	DatatypeID types.NullableDatatypeID
	RouteID    types.RouteID
	AdminMode  bool
}

// ContentFieldAddedMsg signals that a new content field was added.
type ContentFieldAddedMsg struct {
	ContentID  types.ContentID
	DatatypeID types.NullableDatatypeID
	RouteID    types.RouteID
	AdminMode  bool
}

// FieldReorderedMsg signals that field sort_order was swapped.
type FieldReorderedMsg struct {
	DatatypeID types.NullableDatatypeID
	ContentID  types.ContentID
	RouteID    types.RouteID
	Direction  string
}

// HandleEditSingleField updates one content field value.
func (m Model) HandleEditSingleField(contentFieldID types.ContentFieldID, contentID types.ContentID, fieldID types.FieldID, newValue string, routeID types.RouteID, datatypeID types.NullableDatatypeID) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		// Get existing content field
		cf, err := d.GetContentField(contentFieldID)
		if err != nil || cf == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Content field not found: %v", err)}
		}

		_, err = d.UpdateContentField(ctx, ac, db.UpdateContentFieldParams{
			ContentFieldID: contentFieldID,
			RouteID:        cf.RouteID,
			ContentDataID:  types.NullableContentID{ID: contentID, Valid: true},
			FieldID:        types.NullableFieldID{ID: fieldID, Valid: true},
			FieldValue:     newValue,
			AuthorID:       userID,
			DateCreated:    cf.DateCreated,
			DateModified:   types.TimestampNow(),
		})
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to update field: %v", err)}
		}

		return ContentFieldUpdatedMsg{
			ContentID:  contentID,
			DatatypeID: datatypeID,
			RouteID:    routeID,
		}
	}
}

// HandleDeleteContentField deletes a content field record.
func (m Model) HandleDeleteContentField(contentFieldID types.ContentFieldID, contentID types.ContentID, routeID types.RouteID, datatypeID types.NullableDatatypeID) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		err := d.DeleteContentField(ctx, ac, contentFieldID)
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to delete field: %v", err)}
		}

		return ContentFieldDeletedMsg{
			ContentID:  contentID,
			DatatypeID: datatypeID,
			RouteID:    routeID,
		}
	}
}

// HandleAddContentField creates a new content field record for a field not yet populated.
func (m Model) HandleAddContentField(contentID types.ContentID, fieldID types.FieldID, routeID types.RouteID, datatypeID types.NullableDatatypeID) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		_, err := d.CreateContentField(ctx, ac, db.CreateContentFieldParams{
			ContentDataID: types.NullableContentID{ID: contentID, Valid: true},
			FieldID:       types.NullableFieldID{ID: fieldID, Valid: true},
			FieldValue:    "",
			RouteID:       types.NullableRouteID{ID: routeID, Valid: true},
			AuthorID:      userID,
			DateCreated:   types.TimestampNow(),
			DateModified:  types.TimestampNow(),
		})
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to add content field: %v", err)}
		}

		return ContentFieldAddedMsg{
			ContentID:  contentID,
			DatatypeID: datatypeID,
			RouteID:    routeID,
		}
	}
}

// HandleReorderField swaps sort_order between two fields.
func (m Model) HandleReorderField(aID string, bID string, aOrder int64, bOrder int64, datatypeID types.NullableDatatypeID, contentID types.ContentID, routeID types.RouteID, direction string) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		if err := d.UpdateFieldSortOrder(ctx, ac, db.UpdateFieldSortOrderParams{
			FieldID:   types.FieldID(aID),
			SortOrder: bOrder,
		}); err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to reorder: %v", err)}
		}
		if err := d.UpdateFieldSortOrder(ctx, ac, db.UpdateFieldSortOrderParams{
			FieldID:   types.FieldID(bID),
			SortOrder: aOrder,
		}); err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to reorder: %v", err)}
		}

		return FieldReorderedMsg{
			DatatypeID: datatypeID,
			ContentID:  contentID,
			RouteID:    routeID,
			Direction:  direction,
		}
	}
}

// ReorderDatatypeRequestMsg requests a sort_order swap between two datatypes.
type ReorderDatatypeRequestMsg struct {
	AID       types.DatatypeID
	BID       types.DatatypeID
	AOrder    int64
	BOrder    int64
	Direction string
}

// ReorderDatatypeCmd creates a command to request a datatype reorder.
func ReorderDatatypeCmd(aID, bID types.DatatypeID, aOrder, bOrder int64, direction string) tea.Cmd {
	return func() tea.Msg {
		return ReorderDatatypeRequestMsg{AID: aID, BID: bID, AOrder: aOrder, BOrder: bOrder, Direction: direction}
	}
}

// DatatypeReorderedMsg signals that datatype sort_order was swapped.
type DatatypeReorderedMsg struct {
	Direction string
}

// HandleReorderDatatype swaps sort_order between two datatypes.
func (m Model) HandleReorderDatatype(msg ReorderDatatypeRequestMsg) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		if err := d.UpdateDatatypeSortOrder(ctx, ac, db.UpdateDatatypeSortOrderParams{
			DatatypeID: msg.AID,
			SortOrder:  msg.BOrder,
		}); err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to reorder: %v", err)}
		}
		if err := d.UpdateDatatypeSortOrder(ctx, ac, db.UpdateDatatypeSortOrderParams{
			DatatypeID: msg.BID,
			SortOrder:  msg.AOrder,
		}); err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to reorder: %v", err)}
		}

		return DatatypeReorderedMsg{Direction: msg.Direction}
	}
}

// ReorderAdminDatatypeRequestMsg requests a sort_order swap between two admin datatypes.
type ReorderAdminDatatypeRequestMsg struct {
	AID       types.AdminDatatypeID
	BID       types.AdminDatatypeID
	AOrder    int64
	BOrder    int64
	Direction string
}

// ReorderAdminDatatypeCmd creates a command to request an admin datatype reorder.
func ReorderAdminDatatypeCmd(aID, bID types.AdminDatatypeID, aOrder, bOrder int64, direction string) tea.Cmd {
	return func() tea.Msg {
		return ReorderAdminDatatypeRequestMsg{AID: aID, BID: bID, AOrder: aOrder, BOrder: bOrder, Direction: direction}
	}
}

// AdminDatatypeReorderedMsg signals that admin datatype sort_order was swapped.
type AdminDatatypeReorderedMsg struct {
	Direction string
}

// HandleReorderAdminDatatype swaps sort_order between two admin datatypes.
func (m Model) HandleReorderAdminDatatype(msg ReorderAdminDatatypeRequestMsg) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		if err := d.UpdateAdminDatatypeSortOrder(ctx, ac, db.UpdateAdminDatatypeSortOrderParams{
			AdminDatatypeID: msg.AID,
			SortOrder:       msg.BOrder,
		}); err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to reorder: %v", err)}
		}
		if err := d.UpdateAdminDatatypeSortOrder(ctx, ac, db.UpdateAdminDatatypeSortOrderParams{
			AdminDatatypeID: msg.BID,
			SortOrder:       msg.AOrder,
		}); err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to reorder: %v", err)}
		}

		return AdminDatatypeReorderedMsg{Direction: msg.Direction}
	}
}

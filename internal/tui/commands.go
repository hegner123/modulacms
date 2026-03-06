package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/media"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/plugin"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/tree"
	"github.com/hegner123/modulacms/internal/utility"
)

// DatabaseCMD specifies the type of database operation.
type DatabaseCMD string

// Database command types.
const (
	INSERT DatabaseCMD = "insert"
	SELECT DatabaseCMD = "select"
	UPDATE DatabaseCMD = "update"
	DELETE DatabaseCMD = "delete"
	BATCH  DatabaseCMD = "batch"
)

// ForeignKeyReference describes a foreign key constraint.
type ForeignKeyReference struct {
	From   string
	Table  string // Referenced table name.
	Column string // Referenced column name.
}

// FetchTables carries table names from a fetch operation.
type FetchTables struct {
	Tables []string
}

// GetTablesCMD creates a command to fetch all table names from the database.
func GetTablesCMD(c *config.Config) tea.Cmd {
	return func() tea.Msg {
		var (
			d      db.DbDriver
			labels []string
		)
		d = db.ConfigDB(*c)
		tables, err := d.ListTables()
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return FetchErrMsg{Error: err}
		}

		for _, table := range *tables {
			labels = append(labels, table.Label)
		}
		return TablesSet{Tables: labels}
	}
}

// TODO Add default case for generic operations
// DatabaseInsert creates a command to insert a row into the specified table.
func (m Model) DatabaseInsert(c *config.Config, table db.DBTable, columns []string, values []*string) tea.Cmd {
	d := db.ConfigDB(*c)
	con, _, err := d.GetConnection()
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	valuesMap := make(map[string]any, 0)
	for i, v := range values {
		if i == 0 || v == nil {
			continue
		} else {
			valuesMap[columns[i]] = *v
		}
	}
	dialect := db.DialectFromString(string(c.Db_Driver))
	res, err := db.QInsert(context.Background(), con, dialect, db.InsertParams{
		Table:  string(table),
		Values: valuesMap,
	})
	if err != nil {
		return tea.Batch(
			LogMessageCmd(err.Error()),
			LogMessageCmd(fmt.Sprintln(valuesMap)),
		)
	}

	return tea.Batch(
		DbResultCmd(res, string(table)),
	)
}

// DatabaseUpdate creates a command to update a row in the specified table.
func (m Model) DatabaseUpdate(c *config.Config, table db.DBTable, rowID int64, valuesMap map[string]any) tea.Cmd {
	d := db.ConfigDB(*c)

	con, _, err := d.GetConnection()
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	dialect := db.DialectFromString(string(c.Db_Driver))
	res, err := db.QUpdate(context.Background(), con, dialect, db.UpdateParams{
		Table: string(table),
		Set:   valuesMap,
		Where: map[string]any{"id": rowID},
	})
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	m.Logger.Finfo("CLI Update successful", nil)
	return DbResultCmd(res, string(table))
}

// DatabaseGet creates a command to fetch a single row by ID from the specified table.
func (m Model) DatabaseGet(c *config.Config, source FetchSource, table db.DBTable, id int64) tea.Cmd {
	d := db.ConfigDB(*c)

	con, _, err := d.GetConnection()
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	dialect := db.DialectFromString(string(c.Db_Driver))
	r, err := db.QSelectRows(context.Background(), con, dialect, db.SelectParams{
		Table: string(table),
		Where: map[string]any{"id": id},
	})
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	defer utility.HandleRowsCloseDeferErr(r)
	out, err := Parse(r, table)
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	return DatabaseGetRowsCmd(source, out, table)

}

// DatabaseList creates a command to fetch all rows from the specified table.
func (m Model) DatabaseList(c *config.Config, source FetchSource, table db.DBTable) tea.Cmd {
	d := db.ConfigDB(*c)

	con, _, err := d.GetConnection()
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	dialect := db.DialectFromString(string(c.Db_Driver))
	r, err := db.QSelectRows(context.Background(), con, dialect, db.SelectParams{
		Table: string(table),
	})
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	defer utility.HandleRowsCloseDeferErr(r)
	out, err := Parse(r, table)
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	return DatabaseListRowsCmd(source, out, table)

}

// DatabaseFilteredList creates a command to fetch filtered rows from the specified table.
func (m Model) DatabaseFilteredList(c *config.Config, source FetchSource, table db.DBTable, columns []string, whereColumn string, value any) tea.Cmd {
	d := db.ConfigDB(*c)

	con, _, err := d.GetConnection()
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	dialect := db.DialectFromString(string(c.Db_Driver))
	r, err := db.QSelectRows(context.Background(), con, dialect, db.SelectParams{
		Table:   string(table),
		Columns: columns,
		Where:   map[string]any{whereColumn: value},
	})
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	defer utility.HandleRowsCloseDeferErr(r)
	out, err := Parse(r, table)
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	return DatabaseListRowsCmd(source, out, table)
}

// DatabaseDelete creates a command to delete the current row from the specified table.
func (m Model) DatabaseDelete(c *config.Config, table db.DBTable) tea.Cmd {
	id := m.GetCurrentRowId()
	d := db.ConfigDB(*c)

	con, _, err := d.GetConnection()
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	dialect := db.DialectFromString(string(c.Db_Driver))
	res, err := db.QDelete(context.Background(), con, dialect, db.DeleteParams{
		Table: string(table),
		Where: map[string]any{"id": id},
	})
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	return DbResultCmd(res, string(table))
}

// GetContentField marshals the current row to JSON bytes.
func (m Model) GetContentField(node *string) []byte {
	row := m.TableState.Rows[m.Cursor]
	j, err := json.Marshal(row)
	if err != nil {
		m.Logger.Ferror("", err)
	}
	return j
}

// GetFullTree creates a command to fetch the full content tree for a route.
func (m Model) GetFullTree(c *config.Config, id types.RouteID) tea.Cmd {
	// TODO: Implement tree retrieval logic
	d := db.ConfigDB(*c)
	routeID := types.NullableRouteID{ID: id, Valid: true}
	res, err := d.GetRouteTreeByRouteID(routeID)
	if err != nil {
		return ErrorSetCmd(err)
	}
	return GetFullTreeResCMD(*res)
}

// GetContentInstances creates a command to fetch content instances (stub for future implementation).
func (m Model) GetContentInstances(c *config.Config) tea.Cmd {
	//    d := db.ConfigDB(*c)
	//TODO JOIN STATEMENTS FOR CONTENT DATA AND DATATYPES
	//TODO JOIN STATEMENTS FOR CONTENT FIELDS AND FIELDS

	return tea.Batch()
}

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
			ContentDataID: contentData.ContentDataID,
			RouteID:       routeID,
			FieldCount:    createdFields,
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

		// Step 1: Create ContentData using typed DbDriver method
		contentData, err := d.CreateContentData(ctx, ac, db.CreateContentDataParams{
			DatatypeID:    types.NullableDatatypeID{ID: msg.DatatypeID, Valid: true},
			RouteID:       types.NullableRouteID{ID: msg.RouteID, Valid: true},
			AuthorID:      authorID,
			Status:        types.ContentStatusDraft,
			DateCreated:   types.TimestampNow(),
			DateModified:  types.TimestampNow(),
			ParentID:      msg.ParentID,
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
					RouteID:       types.NullableRouteID{ID: msg.RouteID, Valid: true},
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
					RouteID:       types.NullableRouteID{ID: msg.RouteID, Valid: true},
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
					RouteID:       types.NullableRouteID{ID: msg.RouteID, Valid: true},
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
		content, err := d.GetContentData(contentID)
		if err != nil {
			logger.Ferror("Failed to get content for deletion", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Content not found: %v", err),
			}
		}

		// Check if it has children (should have been prevented by UI, but double-check)
		if content.FirstChildID.Valid && content.FirstChildID.ID != "" {
			return ActionResultMsg{
				Title:   "Cannot Delete",
				Message: "This content has children. Delete child nodes first.",
			}
		}

		// Update sibling pointers before deletion
		// If this node has a previous sibling, point its next to our next
		if content.PrevSiblingID.Valid && content.PrevSiblingID.ID != "" {
			prevSiblingID := content.PrevSiblingID.ID
			prevSibling, err := d.GetContentData(prevSiblingID)
			if err == nil && prevSibling != nil {
				updateParams := db.UpdateContentDataParams{
					ContentDataID: prevSibling.ContentDataID,
					RouteID:       prevSibling.RouteID,
					ParentID:      prevSibling.ParentID,
					FirstChildID:  prevSibling.FirstChildID,
					NextSiblingID: content.NextSiblingID, // Point to our next sibling
					PrevSiblingID: prevSibling.PrevSiblingID,
					DatatypeID:    prevSibling.DatatypeID,
					AuthorID:      prevSibling.AuthorID,
					Status:        prevSibling.Status,
					DateCreated:   prevSibling.DateCreated,
					DateModified:  types.TimestampNow(),
				}
				if _, err := d.UpdateContentData(ctx, ac, updateParams); err != nil {
					logger.Ferror("Failed to update prev sibling", err)
				}
			}
		}

		// If this node has a next sibling, point its prev to our prev
		if content.NextSiblingID.Valid && content.NextSiblingID.ID != "" {
			nextSiblingID := content.NextSiblingID.ID
			nextSibling, err := d.GetContentData(nextSiblingID)
			if err == nil && nextSibling != nil {
				updateParams := db.UpdateContentDataParams{
					ContentDataID: nextSibling.ContentDataID,
					RouteID:       nextSibling.RouteID,
					ParentID:      nextSibling.ParentID,
					FirstChildID:  nextSibling.FirstChildID,
					NextSiblingID: nextSibling.NextSiblingID,
					PrevSiblingID: content.PrevSiblingID, // Point to our prev sibling
					DatatypeID:    nextSibling.DatatypeID,
					AuthorID:      nextSibling.AuthorID,
					Status:        nextSibling.Status,
					DateCreated:   nextSibling.DateCreated,
					DateModified:  types.TimestampNow(),
				}
				if _, err := d.UpdateContentData(ctx, ac, updateParams); err != nil {
					logger.Ferror("Failed to update next sibling", err)
				}
			}
		}

		// If this is the first child of parent, update parent's first_child to our next sibling
		if content.ParentID.Valid {
			parent, err := d.GetContentData(content.ParentID.ID)
			if err == nil && parent != nil {
				if parent.FirstChildID.Valid && parent.FirstChildID.ID == contentID {
					updateParams := db.UpdateContentDataParams{
						ContentDataID: parent.ContentDataID,
						RouteID:       parent.RouteID,
						ParentID:      parent.ParentID,
						FirstChildID:  content.NextSiblingID, // Point to our next sibling (or null)
						NextSiblingID: parent.NextSiblingID,
						PrevSiblingID: parent.PrevSiblingID,
						DatatypeID:    parent.DatatypeID,
						AuthorID:      parent.AuthorID,
						Status:        parent.Status,
						DateCreated:   parent.DateCreated,
						DateModified:  types.TimestampNow(),
					}
					if _, err := d.UpdateContentData(ctx, ac, updateParams); err != nil {
						logger.Ferror("Failed to update parent first_child", err)
					}
				}
			}
		}

		// Delete the content data (content_fields will cascade delete)
		if err := d.DeleteContentData(ctx, ac, contentID); err != nil {
			logger.Ferror("Failed to delete content", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to delete content: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Content deleted successfully: %s", contentID))
		return ContentDeletedMsg{
			ContentID: msg.ContentID,
			RouteID:   msg.RouteID,
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

		// Read source node
		source, err := d.GetContentData(msg.SourceContentID)
		if err != nil || source == nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Source content not found: %v", err),
			}
		}

		// Read target (new parent) node
		target, err := d.GetContentData(msg.TargetContentID)
		if err != nil || target == nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Target content not found: %v", err),
			}
		}

		// === STEP 1: Detach source from old position ===

		// If source has a previous sibling, update its NextSiblingID to source's NextSiblingID
		if source.PrevSiblingID.Valid && source.PrevSiblingID.ID != "" {
			prevID := source.PrevSiblingID.ID
			prev, prevErr := d.GetContentData(prevID)
			if prevErr == nil && prev != nil {
				params := db.UpdateContentDataParams{
					ContentDataID: prev.ContentDataID,
					RouteID:       prev.RouteID,
					ParentID:      prev.ParentID,
					FirstChildID:  prev.FirstChildID,
					NextSiblingID: source.NextSiblingID, // point to source's next
					PrevSiblingID: prev.PrevSiblingID,
					DatatypeID:    prev.DatatypeID,
					AuthorID:      prev.AuthorID,
					Status:        prev.Status,
					DateCreated:   prev.DateCreated,
					DateModified:  types.TimestampNow(),
				}
				if _, updateErr := d.UpdateContentData(ctx, ac, params); updateErr != nil {
					logger.Ferror("Failed to update prev sibling during move", updateErr)
				}
			}
		}

		// If source has a next sibling, update its PrevSiblingID to source's PrevSiblingID
		if source.NextSiblingID.Valid && source.NextSiblingID.ID != "" {
			nextID := source.NextSiblingID.ID
			next, nextErr := d.GetContentData(nextID)
			if nextErr == nil && next != nil {
				params := db.UpdateContentDataParams{
					ContentDataID: next.ContentDataID,
					RouteID:       next.RouteID,
					ParentID:      next.ParentID,
					FirstChildID:  next.FirstChildID,
					NextSiblingID: next.NextSiblingID,
					PrevSiblingID: source.PrevSiblingID, // point to source's prev
					DatatypeID:    next.DatatypeID,
					AuthorID:      next.AuthorID,
					Status:        next.Status,
					DateCreated:   next.DateCreated,
					DateModified:  types.TimestampNow(),
				}
				if _, updateErr := d.UpdateContentData(ctx, ac, params); updateErr != nil {
					logger.Ferror("Failed to update next sibling during move", updateErr)
				}
			}
		}

		// If source is first child of its old parent, update parent's FirstChildID
		if source.ParentID.Valid {
			oldParent, parentErr := d.GetContentData(source.ParentID.ID)
			if parentErr == nil && oldParent != nil {
				if oldParent.FirstChildID.Valid && oldParent.FirstChildID.ID == source.ContentDataID {
					params := db.UpdateContentDataParams{
						ContentDataID: oldParent.ContentDataID,
						RouteID:       oldParent.RouteID,
						ParentID:      oldParent.ParentID,
						FirstChildID:  source.NextSiblingID, // point to source's next (or null)
						NextSiblingID: oldParent.NextSiblingID,
						PrevSiblingID: oldParent.PrevSiblingID,
						DatatypeID:    oldParent.DatatypeID,
						AuthorID:      oldParent.AuthorID,
						Status:        oldParent.Status,
						DateCreated:   oldParent.DateCreated,
						DateModified:  types.TimestampNow(),
					}
					if _, updateErr := d.UpdateContentData(ctx, ac, params); updateErr != nil {
						logger.Ferror("Failed to update old parent first_child during move", updateErr)
					}
				}
			}
		}

		// === STEP 2: Attach source as last child of target ===

		// Re-read target to get fresh state after potential updates above
		target, err = d.GetContentData(msg.TargetContentID)
		if err != nil || target == nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Target content not found after detach: %v", err),
			}
		}

		newPrevSiblingID := types.NullableContentID{} // default: no prev sibling

		if !target.FirstChildID.Valid || target.FirstChildID.ID == "" {
			// Target has no children - source becomes first child
			targetParams := db.UpdateContentDataParams{
				ContentDataID: target.ContentDataID,
				RouteID:       target.RouteID,
				ParentID:      target.ParentID,
				FirstChildID:  types.NullableContentID{ID: source.ContentDataID, Valid: true},
				NextSiblingID: target.NextSiblingID,
				PrevSiblingID: target.PrevSiblingID,
				DatatypeID:    target.DatatypeID,
				AuthorID:      target.AuthorID,
				Status:        target.Status,
				DateCreated:   target.DateCreated,
				DateModified:  types.TimestampNow(),
			}
			if _, updateErr := d.UpdateContentData(ctx, ac, targetParams); updateErr != nil {
				logger.Ferror("Failed to set target first_child", updateErr)
				return ActionResultMsg{
					Title:   "Error",
					Message: fmt.Sprintf("Failed to update target: %v", updateErr),
				}
			}
		} else {
			// Target has children - walk to last sibling and append source
			currentID := target.FirstChildID.ID
			for {
				current, walkErr := d.GetContentData(currentID)
				if walkErr != nil || current == nil {
					break
				}
				if !current.NextSiblingID.Valid || current.NextSiblingID.ID == "" {
					// Found last sibling - update its NextSiblingID to source
					lastParams := db.UpdateContentDataParams{
						ContentDataID: current.ContentDataID,
						RouteID:       current.RouteID,
						ParentID:      current.ParentID,
						FirstChildID:  current.FirstChildID,
						NextSiblingID: types.NullableContentID{ID: source.ContentDataID, Valid: true},
						PrevSiblingID: current.PrevSiblingID,
						DatatypeID:    current.DatatypeID,
						AuthorID:      current.AuthorID,
						Status:        current.Status,
						DateCreated:   current.DateCreated,
						DateModified:  types.TimestampNow(),
					}
					if _, updateErr := d.UpdateContentData(ctx, ac, lastParams); updateErr != nil {
						logger.Ferror("Failed to update last sibling next pointer", updateErr)
					}
					newPrevSiblingID = types.NullableContentID{ID: current.ContentDataID, Valid: true}
					break
				}
				currentID = current.NextSiblingID.ID
			}
		}

		// === STEP 3: Update source node with new parent and cleared sibling pointers ===
		sourceParams := db.UpdateContentDataParams{
			ContentDataID: source.ContentDataID,
			RouteID:       source.RouteID,
			ParentID:      types.NullableContentID{ID: msg.TargetContentID, Valid: true},
			FirstChildID:  source.FirstChildID,       // preserve children
			NextSiblingID: types.NullableContentID{}, // last child, no next
			PrevSiblingID: newPrevSiblingID,
			DatatypeID:    source.DatatypeID,
			AuthorID:      source.AuthorID,
			Status:        source.Status,
			DateCreated:   source.DateCreated,
			DateModified:  types.TimestampNow(),
		}
		if _, updateErr := d.UpdateContentData(ctx, ac, sourceParams); updateErr != nil {
			logger.Ferror("Failed to update source node", updateErr)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update source: %v", updateErr),
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

// HandleCreateUserFromDialog processes the user creation request
func (m Model) HandleCreateUserFromDialog(msg CreateUserFromDialogRequestMsg) tea.Cmd {
	cfg := m.Config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create user: configuration not loaded",
			}
		}
	}

	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		email := types.Email(msg.Email)
		if err := email.Validate(); err != nil {
			return ActionResultMsg{
				Title:   "Invalid Email",
				Message: fmt.Sprintf("Could not create user: %v", err),
			}
		}

		if msg.Password == "" {
			return ActionResultMsg{
				Title:   "Password Required",
				Message: "Password must not be empty",
			}
		}

		hash, err := auth.HashPassword(msg.Password)
		if err != nil {
			return ActionResultMsg{
				Title:   "Password Error",
				Message: fmt.Sprintf("Could not hash password: %v", err),
			}
		}

		user, err := d.CreateUser(ctx, ac, db.CreateUserParams{
			Username:     msg.Username,
			Name:         msg.Name,
			Email:        email,
			Hash:         hash,
			Role:         msg.Role,
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		})
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to create user: %v", err),
			}
		}
		if user == nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Failed to create user in database",
			}
		}

		return UserCreatedFromDialogMsg{
			UserID:   user.UserID,
			Username: user.Username,
		}
	}
}

// HandleUpdateUserFromDialog processes the user update request
func (m Model) HandleUpdateUserFromDialog(msg UpdateUserFromDialogRequestMsg) tea.Cmd {
	cfg := m.Config
	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot update user: configuration not loaded",
			}
		}
	}

	callerUserID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, callerUserID)

		userID := types.UserID(msg.UserID)

		// Get existing user to preserve hash
		existingUser, err := d.GetUser(userID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("User not found: %v", err),
			}
		}

		email := types.Email(msg.Email)
		if err := email.Validate(); err != nil {
			return ActionResultMsg{
				Title:   "Invalid Email",
				Message: fmt.Sprintf("Could not update user: %v", err),
			}
		}

		_, err = d.UpdateUser(ctx, ac, db.UpdateUserParams{
			UserID:       userID,
			Username:     msg.Username,
			Name:         msg.Name,
			Email:        email,
			Hash:         existingUser.Hash, // Preserve existing hash
			Role:         msg.Role,
			DateCreated:  existingUser.DateCreated,
			DateModified: types.TimestampNow(),
		})
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update user: %v", err),
			}
		}

		return UserUpdatedFromDialogMsg{
			UserID:   userID,
			Username: msg.Username,
		}
	}
}

// HandleDeleteUser deletes a user
func (m Model) HandleDeleteUser(msg DeleteUserRequestMsg) tea.Cmd {
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

		logger.Finfo(fmt.Sprintf("Deleting user: %s", msg.UserID))

		if err := d.DeleteUser(ctx, ac, msg.UserID); err != nil {
			logger.Ferror("Failed to delete user", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to delete user: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("User deleted successfully: %s", msg.UserID))
		return UserDeletedMsg{UserID: msg.UserID}
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

// ContentFieldUpdatedMsg signals that a single content field was updated.
type ContentFieldUpdatedMsg struct {
	ContentID  types.ContentID
	DatatypeID types.NullableDatatypeID
	RouteID    types.RouteID
}

// ContentFieldDeletedMsg signals that a content field was deleted.
type ContentFieldDeletedMsg struct {
	ContentID  types.ContentID
	DatatypeID types.NullableDatatypeID
	RouteID    types.RouteID
}

// ContentFieldAddedMsg signals that a new content field was added.
type ContentFieldAddedMsg struct {
	ContentID  types.ContentID
	DatatypeID types.NullableDatatypeID
	RouteID    types.RouteID
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

// HandleDeleteDatatype deletes a datatype and its junction records
func (m Model) HandleDeleteDatatype(msg DeleteDatatypeRequestMsg) tea.Cmd {
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

		logger.Finfo(fmt.Sprintf("Deleting datatype: %s", msg.DatatypeID))

		// Check for child datatypes that reference this one as parent
		allDatatypes, err := d.ListDatatypes()
		if err != nil {
			logger.Ferror("Failed to list datatypes for child check", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to check for child datatypes: %v", err),
			}
		}
		if allDatatypes != nil {
			for _, dt := range *allDatatypes {
				if dt.ParentID.Valid && types.DatatypeID(dt.ParentID.ID) == msg.DatatypeID {
					return ActionResultMsg{
						Title:   "Cannot Delete",
						Message: fmt.Sprintf("Datatype has child '%s'. Delete children first.", dt.Label),
					}
				}
			}
		}

		// Delete the datatype (fields with parent_id referencing it are set to NULL by FK cascade)
		if err := d.DeleteDatatype(ctx, ac, msg.DatatypeID); err != nil {
			logger.Ferror("Failed to delete datatype", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to delete datatype: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Datatype deleted successfully: %s", msg.DatatypeID))
		return DatatypeDeletedMsg{DatatypeID: msg.DatatypeID}
	}
}

// HandleDeleteField deletes a field
func (m Model) HandleDeleteField(msg DeleteFieldRequestMsg) tea.Cmd {
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

		logger.Finfo(fmt.Sprintf("Deleting field: %s from datatype: %s", msg.FieldID, msg.DatatypeID))

		// Delete the field
		if err := d.DeleteField(ctx, ac, msg.FieldID); err != nil {
			logger.Ferror("Failed to delete field", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to delete field: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Field deleted successfully: %s", msg.FieldID))
		return FieldDeletedMsg{
			FieldID:    msg.FieldID,
			DatatypeID: msg.DatatypeID,
		}
	}
}

// HandleDeleteRoute deletes a route
func (m Model) HandleDeleteRoute(msg DeleteRouteRequestMsg) tea.Cmd {
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

		logger.Finfo(fmt.Sprintf("Deleting route: %s", msg.RouteID))

		if err := d.DeleteRoute(ctx, ac, msg.RouteID); err != nil {
			logger.Ferror("Failed to delete route", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to delete route: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Route deleted successfully: %s", msg.RouteID))
		return RouteDeletedMsg{RouteID: msg.RouteID}
	}
}

// HandleDeleteMedia deletes a media item
func (m Model) HandleDeleteMedia(msg DeleteMediaRequestMsg) tea.Cmd {
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

		logger.Finfo(fmt.Sprintf("Deleting media: %s", msg.MediaID))

		if err := d.DeleteMedia(ctx, ac, msg.MediaID); err != nil {
			logger.Ferror("Failed to delete media", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to delete media: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Media deleted successfully: %s", msg.MediaID))
		return MediaDeletedMsg{MediaID: msg.MediaID}
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
			logger.Finfo(fmt.Sprintf("No rows returned for route %s", routeID))
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

// MediaUploader is a consumer-defined interface satisfied by RemoteDriver.
// The TUI type-asserts the DbDriver to this interface for remote media uploads.
type MediaUploader interface {
	UploadMedia(ctx context.Context, filePath string) (*db.Media, error)
}

// MediaProgressUploader extends MediaUploader with progress callback support.
type MediaProgressUploader interface {
	MediaUploader
	UploadMediaWithProgress(ctx context.Context, filePath string, progressFn func(bytesSent int64, total int64)) (*db.Media, error)
}

// waitForMsg returns a tea.Cmd that blocks until a message arrives on ch.
func waitForMsg(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

// HandleMediaUpload runs the media upload pipeline asynchronously.
// In remote mode, the file is sent to the server via the SDK with progress.
// In local mode, the existing optimize+S3 pipeline runs.
func (m Model) HandleMediaUpload(msg MediaUploadStartMsg) tea.Cmd {
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}
	cfg := m.Config
	isRemote := m.IsRemote
	return func() tea.Msg {
		filename := filepath.Base(msg.FilePath)
		baseName := strings.TrimSuffix(filename, filepath.Ext(filename))

		logger.Finfo(fmt.Sprintf("Starting media upload: %s", filename))

		// Remote mode: upload via SDK with progress channel
		if isRemote {
			d := db.ConfigDB(*cfg)
			progressCh := make(chan tea.Msg, 1)

			// Try progress-aware uploader first, fall back to basic
			if pu, ok := d.(MediaProgressUploader); ok {
				go func() {
					progressFn := func(bytesSent int64, total int64) {
						select {
						case progressCh <- MediaUploadProgressMsg{
							BytesSent:  bytesSent,
							Total:      total,
							ProgressCh: progressCh,
						}:
						default: // don't block if channel is full
						}
					}
					_, err := pu.UploadMediaWithProgress(context.Background(), msg.FilePath, progressFn)
					if err != nil {
						logger.Ferror("Remote media upload failed", err)
						progressCh <- ActionResultMsg{
							Title:   "Upload Error",
							Message: fmt.Sprintf("Upload failed: %v", err),
							IsError: true,
						}
						return
					}
					logger.Finfo(fmt.Sprintf("Media uploaded remotely: %s", baseName))
					progressCh <- MediaUploadedMsg{Name: baseName}
				}()
				return <-progressCh
			}

			// Fall back to basic uploader without progress
			uploader, ok := d.(MediaUploader)
			if !ok {
				return ActionResultMsg{
					Title:   "Upload Error",
					Message: "Remote driver does not support media upload",
					IsError: true,
				}
			}
			_, err := uploader.UploadMedia(context.Background(), msg.FilePath)
			if err != nil {
				logger.Ferror("Remote media upload failed", err)
				return ActionResultMsg{
					Title:   "Upload Error",
					Message: fmt.Sprintf("Upload failed: %v", err),
					IsError: true,
				}
			}
			logger.Finfo(fmt.Sprintf("Media uploaded remotely: %s", baseName))
			return MediaUploadedMsg{Name: baseName}
		}

		// Local mode: existing pipeline
		// Step 1: Create placeholder DB record
		_, err := media.CreateMedia(baseName, *cfg)
		if err != nil {
			logger.Ferror("Failed to create media record", err)
			return ActionResultMsg{
				Title:   "Upload Error",
				Message: fmt.Sprintf("Failed to create media record: %v", err),
			}
		}

		// Step 2: Create temp directory for optimized files
		tmpDir, err := os.MkdirTemp("", media.TempDirPrefix)
		if err != nil {
			logger.Ferror("Failed to create temp directory", err)
			return ActionResultMsg{
				Title:   "Upload Error",
				Message: fmt.Sprintf("Failed to create temp directory: %v", err),
			}
		}
		defer os.RemoveAll(tmpDir)

		// Step 3: Run upload pipeline (optimize -> S3 upload -> DB update)
		if err := media.HandleMediaUpload(msg.FilePath, tmpDir, *cfg); err != nil {
			logger.Ferror("Media upload failed", err)
			return ActionResultMsg{
				Title:   "Upload Error",
				Message: fmt.Sprintf("Upload failed: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Media uploaded successfully: %s", baseName))
		return MediaUploadedMsg{Name: baseName}
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
			return ActionResultMsg{Title: "Error", Message: "Configuration not loaded"}
		}
	}

	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)
		logger := utility.DefaultLogger

		// Read the node to move
		a, err := d.GetContentData(msg.ContentID)
		if err != nil || a == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Content not found: %v", err)}
		}

		if msg.Direction == "up" {
			// Move up: swap A with its prev sibling B
			if !a.PrevSiblingID.Valid || a.PrevSiblingID.ID == "" {
				return ActionResultMsg{Title: "Info", Message: "Already at top"}
			}
			bID := a.PrevSiblingID.ID
			b, bErr := d.GetContentData(bID)
			if bErr != nil || b == nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Previous sibling not found: %v", bErr)}
			}

			// Before: [C?] <-> B <-> A <-> [D?]
			// After:  [C?] <-> A <-> B <-> [D?]

			// If B has a prev (C), update C.NextSiblingID -> A
			if b.PrevSiblingID.Valid && b.PrevSiblingID.ID != "" {
				cID := b.PrevSiblingID.ID
				c, cErr := d.GetContentData(cID)
				if cErr == nil && c != nil {
					_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
						ContentDataID: c.ContentDataID,
						RouteID:       c.RouteID,
						ParentID:      c.ParentID,
						FirstChildID:  c.FirstChildID,
						NextSiblingID: types.NullableContentID{ID: a.ContentDataID, Valid: true},
						PrevSiblingID: c.PrevSiblingID,
						DatatypeID:    c.DatatypeID,
						AuthorID:      c.AuthorID,
						Status:        c.Status,
						DateCreated:   c.DateCreated,
						DateModified:  types.TimestampNow(),
					})
					if updateErr != nil {
						logger.Ferror("Failed to update C during reorder up", updateErr)
					}
				}
			}

			// If A has a next (D), update D.PrevSiblingID -> B
			if a.NextSiblingID.Valid && a.NextSiblingID.ID != "" {
				dID := a.NextSiblingID.ID
				dNode, dErr := d.GetContentData(dID)
				if dErr == nil && dNode != nil {
					_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
						ContentDataID: dNode.ContentDataID,
						RouteID:       dNode.RouteID,
						ParentID:      dNode.ParentID,
						FirstChildID:  dNode.FirstChildID,
						NextSiblingID: dNode.NextSiblingID,
						PrevSiblingID: types.NullableContentID{ID: b.ContentDataID, Valid: true},
						DatatypeID:    dNode.DatatypeID,
						AuthorID:      dNode.AuthorID,
						Status:        dNode.Status,
						DateCreated:   dNode.DateCreated,
						DateModified:  types.TimestampNow(),
					})
					if updateErr != nil {
						logger.Ferror("Failed to update D during reorder up", updateErr)
					}
				}
			}

			// If parent.FirstChildID == B, update to A
			if a.ParentID.Valid {
				parent, pErr := d.GetContentData(a.ParentID.ID)
				if pErr == nil && parent != nil && parent.FirstChildID.Valid && parent.FirstChildID.ID == b.ContentDataID {
					_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
						ContentDataID: parent.ContentDataID,
						RouteID:       parent.RouteID,
						ParentID:      parent.ParentID,
						FirstChildID:  types.NullableContentID{ID: a.ContentDataID, Valid: true},
						NextSiblingID: parent.NextSiblingID,
						PrevSiblingID: parent.PrevSiblingID,
						DatatypeID:    parent.DatatypeID,
						AuthorID:      parent.AuthorID,
						Status:        parent.Status,
						DateCreated:   parent.DateCreated,
						DateModified:  types.TimestampNow(),
					})
					if updateErr != nil {
						logger.Ferror("Failed to update parent during reorder up", updateErr)
					}
				}
			}

			// Update A: prev = B.prev, next = B
			_, aErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: a.ContentDataID,
				RouteID:       a.RouteID,
				ParentID:      a.ParentID,
				FirstChildID:  a.FirstChildID,
				NextSiblingID: types.NullableContentID{ID: b.ContentDataID, Valid: true},
				PrevSiblingID: b.PrevSiblingID,
				DatatypeID:    a.DatatypeID,
				AuthorID:      a.AuthorID,
				Status:        a.Status,
				DateCreated:   a.DateCreated,
				DateModified:  types.TimestampNow(),
			})
			if aErr != nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to update node: %v", aErr)}
			}

			// Update B: prev = A, next = A's original next
			_, bUpdateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: b.ContentDataID,
				RouteID:       b.RouteID,
				ParentID:      b.ParentID,
				FirstChildID:  b.FirstChildID,
				NextSiblingID: a.NextSiblingID,
				PrevSiblingID: types.NullableContentID{ID: a.ContentDataID, Valid: true},
				DatatypeID:    b.DatatypeID,
				AuthorID:      b.AuthorID,
				Status:        b.Status,
				DateCreated:   b.DateCreated,
				DateModified:  types.TimestampNow(),
			})
			if bUpdateErr != nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to update sibling: %v", bUpdateErr)}
			}

		} else {
			// Move down: swap A with its next sibling B
			if !a.NextSiblingID.Valid || a.NextSiblingID.ID == "" {
				return ActionResultMsg{Title: "Info", Message: "Already at bottom"}
			}
			bID := a.NextSiblingID.ID
			b, bErr := d.GetContentData(bID)
			if bErr != nil || b == nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Next sibling not found: %v", bErr)}
			}

			// Before: [C?] <-> A <-> B <-> [D?]
			// After:  [C?] <-> B <-> A <-> [D?]

			// If A has a prev (C), update C.NextSiblingID -> B
			if a.PrevSiblingID.Valid && a.PrevSiblingID.ID != "" {
				cID := a.PrevSiblingID.ID
				c, cErr := d.GetContentData(cID)
				if cErr == nil && c != nil {
					_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
						ContentDataID: c.ContentDataID,
						RouteID:       c.RouteID,
						ParentID:      c.ParentID,
						FirstChildID:  c.FirstChildID,
						NextSiblingID: types.NullableContentID{ID: b.ContentDataID, Valid: true},
						PrevSiblingID: c.PrevSiblingID,
						DatatypeID:    c.DatatypeID,
						AuthorID:      c.AuthorID,
						Status:        c.Status,
						DateCreated:   c.DateCreated,
						DateModified:  types.TimestampNow(),
					})
					if updateErr != nil {
						logger.Ferror("Failed to update C during reorder down", updateErr)
					}
				}
			}

			// If B has a next (D), update D.PrevSiblingID -> A
			if b.NextSiblingID.Valid && b.NextSiblingID.ID != "" {
				dID := b.NextSiblingID.ID
				dNode, dErr := d.GetContentData(dID)
				if dErr == nil && dNode != nil {
					_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
						ContentDataID: dNode.ContentDataID,
						RouteID:       dNode.RouteID,
						ParentID:      dNode.ParentID,
						FirstChildID:  dNode.FirstChildID,
						NextSiblingID: dNode.NextSiblingID,
						PrevSiblingID: types.NullableContentID{ID: a.ContentDataID, Valid: true},
						DatatypeID:    dNode.DatatypeID,
						AuthorID:      dNode.AuthorID,
						Status:        dNode.Status,
						DateCreated:   dNode.DateCreated,
						DateModified:  types.TimestampNow(),
					})
					if updateErr != nil {
						logger.Ferror("Failed to update D during reorder down", updateErr)
					}
				}
			}

			// If parent.FirstChildID == A, update to B
			if a.ParentID.Valid {
				parent, pErr := d.GetContentData(a.ParentID.ID)
				if pErr == nil && parent != nil && parent.FirstChildID.Valid && parent.FirstChildID.ID == a.ContentDataID {
					_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
						ContentDataID: parent.ContentDataID,
						RouteID:       parent.RouteID,
						ParentID:      parent.ParentID,
						FirstChildID:  types.NullableContentID{ID: b.ContentDataID, Valid: true},
						NextSiblingID: parent.NextSiblingID,
						PrevSiblingID: parent.PrevSiblingID,
						DatatypeID:    parent.DatatypeID,
						AuthorID:      parent.AuthorID,
						Status:        parent.Status,
						DateCreated:   parent.DateCreated,
						DateModified:  types.TimestampNow(),
					})
					if updateErr != nil {
						logger.Ferror("Failed to update parent during reorder down", updateErr)
					}
				}
			}

			// Update B: prev = A.prev, next = A
			_, bUpdateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: b.ContentDataID,
				RouteID:       b.RouteID,
				ParentID:      b.ParentID,
				FirstChildID:  b.FirstChildID,
				NextSiblingID: types.NullableContentID{ID: a.ContentDataID, Valid: true},
				PrevSiblingID: a.PrevSiblingID,
				DatatypeID:    b.DatatypeID,
				AuthorID:      b.AuthorID,
				Status:        b.Status,
				DateCreated:   b.DateCreated,
				DateModified:  types.TimestampNow(),
			})
			if bUpdateErr != nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to update sibling: %v", bUpdateErr)}
			}

			// Update A: prev = B, next = B's original next
			_, aErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: a.ContentDataID,
				RouteID:       a.RouteID,
				ParentID:      a.ParentID,
				FirstChildID:  a.FirstChildID,
				NextSiblingID: b.NextSiblingID,
				PrevSiblingID: types.NullableContentID{ID: b.ContentDataID, Valid: true},
				DatatypeID:    a.DatatypeID,
				AuthorID:      a.AuthorID,
				Status:        a.Status,
				DateCreated:   a.DateCreated,
				DateModified:  types.TimestampNow(),
			})
			if aErr != nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to update node: %v", aErr)}
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
			return ActionResultMsg{Title: "Error", Message: "Configuration not loaded"}
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
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to read source fields: %v", err)}
		}

		// Create new ContentData as sibling after source
		now := types.TimestampNow()
		newContent, createErr := d.CreateContentData(ctx, ac, db.CreateContentDataParams{
			RouteID:       source.RouteID,
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
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to create content copy: %v", createErr)}
		}

		if newContent.ContentDataID.IsZero() {
			return ActionResultMsg{Title: "Error", Message: "Failed to create content copy"}
		}

		// Update source.NextSiblingID -> new node
		_, sErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: source.ContentDataID,
			RouteID:       source.RouteID,
			ParentID:      source.ParentID,
			FirstChildID:  source.FirstChildID,
			NextSiblingID: types.NullableContentID{ID: newContent.ContentDataID, Valid: true},
			PrevSiblingID: source.PrevSiblingID,
			DatatypeID:    source.DatatypeID,
			AuthorID:      source.AuthorID,
			Status:        source.Status,
			DateCreated:   source.DateCreated,
			DateModified:  types.TimestampNow(),
		})
		if sErr != nil {
			logger.Ferror("Failed to update source next pointer after copy", sErr)
		}

		// If source had a next sibling (D), update D.PrevSiblingID -> new node
		if source.NextSiblingID.Valid && source.NextSiblingID.ID != "" {
			dID := source.NextSiblingID.ID
			dNode, dErr := d.GetContentData(dID)
			if dErr == nil && dNode != nil {
				_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
					ContentDataID: dNode.ContentDataID,
					RouteID:       dNode.RouteID,
					ParentID:      dNode.ParentID,
					FirstChildID:  dNode.FirstChildID,
					NextSiblingID: dNode.NextSiblingID,
					PrevSiblingID: types.NullableContentID{ID: newContent.ContentDataID, Valid: true},
					DatatypeID:    dNode.DatatypeID,
					AuthorID:      dNode.AuthorID,
					Status:        dNode.Status,
					DateCreated:   dNode.DateCreated,
					DateModified:  types.TimestampNow(),
				})
				if updateErr != nil {
					logger.Ferror("Failed to update next sibling prev pointer after copy", updateErr)
				}
			}
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
				logger.Ferror("Failed to list datatype fields for copy, falling back to source fields only", fieldListErr)
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
						logger.Ferror(fmt.Sprintf("Failed to copy field: %v", fieldErr), fieldErr)
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
							logger.Ferror(fmt.Sprintf("Failed to copy field: %v", fieldErr), fieldErr)
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
						logger.Ferror(fmt.Sprintf("Failed to copy field: %v", fieldErr), fieldErr)
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
			return ActionResultMsg{Title: "Error", Message: "Configuration not loaded"}
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
		_, pubErr := publishing.PublishContent(ctx, d, msg.ContentID, locale, userID, ac, retentionCap, dispatcher)
		if pubErr != nil {
			logger.Ferror(fmt.Sprintf("Failed to publish content %s", msg.ContentID), pubErr)
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

		unpubErr := publishing.UnpublishContent(ctx, d, msg.ContentID, locale, userID, ac, dispatcher)
		if unpubErr != nil {
			logger.Ferror(fmt.Sprintf("Failed to unpublish content %s", msg.ContentID), unpubErr)
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Unpublish failed: %v", unpubErr)}
		}

		logger.Finfo(fmt.Sprintf("Content %s unpublished", msg.ContentID))
		return UnpublishCompletedMsg{
			ContentID: msg.ContentID,
			RouteID:   msg.RouteID,
		}
	}
}

// LoadEnabledLocalesCmd fetches the list of enabled locales from the database.
func LoadEnabledLocalesCmd(d db.DbDriver) tea.Cmd {
	return func() tea.Msg {
		locales, err := d.ListEnabledLocales()
		if err != nil {
			return LocaleListMsg{Err: err}
		}
		if locales == nil {
			return LocaleListMsg{Locales: []db.Locale{}}
		}
		return LocaleListMsg{Locales: *locales}
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
			return ActionResultMsg{Title: "Error", Message: "Configuration not loaded"}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		versions, err := d.ListContentVersionsByContent(msg.ContentID)
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to list versions: %v", err)}
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
			return ActionResultMsg{Title: "Error", Message: "Configuration not loaded"}
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
			logger.Ferror(fmt.Sprintf("Failed to restore version for %s", msg.ContentID), err)
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Restore failed: %v", err)}
		}

		logger.Finfo(fmt.Sprintf("Content %s restored from version %s (%d fields)", msg.ContentID, msg.VersionID, result.FieldsRestored))
		return VersionRestoredMsg{
			ContentID:      msg.ContentID,
			RouteID:        msg.RouteID,
			FieldsRestored: result.FieldsRestored,
		}
	}
}

// HandlePluginAction dispatches a plugin action request to the appropriate handler.
func (m Model) HandlePluginAction(msg PluginActionRequestMsg) tea.Cmd {
	mgr := m.PluginManager
	adminUser := m.AdminUsername
	if mgr == nil {
		return func() tea.Msg {
			return PluginActionResultMsg{Title: "Error", Message: "Plugin manager not available"}
		}
	}

	switch msg.Action {
	case PluginActionEnable:
		return func() tea.Msg {
			if err := mgr.ActivatePlugin(context.Background(), msg.Name, adminUser); err != nil {
				return PluginActionResultMsg{
					Title:   "Error",
					Message: fmt.Sprintf("Failed to enable plugin %q: %v", msg.Name, err),
				}
			}
			return PluginActionCompleteMsg{Name: msg.Name, Action: PluginActionEnable}
		}
	case PluginActionDisable:
		return func() tea.Msg {
			if err := mgr.DeactivatePlugin(context.Background(), msg.Name); err != nil {
				return PluginActionResultMsg{
					Title:   "Error",
					Message: fmt.Sprintf("Failed to disable plugin %q: %v", msg.Name, err),
				}
			}
			return PluginActionCompleteMsg{Name: msg.Name, Action: PluginActionDisable}
		}
	case PluginActionReload:
		return func() tea.Msg {
			if err := mgr.ReloadPlugin(context.Background(), msg.Name); err != nil {
				return PluginActionResultMsg{
					Title:   "Error",
					Message: fmt.Sprintf("Failed to reload plugin %q: %v", msg.Name, err),
				}
			}
			return PluginActionCompleteMsg{Name: msg.Name, Action: PluginActionReload}
		}
	case PluginActionApproveRoutes:
		return func() tea.Msg {
			bridge := mgr.Bridge()
			if bridge == nil {
				return PluginActionResultMsg{Title: "Error", Message: "HTTP bridge not available"}
			}
			routes := bridge.ListRoutes()
			approved := 0
			for _, r := range routes {
				if r.PluginName != msg.Name || r.Approved {
					continue
				}
				if err := bridge.ApproveRoute(context.Background(), r.PluginName, r.Method, r.Path, adminUser); err != nil {
					return PluginActionResultMsg{
						Title:   "Error",
						Message: fmt.Sprintf("Failed to approve route %s %s: %v", r.Method, r.Path, err),
					}
				}
				approved++
			}
			return PluginRoutesApprovedMsg{Name: msg.Name, Count: approved}
		}
	case PluginActionApproveHooks:
		return func() tea.Msg {
			engine := mgr.HookEngine()
			if engine == nil {
				return PluginActionResultMsg{Title: "Error", Message: "Hook engine not available"}
			}
			hooks := engine.ListHooks()
			approved := 0
			for _, h := range hooks {
				if h.PluginName != msg.Name || h.Approved {
					continue
				}
				if err := engine.ApproveHook(context.Background(), h.PluginName, h.Event, h.Table, adminUser); err != nil {
					return PluginActionResultMsg{
						Title:   "Error",
						Message: fmt.Sprintf("Failed to approve hook %s:%s: %v", h.Event, h.Table, err),
					}
				}
				approved++
			}
			return PluginHooksApprovedMsg{Name: msg.Name, Count: approved}
		}
	default:
		return nil
	}
}

// FetchPendingRoutesForApprovalScreenCmd is a free-function variant for Screen
// implementations that don't have access to Model.
func FetchPendingRoutesForApprovalScreenCmd(mgr *plugin.Manager, pluginName string) tea.Cmd {
	if mgr == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Plugin manager not available"}
		}
	}
	return func() tea.Msg {
		bridge := mgr.Bridge()
		if bridge == nil {
			return ActionResultMsg{Title: "Error", Message: "HTTP bridge not available"}
		}
		allRoutes := bridge.ListRoutes()
		var pending []string
		for _, r := range allRoutes {
			if !r.Approved && r.PluginName == pluginName {
				pending = append(pending, fmt.Sprintf("%s %s", r.Method, r.Path))
			}
		}
		if len(pending) == 0 {
			return ActionResultMsg{
				Title:   "No Pending Routes",
				Message: fmt.Sprintf("Plugin '%s' has no unapproved routes.", pluginName),
			}
		}
		return ShowApproveAllRoutesDialogMsg{PluginName: pluginName, PendingRoutes: pending}
	}
}

// FetchPendingHooksForApprovalScreenCmd is a free-function variant for Screen
// implementations that don't have access to Model.
func FetchPendingHooksForApprovalScreenCmd(mgr *plugin.Manager, pluginName string) tea.Cmd {
	if mgr == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Plugin manager not available"}
		}
	}
	return func() tea.Msg {
		engine := mgr.HookEngine()
		if engine == nil {
			return ActionResultMsg{Title: "Error", Message: "Hook engine not available"}
		}
		allHooks := engine.ListHooks()
		var pending []string
		for _, h := range allHooks {
			if !h.Approved && h.PluginName == pluginName {
				pending = append(pending, fmt.Sprintf("%s.%s", h.Event, h.Table))
			}
		}
		if len(pending) == 0 {
			return ActionResultMsg{
				Title:   "No Pending Hooks",
				Message: fmt.Sprintf("Plugin '%s' has no unapproved hooks.", pluginName),
			}
		}
		return ShowApproveAllHooksDialogMsg{PluginName: pluginName, PendingHooks: pending}
	}
}

// FetchPendingRoutesForApprovalCmd fetches unapproved routes for a plugin and shows a confirmation dialog.
func (m Model) FetchPendingRoutesForApprovalCmd(pluginName string) tea.Cmd {
	return FetchPendingRoutesForApprovalScreenCmd(m.PluginManager, pluginName)
}

// FetchPendingHooksForApprovalCmd fetches unapproved hooks for a plugin and shows a confirmation dialog.
func (m Model) FetchPendingHooksForApprovalCmd(pluginName string) tea.Cmd {
	return FetchPendingHooksForApprovalScreenCmd(m.PluginManager, pluginName)
}

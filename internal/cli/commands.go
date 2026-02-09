package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/media"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/tree"
	"github.com/hegner123/modulacms/internal/utility"
)

type DatabaseCMD string

const (
	INSERT DatabaseCMD = "insert"
	SELECT DatabaseCMD = "select"
	UPDATE DatabaseCMD = "update"
	DELETE DatabaseCMD = "delete"
	BATCH  DatabaseCMD = "batch"
)

type FetchErrMsg struct {
	Error error
}

type ForeignKeyReference struct {
	From   string
	Table  string // Referenced table name.
	Column string // Referenced column name.
}

type FetchTables struct {
	Tables []string
}

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

func (m Model) GetContentField(node *string) []byte {
	row := m.TableState.Rows[m.Cursor]
	j, err := json.Marshal(row)
	if err != nil {
		m.Logger.Ferror("", err)
	}
	return j
}

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

func (m Model) GetContentInstances(c *config.Config) tea.Cmd {
	//	d := db.ConfigDB(*c)
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
			AuthorID:      types.NullableUserID{ID: authorID, Valid: true},
			Status:        types.ContentStatusDraft,
			DateCreated:   types.TimestampNow(),
			DateModified:  types.TimestampNow(),
			ParentID:      types.NullableContentID{}, // NULL - no parent initially
			FirstChildID:  sql.NullString{},          // NULL - no children initially
			NextSiblingID: sql.NullString{},          // NULL - no siblings initially
			PrevSiblingID: sql.NullString{},          // NULL - no siblings initially
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

		// Step 2: Create ContentFields (we have the ID now!)
		var failedFields []types.FieldID
		createdFields := 0

		for fieldID, value := range fieldValues {
			// Skip empty values
			if value == "" {
				continue
			}

			fieldResult, fieldErr := d.CreateContentField(ctx, ac, db.CreateContentFieldParams{
				ContentDataID: types.NullableContentID{ID: contentData.ContentDataID, Valid: true},
				FieldID:       types.NullableFieldID{ID: fieldID, Valid: true},
				FieldValue:    value,
				RouteID:       types.NullableRouteID{ID: routeID, Valid: true},
				AuthorID:      types.NullableUserID{ID: authorID, Valid: true},
				DateCreated:   types.TimestampNow(),
				DateModified:  types.TimestampNow(),
			})

			// Track failures
			if fieldErr != nil || fieldResult.ContentFieldID.IsZero() {
				failedFields = append(failedFields, fieldID)
			} else {
				createdFields++
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
			AuthorID:      types.NullableUserID{ID: authorID, Valid: true},
			Status:        types.ContentStatusDraft,
			DateCreated:   types.TimestampNow(),
			DateModified:  types.TimestampNow(),
			ParentID:      msg.ParentID,
			FirstChildID:  sql.NullString{}, // NULL - no children initially
			NextSiblingID: sql.NullString{}, // NULL - no siblings initially
			PrevSiblingID: sql.NullString{}, // NULL - no siblings initially
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

		// Step 2: Create ContentFields
		var failedFields []types.FieldID
		createdFields := 0

		for fieldID, value := range msg.FieldValues {
			// Skip empty values
			if value == "" {
				continue
			}

			fieldResult, fieldErr := d.CreateContentField(ctx, ac, db.CreateContentFieldParams{
				ContentDataID: types.NullableContentID{ID: contentData.ContentDataID, Valid: true},
				FieldID:       types.NullableFieldID{ID: fieldID, Valid: true},
				FieldValue:    value,
				RouteID:       types.NullableRouteID{ID: msg.RouteID, Valid: true},
				AuthorID:      types.NullableUserID{ID: authorID, Valid: true},
				DateCreated:   types.TimestampNow(),
				DateModified:  types.TimestampNow(),
			})

			// Track failures
			if fieldErr != nil || fieldResult.ContentFieldID.IsZero() {
				failedFields = append(failedFields, fieldID)
			} else {
				createdFields++
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

		// Get field definitions for this datatype via junction table (ordered by id)
		dtFields, err := d.ListDatatypeFieldByDatatypeID(msg.DatatypeID)
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

		// Build the existing fields list in order from junction table
		// dtFields is already ordered by id (ULID = creation order)
		existingFields := make([]ExistingContentField, 0)
		if dtFields != nil {
			for _, dtf := range *dtFields {
				if dtf.FieldID.IsZero() {
					continue
				}
				field, err := d.GetField(dtf.FieldID)
				if err != nil || field == nil {
					continue
				}
				ef := ExistingContentField{
					FieldID: field.FieldID,
					Label:   field.Label,
					Type:    string(field.Type),
					Value:   "",
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
					AuthorID:       types.NullableUserID{ID: authorID, Valid: !authorID.IsZero()},
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
					AuthorID:      types.NullableUserID{ID: authorID, Valid: !authorID.IsZero()},
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
		if content.FirstChildID.Valid && content.FirstChildID.String != "" {
			return ActionResultMsg{
				Title:   "Cannot Delete",
				Message: "This content has children. Delete child nodes first.",
			}
		}

		// Update sibling pointers before deletion
		// If this node has a previous sibling, point its next to our next
		if content.PrevSiblingID.Valid && content.PrevSiblingID.String != "" {
			prevSiblingID := types.ContentID(content.PrevSiblingID.String)
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
		if content.NextSiblingID.Valid && content.NextSiblingID.String != "" {
			nextSiblingID := types.ContentID(content.NextSiblingID.String)
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
				if parent.FirstChildID.Valid && parent.FirstChildID.String == string(contentID) {
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
		if source.PrevSiblingID.Valid && source.PrevSiblingID.String != "" {
			prevID := types.ContentID(source.PrevSiblingID.String)
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
		if source.NextSiblingID.Valid && source.NextSiblingID.String != "" {
			nextID := types.ContentID(source.NextSiblingID.String)
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
				if oldParent.FirstChildID.Valid && oldParent.FirstChildID.String == string(source.ContentDataID) {
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

		newPrevSiblingID := sql.NullString{} // default: no prev sibling

		if !target.FirstChildID.Valid || target.FirstChildID.String == "" {
			// Target has no children - source becomes first child
			targetParams := db.UpdateContentDataParams{
				ContentDataID: target.ContentDataID,
				RouteID:       target.RouteID,
				ParentID:      target.ParentID,
				FirstChildID:  sql.NullString{String: string(source.ContentDataID), Valid: true},
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
			currentID := types.ContentID(target.FirstChildID.String)
			for {
				current, walkErr := d.GetContentData(currentID)
				if walkErr != nil || current == nil {
					break
				}
				if !current.NextSiblingID.Valid || current.NextSiblingID.String == "" {
					// Found last sibling - update its NextSiblingID to source
					lastParams := db.UpdateContentDataParams{
						ContentDataID: current.ContentDataID,
						RouteID:       current.RouteID,
						ParentID:      current.ParentID,
						FirstChildID:  current.FirstChildID,
						NextSiblingID: sql.NullString{String: string(source.ContentDataID), Valid: true},
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
					newPrevSiblingID = sql.NullString{String: string(current.ContentDataID), Valid: true}
					break
				}
				currentID = types.ContentID(current.NextSiblingID.String)
			}
		}

		// === STEP 3: Update source node with new parent and cleared sibling pointers ===
		sourceParams := db.UpdateContentDataParams{
			ContentDataID: source.ContentDataID,
			RouteID:       source.RouteID,
			ParentID:      types.NullableContentID{ID: msg.TargetContentID, Valid: true},
			FirstChildID:  source.FirstChildID, // preserve children
			NextSiblingID: sql.NullString{},     // last child, no next
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

		user, err := d.CreateUser(ctx, ac, db.CreateUserParams{
			Username:     msg.Username,
			Name:         msg.Name,
			Email:        email,
			Hash:         "", // No password set via TUI
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

		// Fetch field definitions for labels via junction table (sorted by sort_order, id)
		var dtFields *[]db.DatatypeFields
		if datatypeID.Valid {
			dtFields, err = d.ListDatatypeFieldByDatatypeID(datatypeID.ID)
			if err != nil {
				dtFields = nil
			}
		}

		// Build field definition map: field_id -> (label, type, junction_id)
		type fieldInfo struct {
			Label           string
			Type            string
			DatatypeFieldID string
		}
		fieldDefMap := make(map[string]fieldInfo)
		var orderedFieldIDs []string
		if dtFields != nil {
			for _, dtf := range *dtFields {
				if dtf.FieldID.IsZero() {
					continue
				}
				field, fErr := d.GetField(dtf.FieldID)
				if fErr != nil || field == nil {
					continue
				}
				fieldDefMap[string(dtf.FieldID)] = fieldInfo{
					Label:           field.Label,
					Type:            string(field.Type),
					DatatypeFieldID: dtf.ID,
				}
				orderedFieldIDs = append(orderedFieldIDs, string(dtf.FieldID))
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

		// Build result in sort_order from junction table
		result := make([]ContentFieldDisplay, 0, len(orderedFieldIDs))
		for _, fid := range orderedFieldIDs {
			info := fieldDefMap[fid]
			display := ContentFieldDisplay{
				DatatypeFieldID: info.DatatypeFieldID,
				FieldID:         types.FieldID(fid),
				Label:           info.Label,
				Type:            info.Type,
			}
			if cf, ok := cfMap[fid]; ok {
				display.ContentFieldID = cf.ContentFieldID
				display.Value = cf.FieldValue
			}
			result = append(result, display)
		}

		return LoadContentFieldsMsg{Fields: result}
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
			AuthorID:       types.NullableUserID{ID: userID, Valid: !userID.IsZero()},
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
			AuthorID:      types.NullableUserID{ID: userID, Valid: !userID.IsZero()},
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

// HandleReorderField swaps sort_order between two junction records.
func (m Model) HandleReorderField(aID string, bID string, aOrder int64, bOrder int64, datatypeID types.NullableDatatypeID, contentID types.ContentID, routeID types.RouteID, direction string) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		if err := d.UpdateDatatypeFieldSortOrder(ctx, ac, aID, bOrder); err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to reorder: %v", err)}
		}
		if err := d.UpdateDatatypeFieldSortOrder(ctx, ac, bID, aOrder); err != nil {
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

		// Delete all junction records (datatypes_fields) for this datatype
		dtFields, err := d.ListDatatypeFieldByDatatypeID(msg.DatatypeID)
		if err != nil {
			logger.Ferror("Failed to list junction records", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to list field associations: %v", err),
			}
		}
		if dtFields != nil {
			for _, dtf := range *dtFields {
				if err := d.DeleteDatatypeField(ctx, ac, dtf.ID); err != nil {
					logger.Ferror(fmt.Sprintf("Failed to delete junction record %s", dtf.ID), err)
				}
			}
		}

		// Delete the datatype
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

// HandleDeleteField deletes a field and its junction record linking it to the datatype
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

		// Delete junction record linking this field to the datatype
		fieldJunctions, err := d.ListDatatypeFieldByFieldID(msg.FieldID)
		if err != nil {
			logger.Ferror("Failed to list junction records for field", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to list field associations: %v", err),
			}
		}

		if fieldJunctions != nil {
			for _, dtf := range *fieldJunctions {
				if !dtf.DatatypeID.IsZero() && dtf.DatatypeID == msg.DatatypeID {
					if err := d.DeleteDatatypeField(ctx, ac, dtf.ID); err != nil {
						logger.Ferror(fmt.Sprintf("Failed to delete junction record %s", dtf.ID), err)
					}
				}
			}
		}

		// Delete the field itself
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

// HandleMediaUpload runs the media upload pipeline asynchronously.
func (m Model) HandleMediaUpload(msg MediaUploadStartMsg) tea.Cmd {
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}
	cfg := m.Config
	return func() tea.Msg {
		filename := filepath.Base(msg.FilePath)
		baseName := strings.TrimSuffix(filename, filepath.Ext(filename))

		logger.Finfo(fmt.Sprintf("Starting media upload: %s", filename))

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
			if !a.PrevSiblingID.Valid || a.PrevSiblingID.String == "" {
				return ActionResultMsg{Title: "Info", Message: "Already at top"}
			}
			bID := types.ContentID(a.PrevSiblingID.String)
			b, bErr := d.GetContentData(bID)
			if bErr != nil || b == nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Previous sibling not found: %v", bErr)}
			}

			// Before: [C?] <-> B <-> A <-> [D?]
			// After:  [C?] <-> A <-> B <-> [D?]

			// If B has a prev (C), update C.NextSiblingID -> A
			if b.PrevSiblingID.Valid && b.PrevSiblingID.String != "" {
				cID := types.ContentID(b.PrevSiblingID.String)
				c, cErr := d.GetContentData(cID)
				if cErr == nil && c != nil {
					_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
						ContentDataID: c.ContentDataID,
						RouteID:       c.RouteID,
						ParentID:      c.ParentID,
						FirstChildID:  c.FirstChildID,
						NextSiblingID: sql.NullString{String: string(a.ContentDataID), Valid: true},
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
			if a.NextSiblingID.Valid && a.NextSiblingID.String != "" {
				dID := types.ContentID(a.NextSiblingID.String)
				dNode, dErr := d.GetContentData(dID)
				if dErr == nil && dNode != nil {
					_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
						ContentDataID: dNode.ContentDataID,
						RouteID:       dNode.RouteID,
						ParentID:      dNode.ParentID,
						FirstChildID:  dNode.FirstChildID,
						NextSiblingID: dNode.NextSiblingID,
						PrevSiblingID: sql.NullString{String: string(b.ContentDataID), Valid: true},
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
				if pErr == nil && parent != nil && parent.FirstChildID.Valid && parent.FirstChildID.String == string(b.ContentDataID) {
					_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
						ContentDataID: parent.ContentDataID,
						RouteID:       parent.RouteID,
						ParentID:      parent.ParentID,
						FirstChildID:  sql.NullString{String: string(a.ContentDataID), Valid: true},
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
				NextSiblingID: sql.NullString{String: string(b.ContentDataID), Valid: true},
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
				PrevSiblingID: sql.NullString{String: string(a.ContentDataID), Valid: true},
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
			if !a.NextSiblingID.Valid || a.NextSiblingID.String == "" {
				return ActionResultMsg{Title: "Info", Message: "Already at bottom"}
			}
			bID := types.ContentID(a.NextSiblingID.String)
			b, bErr := d.GetContentData(bID)
			if bErr != nil || b == nil {
				return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Next sibling not found: %v", bErr)}
			}

			// Before: [C?] <-> A <-> B <-> [D?]
			// After:  [C?] <-> B <-> A <-> [D?]

			// If A has a prev (C), update C.NextSiblingID -> B
			if a.PrevSiblingID.Valid && a.PrevSiblingID.String != "" {
				cID := types.ContentID(a.PrevSiblingID.String)
				c, cErr := d.GetContentData(cID)
				if cErr == nil && c != nil {
					_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
						ContentDataID: c.ContentDataID,
						RouteID:       c.RouteID,
						ParentID:      c.ParentID,
						FirstChildID:  c.FirstChildID,
						NextSiblingID: sql.NullString{String: string(b.ContentDataID), Valid: true},
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
			if b.NextSiblingID.Valid && b.NextSiblingID.String != "" {
				dID := types.ContentID(b.NextSiblingID.String)
				dNode, dErr := d.GetContentData(dID)
				if dErr == nil && dNode != nil {
					_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
						ContentDataID: dNode.ContentDataID,
						RouteID:       dNode.RouteID,
						ParentID:      dNode.ParentID,
						FirstChildID:  dNode.FirstChildID,
						NextSiblingID: dNode.NextSiblingID,
						PrevSiblingID: sql.NullString{String: string(a.ContentDataID), Valid: true},
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
				if pErr == nil && parent != nil && parent.FirstChildID.Valid && parent.FirstChildID.String == string(a.ContentDataID) {
					_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
						ContentDataID: parent.ContentDataID,
						RouteID:       parent.RouteID,
						ParentID:      parent.ParentID,
						FirstChildID:  sql.NullString{String: string(b.ContentDataID), Valid: true},
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
				NextSiblingID: sql.NullString{String: string(a.ContentDataID), Valid: true},
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
				PrevSiblingID: sql.NullString{String: string(b.ContentDataID), Valid: true},
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
			FirstChildID:  sql.NullString{},                                                // no children (flat copy)
			NextSiblingID: source.NextSiblingID,                                             // take source's next
			PrevSiblingID: sql.NullString{String: string(source.ContentDataID), Valid: true}, // prev = source
			DatatypeID:    source.DatatypeID,
			AuthorID:      types.NullableUserID{ID: userID, Valid: !userID.IsZero()},
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
			NextSiblingID: sql.NullString{String: string(newContent.ContentDataID), Valid: true},
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
		if source.NextSiblingID.Valid && source.NextSiblingID.String != "" {
			dID := types.ContentID(source.NextSiblingID.String)
			dNode, dErr := d.GetContentData(dID)
			if dErr == nil && dNode != nil {
				_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
					ContentDataID: dNode.ContentDataID,
					RouteID:       dNode.RouteID,
					ParentID:      dNode.ParentID,
					FirstChildID:  dNode.FirstChildID,
					NextSiblingID: dNode.NextSiblingID,
					PrevSiblingID: sql.NullString{String: string(newContent.ContentDataID), Valid: true},
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

		// Copy fields
		fieldCount := 0
		if sourceFields != nil {
			for _, field := range *sourceFields {
				_, fieldErr := d.CreateContentField(ctx, ac, db.CreateContentFieldParams{
					RouteID:       field.RouteID,
					ContentDataID: types.NullableContentID{ID: newContent.ContentDataID, Valid: true},
					FieldID:       field.FieldID,
					FieldValue:    field.FieldValue,
					AuthorID:      types.NullableUserID{ID: userID, Valid: !userID.IsZero()},
					DateCreated:   now,
					DateModified:  now,
				})
				if fieldErr != nil {
					logger.Ferror(fmt.Sprintf("Failed to copy field: %v", fieldErr), fieldErr)
				}
				fieldCount++
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

// TogglePublishCmd creates a command to toggle content publish status.
func TogglePublishCmd(contentID types.ContentID, routeID types.RouteID) tea.Cmd {
	return func() tea.Msg {
		return TogglePublishRequestMsg{
			ContentID: contentID,
			RouteID:   routeID,
		}
	}
}

// HandleTogglePublish toggles a content node between draft and published status.
func (m Model) HandleTogglePublish(msg TogglePublishRequestMsg) tea.Cmd {
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

		content, err := d.GetContentData(msg.ContentID)
		if err != nil || content == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Content not found: %v", err)}
		}

		// Toggle status
		newStatus := types.ContentStatusPublished
		if content.Status == types.ContentStatusPublished {
			newStatus = types.ContentStatusDraft
		}

		_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: content.ContentDataID,
			RouteID:       content.RouteID,
			ParentID:      content.ParentID,
			FirstChildID:  content.FirstChildID,
			NextSiblingID: content.NextSiblingID,
			PrevSiblingID: content.PrevSiblingID,
			DatatypeID:    content.DatatypeID,
			AuthorID:      content.AuthorID,
			Status:        newStatus,
			DateCreated:   content.DateCreated,
			DateModified:  types.TimestampNow(),
		})
		if updateErr != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to update status: %v", updateErr)}
		}

		logger.Finfo(fmt.Sprintf("Content %s status changed to %s", msg.ContentID, newStatus))
		return ContentPublishToggledMsg{
			ContentID: msg.ContentID,
			RouteID:   msg.RouteID,
			NewStatus: newStatus,
		}
	}
}

// ArchiveContentCmd creates a command to toggle content archive status.
func ArchiveContentCmd(contentID types.ContentID, routeID types.RouteID) tea.Cmd {
	return func() tea.Msg {
		return ArchiveContentRequestMsg{
			ContentID: contentID,
			RouteID:   routeID,
		}
	}
}

// HandleArchiveContent toggles a content node between archived and draft status.
// If currently archived, reverts to draft. Otherwise, sets to archived.
func (m Model) HandleArchiveContent(msg ArchiveContentRequestMsg) tea.Cmd {
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

		content, err := d.GetContentData(msg.ContentID)
		if err != nil || content == nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Content not found: %v", err)}
		}

		// Toggle archive: if archived -> draft, otherwise -> archived
		newStatus := types.ContentStatusArchived
		if content.Status == types.ContentStatusArchived {
			newStatus = types.ContentStatusDraft
		}

		_, updateErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: content.ContentDataID,
			RouteID:       content.RouteID,
			ParentID:      content.ParentID,
			FirstChildID:  content.FirstChildID,
			NextSiblingID: content.NextSiblingID,
			PrevSiblingID: content.PrevSiblingID,
			DatatypeID:    content.DatatypeID,
			AuthorID:      content.AuthorID,
			Status:        newStatus,
			DateCreated:   content.DateCreated,
			DateModified:  types.TimestampNow(),
		})
		if updateErr != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to update status: %v", updateErr)}
		}

		logger.Finfo(fmt.Sprintf("Content %s status changed to %s", msg.ContentID, newStatus))
		return ContentArchivedMsg{
			ContentID: msg.ContentID,
			RouteID:   msg.RouteID,
			NewStatus: newStatus,
		}
	}
}

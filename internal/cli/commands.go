package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
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
	// Using secure query builder
	sqb := db.NewSecureQueryBuilder(con)
	query, args, err := sqb.SecureBuildInsertQuery(string(table), valuesMap)
	if err != nil {
		return tea.Batch(
			LogMessageCmd(err.Error()),
			LogMessageCmd(fmt.Sprintln(valuesMap)),
		)
	}
	res, err := sqb.SecureExecuteModifyQuery(query, args)
	if err != nil {
		return LogMessageCmd(err.Error())
	}


	return tea.Batch(
		DbResultCmd(res, string(table)),
	)
}

func (m Model) DatabaseUpdate(c *config.Config, table db.DBTable) tea.Cmd {
	id := m.GetCurrentRowId()
	d := db.ConfigDB(*c)

	con, _, err := d.GetConnection()
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	valuesMap := make(map[string]any, 0)
	for i, v := range m.FormState.FormValues {
		valuesMap[m.TableState.Headers[i]] = *v
	}

	// Using secure query builder
	sqb := db.NewSecureQueryBuilder(con)
	query, args, err := sqb.SecureBuildUpdateQuery(string(table), id, valuesMap)
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	res, err := sqb.SecureExecuteModifyQuery(query, args)
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	// Reset the form values after update
	m.FormState.FormValues = nil

	m.Logger.Finfo("CLI Update successful", nil)
	return DbResultCmd(res, string(table))
}
func (m Model) DatabaseGet(c *config.Config, source FetchSource, table db.DBTable, id int64) tea.Cmd {
	d := db.ConfigDB(*c)

	con, _, err := d.GetConnection()
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	// Using secure query builder
	sqb := db.NewSecureQueryBuilder(con)
	query, args, err := sqb.SecureBuildSelectQuery(string(table), id)
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	r, err := sqb.SecureExecuteSelectQuery(query, args)
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

	// Using secure query builder
	sqb := db.NewSecureQueryBuilder(con)
	query, args, err := sqb.SecureBuildListQuery(string(table))
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	r, err := sqb.SecureExecuteSelectQuery(query, args)
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

	// Using secure query builder
	sqb := db.NewSecureQueryBuilder(con)
	query, args, err := sqb.SecureBuildSelectWithColumnsQuery(string(table), columns, whereColumn, value)
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	r, err := sqb.SecureExecuteSelectQuery(query, args)
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

	// Using secure query builder
	sqb := db.NewSecureQueryBuilder(con)
	query, args, err := sqb.SecureBuildDeleteQuery(string(table), id)
	if err != nil {
		return LogMessageCmd(err.Error())
	}
	res, err := sqb.SecureExecuteModifyQuery(query, args)
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
	out := db.LogRouteTree("GetFullTree", res)
	return GetFullTreeResCMD(out, *res)
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
	return func() tea.Msg {
		d := db.ConfigDB(*c)

		// Debug logging
		logger.Finfo(fmt.Sprintf("Creating ContentData: DatatypeID=%s, RouteID=%s, AuthorID=%s", datatypeID, routeID, authorID))

		// Step 1: Create ContentData using typed DbDriver method
		contentData := d.CreateContentData(db.CreateContentDataParams{
			DatatypeID:    types.NullableDatatypeID{ID: datatypeID, Valid: true},
			RouteID:       types.NullableRouteID{ID: routeID, Valid: true},
			AuthorID:      types.NullableUserID{ID: authorID, Valid: true},
			DateCreated:   types.TimestampNow(),
			DateModified:  types.TimestampNow(),
			ParentID:      types.NullableContentID{}, // NULL - no parent initially
			FirstChildID:  sql.NullString{},          // NULL - no children initially
			NextSiblingID: sql.NullString{},          // NULL - no siblings initially
			PrevSiblingID: sql.NullString{},          // NULL - no siblings initially
		})

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

			fieldResult := d.CreateContentField(db.CreateContentFieldParams{
				ContentDataID: types.NullableContentID{ID: contentData.ContentDataID, Valid: true},
				FieldID:       types.NullableFieldID{ID: fieldID, Valid: true},
				FieldValue:    value,
				RouteID:       types.NullableRouteID{ID: routeID, Valid: true},
				AuthorID:      types.NullableUserID{ID: authorID, Valid: true},
				DateCreated:   types.TimestampNow(),
				DateModified:  types.TimestampNow(),
			})

			// Track failures
			if fieldResult.ContentFieldID.IsZero() {
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
	return func() tea.Msg {
		d := db.ConfigDB(*m.Config)

		// Debug logging
		logger.Finfo(fmt.Sprintf("Creating ContentData from dialog: DatatypeID=%s, RouteID=%s, AuthorID=%s, HasParent=%v",
			msg.DatatypeID, msg.RouteID, authorID, msg.ParentID.Valid))

		// Step 1: Create ContentData using typed DbDriver method
		contentData := d.CreateContentData(db.CreateContentDataParams{
			DatatypeID:    types.NullableDatatypeID{ID: msg.DatatypeID, Valid: true},
			RouteID:       types.NullableRouteID{ID: msg.RouteID, Valid: true},
			AuthorID:      types.NullableUserID{ID: authorID, Valid: true},
			DateCreated:   types.TimestampNow(),
			DateModified:  types.TimestampNow(),
			ParentID:      msg.ParentID,
			FirstChildID:  sql.NullString{}, // NULL - no children initially
			NextSiblingID: sql.NullString{}, // NULL - no siblings initially
			PrevSiblingID: sql.NullString{}, // NULL - no siblings initially
		})

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

			fieldResult := d.CreateContentField(db.CreateContentFieldParams{
				ContentDataID: types.NullableContentID{ID: contentData.ContentDataID, Valid: true},
				FieldID:       types.NullableFieldID{ID: fieldID, Valid: true},
				FieldValue:    value,
				RouteID:       types.NullableRouteID{ID: msg.RouteID, Valid: true},
				AuthorID:      types.NullableUserID{ID: authorID, Valid: true},
				DateCreated:   types.TimestampNow(),
				DateModified:  types.TimestampNow(),
			})

			// Track failures
			if fieldResult.ContentFieldID.IsZero() {
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
		datatypeID := types.NullableDatatypeID{ID: msg.DatatypeID, Valid: true}
		dtFields, err := d.ListDatatypeFieldByDatatypeID(datatypeID)
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
				if !dtf.FieldID.Valid {
					continue
				}
				field, err := d.GetField(dtf.FieldID.ID)
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
				_, err := d.UpdateContentField(db.UpdateContentFieldParams{
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
				fieldResult := d.CreateContentField(db.CreateContentFieldParams{
					ContentDataID: contentDataID,
					FieldID:       types.NullableFieldID{ID: fieldID, Valid: true},
					FieldValue:    value,
					RouteID:       types.NullableRouteID{ID: msg.RouteID, Valid: true},
					AuthorID:      types.NullableUserID{ID: authorID, Valid: !authorID.IsZero()},
					DateCreated:   types.TimestampNow(),
					DateModified:  types.TimestampNow(),
				})
				if fieldResult.ContentFieldID.IsZero() {
					logger.Ferror(fmt.Sprintf("Failed to create field %s", fieldID), nil)
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
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)

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
					DateCreated:   prevSibling.DateCreated,
					DateModified:  types.TimestampNow(),
				}
				if _, err := d.UpdateContentData(updateParams); err != nil {
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
					DateCreated:   nextSibling.DateCreated,
					DateModified:  types.TimestampNow(),
				}
				if _, err := d.UpdateContentData(updateParams); err != nil {
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
						DateCreated:   parent.DateCreated,
						DateModified:  types.TimestampNow(),
					}
					if _, err := d.UpdateContentData(updateParams); err != nil {
						logger.Ferror("Failed to update parent first_child", err)
					}
				}
			}
		}

		// Delete the content data (content_fields will cascade delete)
		if err := d.DeleteContentData(contentID); err != nil {
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
				Stats:    &LoadStats{},
				RootNode: nil,
			}
		}

		logger.Finfo(fmt.Sprintf("GetContentTreeByRoute returned %d rows for route %s", len(*rows), routeID))

		if len(*rows) == 0 {
			logger.Finfo(fmt.Sprintf("No rows returned for route %s", routeID))
			return TreeLoadedMsg{
				RouteID:  routeID,
				Stats:    &LoadStats{},
				RootNode: nil,
			}
		}

		// Create new tree root and load from rows
		newRoot := NewTreeRoot()
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

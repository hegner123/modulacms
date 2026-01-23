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

	utility.DefaultLogger.Finfo("CLI Update successful", nil)
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
		utility.DefaultLogger.Ferror("", err)
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
	return func() tea.Msg {
		d := db.ConfigDB(*c)

		// Debug logging
		utility.DefaultLogger.Finfo(fmt.Sprintf("Creating ContentData: DatatypeID=%s, RouteID=%s, AuthorID=%s", datatypeID, routeID, authorID))

		// Step 1: Create ContentData using typed DbDriver method
		contentData := d.CreateContentData(db.CreateContentDataParams{
			DatatypeID:    types.NullableDatatypeID{ID: datatypeID, Valid: true},
			RouteID:       types.NullableRouteID{ID: routeID, Valid: true},
			AuthorID:      types.NullableUserID{ID: authorID, Valid: true},
			DateCreated:   types.TimestampNow(),
			DateModified:  types.TimestampNow(),
			ParentID:      types.NullableContentID{}, // NULL - no parent initially
			FirstChildID:  sql.NullInt64{},           // NULL - no children initially
			NextSiblingID: sql.NullInt64{},           // NULL - no siblings initially
			PrevSiblingID: sql.NullInt64{},           // NULL - no siblings initially
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

// ReloadContentTree fetches tree data from database and loads it into the Root
func (m Model) ReloadContentTree(c *config.Config, routeID types.RouteID) tea.Cmd {
	return func() tea.Msg {
		d := db.ConfigDB(*c)

		// Fetch tree data from database
		nullableRouteID := types.NullableRouteID{ID: routeID, Valid: !routeID.IsZero()}
		rows, err := d.GetContentTreeByRoute(nullableRouteID)
		if err != nil {
			utility.DefaultLogger.Ferror(fmt.Sprintf("GetContentTreeByRoute error for route %s", routeID), err)
			return FetchErrMsg{Error: fmt.Errorf("failed to fetch content tree: %w", err)}
		}

		if rows == nil {
			utility.DefaultLogger.Finfo(fmt.Sprintf("GetContentTreeByRoute returned nil rows for route %s", routeID))
			return TreeLoadedMsg{
				RouteID:  routeID,
				Stats:    &LoadStats{},
				RootNode: nil,
			}
		}

		utility.DefaultLogger.Finfo(fmt.Sprintf("GetContentTreeByRoute returned %d rows for route %s", len(*rows), routeID))

		if len(*rows) == 0 {
			utility.DefaultLogger.Finfo(fmt.Sprintf("No rows returned for route %s", routeID))
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

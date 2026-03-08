package tui

import (
	"context"
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
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
	col, val := m.GetCurrentRowPK()
	if col == "" || val == "" {
		return LogMessageCmd("Delete failed: no primary key selected")
	}
	d := db.ConfigDB(*c)

	con, _, err := d.GetConnection()
	if err != nil {
		return LogMessageCmd(err.Error())
	}

	dialect := db.DialectFromString(string(c.Db_Driver))
	res, err := db.QDelete(context.Background(), con, dialect, db.DeleteParams{
		Table: string(table),
		Where: map[string]any{col: val},
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

// =============================================================================
// USER CRUD
// =============================================================================

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

// =============================================================================
// ENTITY DELETE HANDLERS
// =============================================================================

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

// =============================================================================
// LOCALE
// =============================================================================

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

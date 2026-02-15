// White-box tests for table.go: wrapper structs, mapper methods
// across all three database drivers (SQLite, MySQL, PostgreSQL), string mapping
// for TUI display, and audited command struct accessors.
//
// White-box access is needed because:
//   - Audited command structs have unexported fields (ctx, auditCtx, params, conn)
//     that can only be constructed through the Database/MysqlDatabase/PsqlDatabase
//     factory methods, which require access to the package internals.
//   - We verify that the SQLiteRecorder, MysqlRecorder, and PsqlRecorder package-level
//     vars are correctly wired into command constructors.
package db

import (
	"context"
	"testing"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// --- Test data helpers ---

// tableTestFixture returns a fully populated Tables and its individual parts.
func tableTestFixture() (Tables, string, string, types.NullableUserID) {
	id := string(types.NewTableID())
	label := "blog_posts"
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	t := Tables{
		ID:       id,
		Label:    label,
		AuthorID: authorID,
	}
	return t, id, label, authorID
}

// --- MapStringTable tests ---

func TestMapStringTable_AllFields(t *testing.T) {
	t.Parallel()
	tbl, id, label, authorID := tableTestFixture()

	got := MapStringTable(tbl)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ID", got.ID, id},
		{"Label", got.Label, label},
		{"AuthorID", got.AuthorID, authorID.String()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestMapStringTable_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringTable(Tables{})

	if got.ID != "" {
		t.Errorf("ID = %q, want empty string", got.ID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	// Zero-value NullableUserID has Valid=false, so String() returns "null"
	if got.AuthorID != "null" {
		t.Errorf("AuthorID = %q, want %q", got.AuthorID, "null")
	}
}

func TestMapStringTable_NullAuthorID(t *testing.T) {
	t.Parallel()
	tbl := Tables{
		ID:       "some-id",
		Label:    "pages",
		AuthorID: types.NullableUserID{Valid: false},
	}
	got := MapStringTable(tbl)
	if got.AuthorID != "null" {
		t.Errorf("AuthorID = %q, want %q", got.AuthorID, "null")
	}
}

func TestMapStringTable_ValidAuthorID(t *testing.T) {
	t.Parallel()
	userID := types.NewUserID()
	tbl := Tables{
		AuthorID: types.NullableUserID{ID: userID, Valid: true},
	}
	got := MapStringTable(tbl)
	if got.AuthorID != userID.String() {
		t.Errorf("AuthorID = %q, want %q", got.AuthorID, userID.String())
	}
}

// --- SQLite Database.MapTable tests ---

func TestDatabase_MapTable_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdb.Tables{
		ID:       "tbl-001",
		Label:    "products",
		AuthorID: authorID,
	}

	got := d.MapTable(input)

	if got.ID != "tbl-001" {
		t.Errorf("ID = %q, want %q", got.ID, "tbl-001")
	}
	if got.Label != "products" {
		t.Errorf("Label = %q, want %q", got.Label, "products")
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
}

func TestDatabase_MapTable_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapTable(mdb.Tables{})

	if got.ID != "" {
		t.Errorf("ID = %q, want empty string", got.ID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.AuthorID.Valid {
		t.Errorf("AuthorID.Valid = true, want false")
	}
}

// --- SQLite Database.MapCreateTableParams tests ---

func TestDatabase_MapCreateTableParams_GeneratesNewID(t *testing.T) {
	t.Parallel()
	d := Database{}

	input := CreateTableParams{
		Label: "categories",
	}

	got := d.MapCreateTableParams(input)

	if got.ID == "" {
		t.Fatal("expected non-empty ID to be generated")
	}
	if got.Label != "categories" {
		t.Errorf("Label = %q, want %q", got.Label, "categories")
	}
}

func TestDatabase_MapCreateTableParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}

	got1 := d.MapCreateTableParams(CreateTableParams{})
	got2 := d.MapCreateTableParams(CreateTableParams{})

	if got1.ID == got2.ID {
		t.Error("two calls produced identical IDs")
	}
}

func TestDatabase_MapCreateTableParams_ZeroInput(t *testing.T) {
	t.Parallel()
	d := Database{}

	got := d.MapCreateTableParams(CreateTableParams{})

	if got.ID == "" {
		t.Fatal("expected non-empty ID even with zero input")
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
}

// --- SQLite Database.MapUpdateTableParams tests ---

func TestDatabase_MapUpdateTableParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}

	input := UpdateTableParams{
		Label: "updated_label",
		ID:    "tbl-update-001",
	}

	got := d.MapUpdateTableParams(input)

	if got.Label != "updated_label" {
		t.Errorf("Label = %q, want %q", got.Label, "updated_label")
	}
	if got.ID != "tbl-update-001" {
		t.Errorf("ID = %q, want %q", got.ID, "tbl-update-001")
	}
}

func TestDatabase_MapUpdateTableParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapUpdateTableParams(UpdateTableParams{})

	if got.ID != "" {
		t.Errorf("ID = %q, want empty string", got.ID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
}

// --- MySQL MysqlDatabase.MapTable tests ---

func TestMysqlDatabase_MapTable_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdbm.Tables{
		ID:       "tbl-mysql-001",
		Label:    "mysql_products",
		AuthorID: authorID,
	}

	got := d.MapTable(input)

	if got.ID != "tbl-mysql-001" {
		t.Errorf("ID = %q, want %q", got.ID, "tbl-mysql-001")
	}
	if got.Label != "mysql_products" {
		t.Errorf("Label = %q, want %q", got.Label, "mysql_products")
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
}

func TestMysqlDatabase_MapTable_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapTable(mdbm.Tables{})

	if got.ID != "" {
		t.Errorf("ID = %q, want empty string", got.ID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.AuthorID.Valid {
		t.Errorf("AuthorID.Valid = true, want false")
	}
}

// --- MySQL MysqlDatabase.MapCreateTableParams tests ---

func TestMysqlDatabase_MapCreateTableParams(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	input := CreateTableParams{
		Label: "mysql_categories",
	}

	got := d.MapCreateTableParams(input)

	if got.ID == "" {
		t.Fatal("expected non-empty ID to be generated")
	}
	if got.Label != "mysql_categories" {
		t.Errorf("Label = %q, want %q", got.Label, "mysql_categories")
	}
}

func TestMysqlDatabase_MapCreateTableParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	got1 := d.MapCreateTableParams(CreateTableParams{})
	got2 := d.MapCreateTableParams(CreateTableParams{})

	if got1.ID == got2.ID {
		t.Error("two calls produced identical IDs")
	}
}

// --- MySQL MysqlDatabase.MapUpdateTableParams tests ---

func TestMysqlDatabase_MapUpdateTableParams(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	input := UpdateTableParams{
		Label: "mysql_updated",
		ID:    "tbl-mysql-update",
	}

	got := d.MapUpdateTableParams(input)

	if got.Label != "mysql_updated" {
		t.Errorf("Label = %q, want %q", got.Label, "mysql_updated")
	}
	if got.ID != "tbl-mysql-update" {
		t.Errorf("ID = %q, want %q", got.ID, "tbl-mysql-update")
	}
}

// --- PostgreSQL PsqlDatabase.MapTable tests ---

func TestPsqlDatabase_MapTable_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	input := mdbp.Tables{
		ID:       "tbl-psql-001",
		Label:    "psql_products",
		AuthorID: authorID,
	}

	got := d.MapTable(input)

	if got.ID != "tbl-psql-001" {
		t.Errorf("ID = %q, want %q", got.ID, "tbl-psql-001")
	}
	if got.Label != "psql_products" {
		t.Errorf("Label = %q, want %q", got.Label, "psql_products")
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID = %v, want %v", got.AuthorID, authorID)
	}
}

func TestPsqlDatabase_MapTable_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapTable(mdbp.Tables{})

	if got.ID != "" {
		t.Errorf("ID = %q, want empty string", got.ID)
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want empty string", got.Label)
	}
	if got.AuthorID.Valid {
		t.Errorf("AuthorID.Valid = true, want false")
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateTableParams tests ---

func TestPsqlDatabase_MapCreateTableParams(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	input := CreateTableParams{
		Label: "psql_categories",
	}

	got := d.MapCreateTableParams(input)

	if got.ID == "" {
		t.Fatal("expected non-empty ID to be generated")
	}
	if got.Label != "psql_categories" {
		t.Errorf("Label = %q, want %q", got.Label, "psql_categories")
	}
}

func TestPsqlDatabase_MapCreateTableParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	got1 := d.MapCreateTableParams(CreateTableParams{})
	got2 := d.MapCreateTableParams(CreateTableParams{})

	if got1.ID == got2.ID {
		t.Error("two calls produced identical IDs")
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateTableParams tests ---

func TestPsqlDatabase_MapUpdateTableParams(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	input := UpdateTableParams{
		Label: "psql_updated",
		ID:    "tbl-psql-update",
	}

	got := d.MapUpdateTableParams(input)

	if got.Label != "psql_updated" {
		t.Errorf("Label = %q, want %q", got.Label, "psql_updated")
	}
	if got.ID != "tbl-psql-update" {
		t.Errorf("ID = %q, want %q", got.ID, "tbl-psql-update")
	}
}

// --- Cross-database mapper consistency ---

func TestCrossDatabaseMapTable_Consistency(t *testing.T) {
	t.Parallel()
	authorID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	sqliteInput := mdb.Tables{
		ID:       "cross-tbl-001",
		Label:    "cross_label",
		AuthorID: authorID,
	}
	mysqlInput := mdbm.Tables{
		ID:       "cross-tbl-001",
		Label:    "cross_label",
		AuthorID: authorID,
	}
	psqlInput := mdbp.Tables{
		ID:       "cross-tbl-001",
		Label:    "cross_label",
		AuthorID: authorID,
	}

	sqliteResult := Database{}.MapTable(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapTable(mysqlInput)
	psqlResult := PsqlDatabase{}.MapTable(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

func TestCrossDatabaseMapTable_ConsistencyWithNullAuthor(t *testing.T) {
	t.Parallel()
	nullAuthor := types.NullableUserID{Valid: false}

	sqliteInput := mdb.Tables{
		ID:       "null-author-001",
		Label:    "null_test",
		AuthorID: nullAuthor,
	}
	mysqlInput := mdbm.Tables{
		ID:       "null-author-001",
		Label:    "null_test",
		AuthorID: nullAuthor,
	}
	psqlInput := mdbp.Tables{
		ID:       "null-author-001",
		Label:    "null_test",
		AuthorID: nullAuthor,
	}

	sqliteResult := Database{}.MapTable(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapTable(mysqlInput)
	psqlResult := PsqlDatabase{}.MapTable(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL differ with null author:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL differ with null author:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateTableParams - ID generation ---

func TestCrossDatabaseMapCreateTableParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()

	input := CreateTableParams{
		Label: "cross_create",
	}

	sqliteResult := Database{}.MapCreateTableParams(input)
	mysqlResult := MysqlDatabase{}.MapCreateTableParams(input)
	psqlResult := PsqlDatabase{}.MapCreateTableParams(input)

	if sqliteResult.ID == "" {
		t.Error("SQLite: expected non-empty generated ID")
	}
	if mysqlResult.ID == "" {
		t.Error("MySQL: expected non-empty generated ID")
	}
	if psqlResult.ID == "" {
		t.Error("PostgreSQL: expected non-empty generated ID")
	}

	// Each call should generate a unique ID
	if sqliteResult.ID == mysqlResult.ID {
		t.Error("SQLite and MySQL generated the same ID -- each call should be unique")
	}
	if sqliteResult.ID == psqlResult.ID {
		t.Error("SQLite and PostgreSQL generated the same ID -- each call should be unique")
	}
}

// --- Cross-database MapCreateTableParams - Label passthrough ---

func TestCrossDatabaseMapCreateTableParams_LabelPassthrough(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		label string
	}{
		{"empty label", ""},
		{"simple label", "articles"},
		{"label with spaces", "blog posts"},
		{"unicode label", "inhalte"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := CreateTableParams{Label: tt.label}

			sqliteGot := Database{}.MapCreateTableParams(input)
			mysqlGot := MysqlDatabase{}.MapCreateTableParams(input)
			psqlGot := PsqlDatabase{}.MapCreateTableParams(input)

			if sqliteGot.Label != tt.label {
				t.Errorf("SQLite Label = %q, want %q", sqliteGot.Label, tt.label)
			}
			if mysqlGot.Label != tt.label {
				t.Errorf("MySQL Label = %q, want %q", mysqlGot.Label, tt.label)
			}
			if psqlGot.Label != tt.label {
				t.Errorf("PostgreSQL Label = %q, want %q", psqlGot.Label, tt.label)
			}
		})
	}
}

// --- Cross-database MapUpdateTableParams - field passthrough ---

func TestCrossDatabaseMapUpdateTableParams_FieldPassthrough(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		label string
		id    string
	}{
		{"normal values", "updated_articles", "tbl-001"},
		{"empty label", "", "tbl-002"},
		{"empty id", "some_label", ""},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := UpdateTableParams{Label: tt.label, ID: tt.id}

			sqliteGot := Database{}.MapUpdateTableParams(input)
			mysqlGot := MysqlDatabase{}.MapUpdateTableParams(input)
			psqlGot := PsqlDatabase{}.MapUpdateTableParams(input)

			if sqliteGot.Label != tt.label {
				t.Errorf("SQLite Label = %q, want %q", sqliteGot.Label, tt.label)
			}
			if sqliteGot.ID != tt.id {
				t.Errorf("SQLite ID = %q, want %q", sqliteGot.ID, tt.id)
			}
			if mysqlGot.Label != tt.label {
				t.Errorf("MySQL Label = %q, want %q", mysqlGot.Label, tt.label)
			}
			if mysqlGot.ID != tt.id {
				t.Errorf("MySQL ID = %q, want %q", mysqlGot.ID, tt.id)
			}
			if psqlGot.Label != tt.label {
				t.Errorf("PostgreSQL Label = %q, want %q", psqlGot.Label, tt.label)
			}
			if psqlGot.ID != tt.id {
				t.Errorf("PostgreSQL ID = %q, want %q", psqlGot.ID, tt.id)
			}
		})
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewTableCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("node-1"),
		RequestID: "req-table-001",
		IP:        "10.0.0.1",
	}
	params := CreateTableParams{
		Label: "blog_posts",
	}

	cmd := Database{}.NewTableCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "tables" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tables")
	}
	p, ok := cmd.Params().(CreateTableParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateTableParams", cmd.Params())
	}
	if p.Label != "blog_posts" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "blog_posts")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewTableCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	cmd := NewTableCmd{}

	row := mdb.Tables{ID: "tbl-row-001"}
	got := cmd.GetID(row)
	if got != "tbl-row-001" {
		t.Errorf("GetID() = %q, want %q", got, "tbl-row-001")
	}
}

func TestNewTableCmd_GetID_EmptyRow(t *testing.T) {
	t.Parallel()
	cmd := NewTableCmd{}

	row := mdb.Tables{}
	got := cmd.GetID(row)
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdateTableCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateTableParams{
		Label: "updated_posts",
		ID:    "tbl-upd-001",
	}

	cmd := Database{}.UpdateTableCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "tables" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tables")
	}
	if cmd.GetID() != "tbl-upd-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "tbl-upd-001")
	}
	p, ok := cmd.Params().(UpdateTableParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateTableParams", cmd.Params())
	}
	if p.Label != "updated_posts" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "updated_posts")
	}
	if p.ID != "tbl-upd-001" {
		t.Errorf("Params().ID = %q, want %q", p.ID, "tbl-upd-001")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteTableCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := "tbl-del-001"

	cmd := Database{}.DeleteTableCmd(ctx, ac, id)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "tables" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tables")
	}
	if cmd.GetID() != "tbl-del-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "tbl-del-001")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewTableCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-node"),
		RequestID: "mysql-table-001",
		IP:        "192.168.1.1",
	}
	params := CreateTableParams{
		Label: "mysql_posts",
	}

	cmd := MysqlDatabase{}.NewTableCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "tables" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tables")
	}
	p, ok := cmd.Params().(CreateTableParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateTableParams", cmd.Params())
	}
	if p.Label != "mysql_posts" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "mysql_posts")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewTableCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	cmd := NewTableCmdMysql{}

	row := mdbm.Tables{ID: "mysql-tbl-row-001"}
	got := cmd.GetID(row)
	if got != "mysql-tbl-row-001" {
		t.Errorf("GetID() = %q, want %q", got, "mysql-tbl-row-001")
	}
}

func TestUpdateTableCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateTableParams{
		Label: "mysql_updated",
		ID:    "tbl-mysql-upd-001",
	}

	cmd := MysqlDatabase{}.UpdateTableCmd(ctx, ac, params)

	if cmd.TableName() != "tables" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tables")
	}
	if cmd.GetID() != "tbl-mysql-upd-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "tbl-mysql-upd-001")
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateTableParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateTableParams", cmd.Params())
	}
	if p.Label != "mysql_updated" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "mysql_updated")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteTableCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := "tbl-mysql-del-001"

	cmd := MysqlDatabase{}.DeleteTableCmd(ctx, ac, id)

	if cmd.TableName() != "tables" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tables")
	}
	if cmd.GetID() != "tbl-mysql-del-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "tbl-mysql-del-001")
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- PostgreSQL Audited Command Accessor tests ---

func TestNewTableCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-node"),
		RequestID: "psql-table-001",
		IP:        "172.16.0.1",
	}
	params := CreateTableParams{
		Label: "psql_posts",
	}

	cmd := PsqlDatabase{}.NewTableCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "tables" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tables")
	}
	p, ok := cmd.Params().(CreateTableParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateTableParams", cmd.Params())
	}
	if p.Label != "psql_posts" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "psql_posts")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewTableCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	cmd := NewTableCmdPsql{}

	row := mdbp.Tables{ID: "psql-tbl-row-001"}
	got := cmd.GetID(row)
	if got != "psql-tbl-row-001" {
		t.Errorf("GetID() = %q, want %q", got, "psql-tbl-row-001")
	}
}

func TestUpdateTableCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := UpdateTableParams{
		Label: "psql_updated",
		ID:    "tbl-psql-upd-001",
	}

	cmd := PsqlDatabase{}.UpdateTableCmd(ctx, ac, params)

	if cmd.TableName() != "tables" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tables")
	}
	if cmd.GetID() != "tbl-psql-upd-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "tbl-psql-upd-001")
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateTableParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateTableParams", cmd.Params())
	}
	if p.Label != "psql_updated" {
		t.Errorf("Params().Label = %q, want %q", p.Label, "psql_updated")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteTableCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	id := "tbl-psql-del-001"

	cmd := PsqlDatabase{}.DeleteTableCmd(ctx, ac, id)

	if cmd.TableName() != "tables" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "tables")
	}
	if cmd.GetID() != "tbl-psql-del-001" {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), "tbl-psql-del-001")
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- Cross-database audited command consistency ---

func TestAuditedTableCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateTableParams{}
	updateParams := UpdateTableParams{
		ID: "tbl-consistency-001",
	}
	id := "tbl-consistency-del"

	commands := []struct {
		label string
		name  string
	}{
		{"SQLite Create", Database{}.NewTableCmd(ctx, ac, createParams).TableName()},
		{"SQLite Update", Database{}.UpdateTableCmd(ctx, ac, updateParams).TableName()},
		{"SQLite Delete", Database{}.DeleteTableCmd(ctx, ac, id).TableName()},
		{"MySQL Create", MysqlDatabase{}.NewTableCmd(ctx, ac, createParams).TableName()},
		{"MySQL Update", MysqlDatabase{}.UpdateTableCmd(ctx, ac, updateParams).TableName()},
		{"MySQL Delete", MysqlDatabase{}.DeleteTableCmd(ctx, ac, id).TableName()},
		{"PostgreSQL Create", PsqlDatabase{}.NewTableCmd(ctx, ac, createParams).TableName()},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateTableCmd(ctx, ac, updateParams).TableName()},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteTableCmd(ctx, ac, id).TableName()},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != "tables" {
				t.Errorf("TableName() = %q, want %q", c.name, "tables")
			}
		})
	}
}

func TestAuditedTableCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateTableParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewTableCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewTableCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewTableCmd(ctx, ac, createParams).Recorder()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.recorder == nil {
				t.Fatalf("%s recorder is nil", tt.name)
			}
		})
	}
}

// --- Audited Command GetID consistency ---

func TestAuditedTableCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	id := "tbl-getid-cross-001"

	t.Run("UpdateCmd GetID returns table ID", func(t *testing.T) {
		t.Parallel()
		params := UpdateTableParams{
			Label: "some_label",
			ID:    id,
		}

		sqliteCmd := Database{}.UpdateTableCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateTableCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateTableCmd(ctx, ac, params)

		if sqliteCmd.GetID() != id {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(), id)
		}
		if mysqlCmd.GetID() != id {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(), id)
		}
		if psqlCmd.GetID() != id {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(), id)
		}
	})

	t.Run("DeleteCmd GetID returns table ID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteTableCmd(ctx, ac, id)
		mysqlCmd := MysqlDatabase{}.DeleteTableCmd(ctx, ac, id)
		psqlCmd := PsqlDatabase{}.DeleteTableCmd(ctx, ac, id)

		if sqliteCmd.GetID() != id {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(), id)
		}
		if mysqlCmd.GetID() != id {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(), id)
		}
		if psqlCmd.GetID() != id {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(), id)
		}
	})

	t.Run("CreateCmd GetID extracts from result row", func(t *testing.T) {
		t.Parallel()
		wantID := "tbl-create-row-001"

		sqliteCmd := NewTableCmd{}
		mysqlCmd := NewTableCmdMysql{}
		psqlCmd := NewTableCmdPsql{}

		sqliteRow := mdb.Tables{ID: wantID}
		mysqlRow := mdbm.Tables{ID: wantID}
		psqlRow := mdbp.Tables{ID: wantID}

		if sqliteCmd.GetID(sqliteRow) != wantID {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(sqliteRow), wantID)
		}
		if mysqlCmd.GetID(mysqlRow) != wantID {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(mysqlRow), wantID)
		}
		if psqlCmd.GetID(psqlRow) != wantID {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(psqlRow), wantID)
		}
	})
}

// --- Edge cases: empty IDs ---

func TestUpdateTableCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateTableParams{
		ID: "",
	}

	sqliteCmd := Database{}.UpdateTableCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateTableCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateTableCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestDeleteTableCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := ""

	sqliteCmd := Database{}.DeleteTableCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteTableCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteTableCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// --- Verify audited commands satisfy their interfaces ---
// These are compile-time checks. If these fail to compile, the command structs
// don't implement the required audited interfaces.

var (
	_ audited.CreateCommand[mdb.Tables]  = NewTableCmd{}
	_ audited.UpdateCommand[mdb.Tables]  = UpdateTableCmd{}
	_ audited.DeleteCommand[mdb.Tables]  = DeleteTableCmd{}
	_ audited.CreateCommand[mdbm.Tables] = NewTableCmdMysql{}
	_ audited.UpdateCommand[mdbm.Tables] = UpdateTableCmdMysql{}
	_ audited.DeleteCommand[mdbm.Tables] = DeleteTableCmdMysql{}
	_ audited.CreateCommand[mdbp.Tables] = NewTableCmdPsql{}
	_ audited.UpdateCommand[mdbp.Tables] = UpdateTableCmdPsql{}
	_ audited.DeleteCommand[mdbp.Tables] = DeleteTableCmdPsql{}
)

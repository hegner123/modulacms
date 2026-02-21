package main

// Field describes a single field in an entity struct.
type Field struct {
	AppName     string // field name in wrapper struct (e.g., "Role")
	SqlcName    string // field name in sqlc struct if different (e.g., "Roles"), empty = same as AppName
	Type        string // Go type (e.g., "types.UserID", "sql.NullString")
	JSONTag     string // json tag value
	IsPrimaryID bool   // this is the entity's primary key
	InCreate    bool   // include in CreateParams
	InUpdate    bool   // include in UpdateParams
}

// SqlcFieldName returns the sqlc struct field name for this field.
// If SqlcName is set, returns it; otherwise returns AppName.
func (f Field) SqlcFieldName() string {
	if f.SqlcName != "" {
		return f.SqlcName
	}
	return f.AppName
}

// Entity defines a database entity for code generation.
type Entity struct {
	Name     string // struct name: "Users", "Media"
	Singular string // method name component: "User", "Media"
	Plural   string // list method name: "Users", "Media"

	SqlcTypeName string // sqlc struct name (usually == Name, but "DatatypesFields" for DatatypeFields)
	TableName    string // SQL table: "users", "media"

	IDType    string // "types.UserID" or "string"
	IDField   string // "UserID" or "ID"
	NewIDFunc string // "types.NewUserID()" or "string(types.NewTokenID())"

	HasPaginated     bool // generate ListPaginated
	CallerSuppliedID bool // ID in CreateParams, generate-if-empty pattern

	UpdateSuccessField string // e.g., "s.Username" for update message

	// Skip flags for entities with per-driver differences
	SkipMappers         bool // skip MapEntity, MapCreate, MapUpdate (stay hand-written due to per-driver conversion)
	SkipAuditedCommands bool // skip audited command structs + factory methods (stay hand-written)
	SkipGet             bool // skip Get CRUD method (no matching sqlc Get query, or ID field name differs)

	// Query name overrides (empty = use default pattern)
	SqlcCreateTableName string // e.g., "CreateDatatypesFieldsTable"
	SqlcCountName       string // e.g., "CountAdminroute" when sqlc lowercases
	SqlcGetName         string // e.g., "GetDatatypeField" override
	SqlcListName        string // e.g., "ListDatatypeField" override

	Fields     []Field
	OutputFile string // "user_gen.go"
}

// CreateTableQueryName returns the sqlc query name for CreateTable.
func (e Entity) CreateTableQueryName() string {
	if e.SqlcCreateTableName != "" {
		return e.SqlcCreateTableName
	}
	return "Create" + e.Singular + "Table"
}

// CountQueryName returns the sqlc query name for Count.
func (e Entity) CountQueryName() string {
	if e.SqlcCountName != "" {
		return e.SqlcCountName
	}
	return "Count" + e.Singular
}

// GetQueryName returns the sqlc query name for Get.
func (e Entity) GetQueryName() string {
	if e.SqlcGetName != "" {
		return e.SqlcGetName
	}
	return "Get" + e.Singular
}

// ListQueryName returns the sqlc query name for List.
func (e Entity) ListQueryName() string {
	if e.SqlcListName != "" {
		return e.SqlcListName
	}
	return "List" + e.Singular
}

// CreateFields returns fields where InCreate is true.
func (e Entity) CreateFields() []Field {
	var fields []Field
	for _, f := range e.Fields {
		if f.InCreate {
			fields = append(fields, f)
		}
	}
	return fields
}

// UpdateFields returns fields where InUpdate is true.
func (e Entity) UpdateFields() []Field {
	var fields []Field
	for _, f := range e.Fields {
		if f.InUpdate {
			fields = append(fields, f)
		}
	}
	return fields
}

// NonIDCreateFields returns create fields excluding the primary ID.
func (e Entity) NonIDCreateFields() []Field {
	var fields []Field
	for _, f := range e.Fields {
		if f.InCreate && !f.IsPrimaryID {
			fields = append(fields, f)
		}
	}
	return fields
}

// NonIDUpdateFields returns update fields excluding the primary ID.
func (e Entity) NonIDUpdateFields() []Field {
	var fields []Field
	for _, f := range e.Fields {
		if f.InUpdate && !f.IsPrimaryID {
			fields = append(fields, f)
		}
	}
	return fields
}

// IDIsTyped returns whether the entity's ID type is a typed ID (not plain string).
func (e Entity) IDIsTyped() bool {
	return e.IDType != "string"
}

// IDToString returns the expression to convert an ID value to string.
// For typed IDs: string(expr), for string IDs: just expr.
func (e Entity) IDToString(expr string) string {
	if e.IDIsTyped() {
		return "string(" + expr + ")"
	}
	return expr
}

// DriverConfig describes a database driver for code generation.
type DriverConfig struct {
	Name              string // "sqlite", "mysql", "psql"
	Struct            string // "Database", "MysqlDatabase", "PsqlDatabase"
	Package           string // "mdb", "mdbm", "mdbp"
	Recorder          string // "SQLiteRecorder", "MysqlRecorder", "PsqlRecorder"
	CmdSuffix         string // "", "Mysql", "Psql"
	MysqlReturningGap bool   // true for MySQL (exec then get pattern)
	Int32Pagination   bool   // true for MySQL and PostgreSQL
}

// DriverConfigs defines all three database driver configurations.
var DriverConfigs = []DriverConfig{
	{
		Name:              "sqlite",
		Struct:            "Database",
		Package:           "mdb",
		Recorder:          "SQLiteRecorder",
		CmdSuffix:         "",
		MysqlReturningGap: false,
		Int32Pagination:   false,
	},
	{
		Name:              "mysql",
		Struct:            "MysqlDatabase",
		Package:           "mdbm",
		Recorder:          "MysqlRecorder",
		CmdSuffix:         "Mysql",
		MysqlReturningGap: true,
		Int32Pagination:   true,
	},
	{
		Name:              "psql",
		Struct:            "PsqlDatabase",
		Package:           "mdbp",
		Recorder:          "PsqlRecorder",
		CmdSuffix:         "Psql",
		MysqlReturningGap: false,
		Int32Pagination:   true,
	},
}

// TemplateData is the top-level data passed to the template.
type TemplateData struct {
	Entity  Entity
	Drivers []DriverConfig
}

// DriverEntityData combines a driver config with entity data for per-driver template sections.
type DriverEntityData struct {
	Entity Entity
	Driver DriverConfig
}

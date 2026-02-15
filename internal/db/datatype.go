package db

import (
	"context"
	"database/sql"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

///////////////////////////////
// STRUCTS
//////////////////////////////

// Datatypes represents a datatype record in the database.
type Datatypes struct {
	DatatypeID   types.DatatypeID        `json:"datatype_id"`
	ParentID     types.NullableDatatypeID `json:"parent_id"`
	Label        string                  `json:"label"`
	Type         string                  `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
}

// CreateDatatypeParams holds the parameters for creating a new datatype.
type CreateDatatypeParams struct {
	DatatypeID   types.DatatypeID        `json:"datatype_id"`
	ParentID     types.NullableDatatypeID `json:"parent_id"`
	Label        string                  `json:"label"`
	Type         string                  `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
}

// UpdateDatatypeParams holds the parameters for updating an existing datatype.
type UpdateDatatypeParams struct {
	ParentID     types.NullableDatatypeID `json:"parent_id"`
	Label        string                  `json:"label"`
	Type         string                  `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
	DatatypeID   types.DatatypeID        `json:"datatype_id"`
}

// ListDatatypeChildrenPaginatedParams holds parameters for paginated listing of datatype children.
type ListDatatypeChildrenPaginatedParams struct {
	ParentID types.DatatypeID
	Limit    int64
	Offset   int64
}

// DatatypeJSON provides a string-based representation for JSON serialization
// DatatypeJSON provides a string-based representation for JSON serialization.
type DatatypeJSON struct {
	DatatypeID   string `json:"datatype_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Type         string `json:"type"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

// MapDatatypeJSON converts Datatypes to DatatypeJSON for JSON serialization.
func MapDatatypeJSON(a Datatypes) DatatypeJSON {
	return DatatypeJSON{
		DatatypeID:   a.DatatypeID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
	}
}

// MapStringDatatype converts Datatypes to StringDatatypes for table display.
func MapStringDatatype(a Datatypes) StringDatatypes {
	return StringDatatypes{
		DatatypeID:   a.DatatypeID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
		History:      "", // History field removed
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

// MapDatatype converts a sqlc-generated type to the wrapper type.
func (d Database) MapDatatype(a mdb.Datatypes) Datatypes {
	return Datatypes{
		DatatypeID:   a.DatatypeID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapCreateDatatypeParams converts wrapper params to sqlc-generated params.
func (d Database) MapCreateDatatypeParams(a CreateDatatypeParams) mdb.CreateDatatypeParams {
	id := a.DatatypeID
	if id.IsZero() {
		id = types.NewDatatypeID()
	}
	return mdb.CreateDatatypeParams{
		DatatypeID:   id,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateDatatypeParams converts wrapper update params to sqlc-generated params.
func (d Database) MapUpdateDatatypeParams(a UpdateDatatypeParams) mdb.UpdateDatatypeParams {
	return mdb.UpdateDatatypeParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		DatatypeID:   a.DatatypeID,
	}
}

// QUERIES

// CountDatatypes returns the total number of datatypes.
func (d Database) CountDatatypes() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateDatatypeTable creates the datatypes table.
func (d Database) CreateDatatypeTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateDatatypeTable(d.Context)
	return err
}

// CreateDatatype inserts a new datatype record.
func (d Database) CreateDatatype(ctx context.Context, ac audited.AuditContext, s CreateDatatypeParams) (*Datatypes, error) {
	cmd := d.NewDatatypeCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create datatype: %w", err)
	}
	r := d.MapDatatype(result)
	return &r, nil
}

// DeleteDatatype removes a datatype record by ID.
func (d Database) DeleteDatatype(ctx context.Context, ac audited.AuditContext, id types.DatatypeID) error {
	cmd := d.DeleteDatatypeCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetDatatype retrieves a datatype by ID.
func (d Database) GetDatatype(id types.DatatypeID) (*Datatypes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetDatatype(d.Context, mdb.GetDatatypeParams{DatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapDatatype(row)
	return &res, nil
}

// ListDatatypes returns all datatypes.
func (d Database) ListDatatypes() (*[]Datatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypesRoot returns all root-level datatypes.
func (d Database) ListDatatypesRoot() (*[]Datatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatypeRoot(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypeChildren returns all child datatypes of a parent.
func (d Database) ListDatatypeChildren(parentID types.DatatypeID) (*[]Datatypes, error) {
	queries := mdb.New(d.Connection)
	// Convert DatatypeID to NullableDatatypeID for the query parameter
	params := mdb.ListDatatypeChildrenParams{
		ParentID: types.NullableDatatypeID{ID: parentID, Valid: true},
	}
	rows, err := queries.ListDatatypeChildren(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get child datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypesPaginated returns paginated datatypes.
func (d Database) ListDatatypesPaginated(params PaginationParams) (*[]Datatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatypePaginated(d.Context, mdb.ListDatatypePaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes paginated: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypeChildrenPaginated returns paginated child datatypes of a parent.
func (d Database) ListDatatypeChildrenPaginated(params ListDatatypeChildrenPaginatedParams) (*[]Datatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatypeChildrenPaginated(d.Context, mdb.ListDatatypeChildrenPaginatedParams{
		ParentID: types.NullableDatatypeID{ID: params.ParentID, Valid: true},
		Limit:    params.Limit,
		Offset:   params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get child datatypes paginated: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateDatatype modifies an existing datatype record.
func (d Database) UpdateDatatype(ctx context.Context, ac audited.AuditContext, s UpdateDatatypeParams) (*string, error) {
	cmd := d.UpdateDatatypeCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update datatype: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

// MapDatatype converts a sqlc-generated type to the wrapper type.
func (d MysqlDatabase) MapDatatype(a mdbm.Datatypes) Datatypes {
	return Datatypes{
		DatatypeID:   a.DatatypeID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapCreateDatatypeParams converts wrapper params to sqlc-generated params.
func (d MysqlDatabase) MapCreateDatatypeParams(a CreateDatatypeParams) mdbm.CreateDatatypeParams {
	id := a.DatatypeID
	if id.IsZero() {
		id = types.NewDatatypeID()
	}
	return mdbm.CreateDatatypeParams{
		DatatypeID: id,
		ParentID:   a.ParentID,
		Label:    a.Label,
		Type:     a.Type,
		AuthorID: a.AuthorID,
	}
}

// MapUpdateDatatypeParams converts wrapper update params to sqlc-generated params.
func (d MysqlDatabase) MapUpdateDatatypeParams(a UpdateDatatypeParams) mdbm.UpdateDatatypeParams {
	return mdbm.UpdateDatatypeParams{
		ParentID:   a.ParentID,
		Label:      a.Label,
		Type:       a.Type,
		AuthorID:   a.AuthorID,
		DatatypeID: a.DatatypeID,
	}
}

// QUERIES

// CountDatatypes returns the total number of datatypes.
func (d MysqlDatabase) CountDatatypes() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateDatatypeTable creates the datatypes table.
func (d MysqlDatabase) CreateDatatypeTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateDatatypeTable(d.Context)
	return err
}

// CreateDatatype inserts a new datatype record.
func (d MysqlDatabase) CreateDatatype(ctx context.Context, ac audited.AuditContext, s CreateDatatypeParams) (*Datatypes, error) {
	cmd := d.NewDatatypeCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create datatype: %w", err)
	}
	r := d.MapDatatype(result)
	return &r, nil
}

// DeleteDatatype removes a datatype record by ID.
func (d MysqlDatabase) DeleteDatatype(ctx context.Context, ac audited.AuditContext, id types.DatatypeID) error {
	cmd := d.DeleteDatatypeCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetDatatype retrieves a datatype by ID.
func (d MysqlDatabase) GetDatatype(id types.DatatypeID) (*Datatypes, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetDatatype(d.Context, mdbm.GetDatatypeParams{DatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapDatatype(row)
	return &res, nil
}

// ListDatatypes returns all datatypes.
func (d MysqlDatabase) ListDatatypes() (*[]Datatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypesRoot returns all root-level datatypes.
func (d MysqlDatabase) ListDatatypesRoot() (*[]Datatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatypeRoot(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypeChildren returns all child datatypes of a parent.
func (d MysqlDatabase) ListDatatypeChildren(parentID types.DatatypeID) (*[]Datatypes, error) {
	queries := mdbm.New(d.Connection)
	params := mdbm.ListDatatypeChildrenParams{
		ParentID: types.NullableDatatypeID{ID: parentID, Valid: true},
	}
	rows, err := queries.ListDatatypeChildren(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get child datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypesPaginated returns paginated datatypes.
func (d MysqlDatabase) ListDatatypesPaginated(params PaginationParams) (*[]Datatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatypePaginated(d.Context, mdbm.ListDatatypePaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes paginated: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypeChildrenPaginated returns paginated child datatypes of a parent.
func (d MysqlDatabase) ListDatatypeChildrenPaginated(params ListDatatypeChildrenPaginatedParams) (*[]Datatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatypeChildrenPaginated(d.Context, mdbm.ListDatatypeChildrenPaginatedParams{
		ParentID: types.NullableDatatypeID{ID: params.ParentID, Valid: true},
		Limit:    int32(params.Limit),
		Offset:   int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get child datatypes paginated: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateDatatype modifies an existing datatype record.
func (d MysqlDatabase) UpdateDatatype(ctx context.Context, ac audited.AuditContext, s UpdateDatatypeParams) (*string, error) {
	cmd := d.UpdateDatatypeCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update datatype: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

// MapDatatype converts a sqlc-generated type to the wrapper type.
func (d PsqlDatabase) MapDatatype(a mdbp.Datatypes) Datatypes {
	return Datatypes{
		DatatypeID:   a.DatatypeID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapCreateDatatypeParams converts wrapper params to sqlc-generated params.
func (d PsqlDatabase) MapCreateDatatypeParams(a CreateDatatypeParams) mdbp.CreateDatatypeParams {
	id := a.DatatypeID
	if id.IsZero() {
		id = types.NewDatatypeID()
	}
	return mdbp.CreateDatatypeParams{
		DatatypeID:   id,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateDatatypeParams converts wrapper update params to sqlc-generated params.
func (d PsqlDatabase) MapUpdateDatatypeParams(a UpdateDatatypeParams) mdbp.UpdateDatatypeParams {
	return mdbp.UpdateDatatypeParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		DatatypeID:   a.DatatypeID,
	}
}

// QUERIES

// CountDatatypes returns the total number of datatypes.
func (d PsqlDatabase) CountDatatypes() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateDatatypeTable creates the datatypes table.
func (d PsqlDatabase) CreateDatatypeTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateDatatypeTable(d.Context)
	return err
}

// CreateDatatype inserts a new datatype record.
func (d PsqlDatabase) CreateDatatype(ctx context.Context, ac audited.AuditContext, s CreateDatatypeParams) (*Datatypes, error) {
	cmd := d.NewDatatypeCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create datatype: %w", err)
	}
	r := d.MapDatatype(result)
	return &r, nil
}

// DeleteDatatype removes a datatype record by ID.
func (d PsqlDatabase) DeleteDatatype(ctx context.Context, ac audited.AuditContext, id types.DatatypeID) error {
	cmd := d.DeleteDatatypeCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetDatatype retrieves a datatype by ID.
func (d PsqlDatabase) GetDatatype(id types.DatatypeID) (*Datatypes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetDatatype(d.Context, mdbp.GetDatatypeParams{DatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapDatatype(row)
	return &res, nil
}

// ListDatatypes returns all datatypes.
func (d PsqlDatabase) ListDatatypes() (*[]Datatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypesRoot returns all root-level datatypes.
func (d PsqlDatabase) ListDatatypesRoot() (*[]Datatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatypeRoot(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypeChildren returns all child datatypes of a parent.
func (d PsqlDatabase) ListDatatypeChildren(parentID types.DatatypeID) (*[]Datatypes, error) {
	queries := mdbp.New(d.Connection)
	params := mdbp.ListDatatypeChildrenParams{
		ParentID: types.NullableDatatypeID{ID: parentID, Valid: true},
	}
	rows, err := queries.ListDatatypeChildren(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get child datatypes: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypesPaginated returns paginated datatypes.
func (d PsqlDatabase) ListDatatypesPaginated(params PaginationParams) (*[]Datatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatypePaginated(d.Context, mdbp.ListDatatypePaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes paginated: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListDatatypeChildrenPaginated returns paginated child datatypes of a parent.
func (d PsqlDatabase) ListDatatypeChildrenPaginated(params ListDatatypeChildrenPaginatedParams) (*[]Datatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListDatatypeChildrenPaginated(d.Context, mdbp.ListDatatypeChildrenPaginatedParams{
		ParentID: types.NullableDatatypeID{ID: params.ParentID, Valid: true},
		Limit:    int32(params.Limit),
		Offset:   int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get child datatypes paginated: %v", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateDatatype modifies an existing datatype record.
func (d PsqlDatabase) UpdateDatatype(ctx context.Context, ac audited.AuditContext, s UpdateDatatypeParams) (*string, error) {
	cmd := d.UpdateDatatypeCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update datatype: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ----- SQLite CREATE -----

// NewDatatypeCmd is an audited command for creating a new datatype.
type NewDatatypeCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c NewDatatypeCmd) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c NewDatatypeCmd) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewDatatypeCmd) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewDatatypeCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewDatatypeCmd) TableName() string                     { return "datatypes" }

// Params returns the command parameters.
func (c NewDatatypeCmd) Params() any                           { return c.params }

// GetID returns the ID from the result.
func (c NewDatatypeCmd) GetID(d mdb.Datatypes) string          { return string(d.DatatypeID) }

// Execute executes the create command.
func (c NewDatatypeCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Datatypes, error) {
	id := c.params.DatatypeID
	if id.IsZero() {
		id = types.NewDatatypeID()
	}
	queries := mdb.New(tx)
	return queries.CreateDatatype(ctx, mdb.CreateDatatypeParams{
		DatatypeID:   id,
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

// NewDatatypeCmd creates a new create command.
func (d Database) NewDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateDatatypeParams) NewDatatypeCmd {
	return NewDatatypeCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE -----

// UpdateDatatypeCmd is an audited command for updating a datatype.
type UpdateDatatypeCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c UpdateDatatypeCmd) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateDatatypeCmd) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateDatatypeCmd) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateDatatypeCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c UpdateDatatypeCmd) TableName() string                     { return "datatypes" }

// Params returns the command parameters.
func (c UpdateDatatypeCmd) Params() any                           { return c.params }

// GetID returns the ID to update.
func (c UpdateDatatypeCmd) GetID() string                         { return string(c.params.DatatypeID) }

// GetBefore retrieves the record before updating.
func (c UpdateDatatypeCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Datatypes, error) {
	queries := mdb.New(tx)
	return queries.GetDatatype(ctx, mdb.GetDatatypeParams{DatatypeID: c.params.DatatypeID})
}

// Execute executes the update command.
func (c UpdateDatatypeCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateDatatype(ctx, mdb.UpdateDatatypeParams{
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		DatatypeID:   c.params.DatatypeID,
	})
}

// UpdateDatatypeCmd creates a new update command.
func (d Database) UpdateDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateDatatypeParams) UpdateDatatypeCmd {
	return UpdateDatatypeCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

// DeleteDatatypeCmd is an audited command for deleting a datatype.
type DeleteDatatypeCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.DatatypeID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c DeleteDatatypeCmd) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteDatatypeCmd) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteDatatypeCmd) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteDatatypeCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteDatatypeCmd) TableName() string                     { return "datatypes" }

// GetID returns the ID to delete.
func (c DeleteDatatypeCmd) GetID() string                         { return string(c.id) }

// GetBefore retrieves the record before deleting.
func (c DeleteDatatypeCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Datatypes, error) {
	queries := mdb.New(tx)
	return queries.GetDatatype(ctx, mdb.GetDatatypeParams{DatatypeID: c.id})
}

// Execute executes the delete command.
func (c DeleteDatatypeCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteDatatype(ctx, mdb.DeleteDatatypeParams{DatatypeID: c.id})
}

// DeleteDatatypeCmd creates a new delete command.
func (d Database) DeleteDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, id types.DatatypeID) DeleteDatatypeCmd {
	return DeleteDatatypeCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

// NewDatatypeCmdMysql is an audited command for creating a new datatype in MySQL.
type NewDatatypeCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c NewDatatypeCmdMysql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c NewDatatypeCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewDatatypeCmdMysql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewDatatypeCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewDatatypeCmdMysql) TableName() string                     { return "datatypes" }

// Params returns the command parameters.
func (c NewDatatypeCmdMysql) Params() any                           { return c.params }

// GetID returns the ID from the result.
func (c NewDatatypeCmdMysql) GetID(d mdbm.Datatypes) string         { return string(d.DatatypeID) }

// Execute executes the create command.
func (c NewDatatypeCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Datatypes, error) {
	id := c.params.DatatypeID
	if id.IsZero() {
		id = types.NewDatatypeID()
	}
	queries := mdbm.New(tx)
	params := mdbm.CreateDatatypeParams{
		DatatypeID: id,
		ParentID:   c.params.ParentID,
		Label:      c.params.Label,
		Type:       c.params.Type,
		AuthorID:   c.params.AuthorID,
	}
	if err := queries.CreateDatatype(ctx, params); err != nil {
		return mdbm.Datatypes{}, err
	}
	return queries.GetDatatype(ctx, mdbm.GetDatatypeParams{DatatypeID: params.DatatypeID})
}

// NewDatatypeCmd creates a new create command.
func (d MysqlDatabase) NewDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateDatatypeParams) NewDatatypeCmdMysql {
	return NewDatatypeCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE -----

// UpdateDatatypeCmdMysql is an audited command for updating a datatype in MySQL.
type UpdateDatatypeCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c UpdateDatatypeCmdMysql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateDatatypeCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateDatatypeCmdMysql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateDatatypeCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c UpdateDatatypeCmdMysql) TableName() string                     { return "datatypes" }

// Params returns the command parameters.
func (c UpdateDatatypeCmdMysql) Params() any                           { return c.params }

// GetID returns the ID to update.
func (c UpdateDatatypeCmdMysql) GetID() string                         { return string(c.params.DatatypeID) }

// GetBefore retrieves the record before updating.
func (c UpdateDatatypeCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Datatypes, error) {
	queries := mdbm.New(tx)
	return queries.GetDatatype(ctx, mdbm.GetDatatypeParams{DatatypeID: c.params.DatatypeID})
}

// Execute executes the update command.
func (c UpdateDatatypeCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateDatatype(ctx, mdbm.UpdateDatatypeParams{
		ParentID:   c.params.ParentID,
		Label:      c.params.Label,
		Type:       c.params.Type,
		AuthorID:   c.params.AuthorID,
		DatatypeID: c.params.DatatypeID,
	})
}

// UpdateDatatypeCmd creates a new update command.
func (d MysqlDatabase) UpdateDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateDatatypeParams) UpdateDatatypeCmdMysql {
	return UpdateDatatypeCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

// DeleteDatatypeCmdMysql is an audited command for deleting a datatype in MySQL.
type DeleteDatatypeCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.DatatypeID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c DeleteDatatypeCmdMysql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteDatatypeCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteDatatypeCmdMysql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteDatatypeCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteDatatypeCmdMysql) TableName() string                     { return "datatypes" }

// GetID returns the ID to delete.
func (c DeleteDatatypeCmdMysql) GetID() string                         { return string(c.id) }

// GetBefore retrieves the record before deleting.
func (c DeleteDatatypeCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Datatypes, error) {
	queries := mdbm.New(tx)
	return queries.GetDatatype(ctx, mdbm.GetDatatypeParams{DatatypeID: c.id})
}

// Execute executes the delete command.
func (c DeleteDatatypeCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteDatatype(ctx, mdbm.DeleteDatatypeParams{DatatypeID: c.id})
}

// DeleteDatatypeCmd creates a new delete command.
func (d MysqlDatabase) DeleteDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, id types.DatatypeID) DeleteDatatypeCmdMysql {
	return DeleteDatatypeCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

// NewDatatypeCmdPsql is an audited command for creating a new datatype in PostgreSQL.
type NewDatatypeCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c NewDatatypeCmdPsql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c NewDatatypeCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewDatatypeCmdPsql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewDatatypeCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewDatatypeCmdPsql) TableName() string                     { return "datatypes" }

// Params returns the command parameters.
func (c NewDatatypeCmdPsql) Params() any                           { return c.params }

// GetID returns the ID from the result.
func (c NewDatatypeCmdPsql) GetID(d mdbp.Datatypes) string         { return string(d.DatatypeID) }

// Execute executes the create command.
func (c NewDatatypeCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Datatypes, error) {
	id := c.params.DatatypeID
	if id.IsZero() {
		id = types.NewDatatypeID()
	}
	queries := mdbp.New(tx)
	return queries.CreateDatatype(ctx, mdbp.CreateDatatypeParams{
		DatatypeID:   id,
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

// NewDatatypeCmd creates a new create command.
func (d PsqlDatabase) NewDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateDatatypeParams) NewDatatypeCmdPsql {
	return NewDatatypeCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE -----

// UpdateDatatypeCmdPsql is an audited command for updating a datatype in PostgreSQL.
type UpdateDatatypeCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c UpdateDatatypeCmdPsql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateDatatypeCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateDatatypeCmdPsql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateDatatypeCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c UpdateDatatypeCmdPsql) TableName() string                     { return "datatypes" }

// Params returns the command parameters.
func (c UpdateDatatypeCmdPsql) Params() any                           { return c.params }

// GetID returns the ID to update.
func (c UpdateDatatypeCmdPsql) GetID() string                         { return string(c.params.DatatypeID) }

// GetBefore retrieves the record before updating.
func (c UpdateDatatypeCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Datatypes, error) {
	queries := mdbp.New(tx)
	return queries.GetDatatype(ctx, mdbp.GetDatatypeParams{DatatypeID: c.params.DatatypeID})
}

// Execute executes the update command.
func (c UpdateDatatypeCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateDatatype(ctx, mdbp.UpdateDatatypeParams{
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		DatatypeID:   c.params.DatatypeID,
	})
}

// UpdateDatatypeCmd creates a new update command.
func (d PsqlDatabase) UpdateDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateDatatypeParams) UpdateDatatypeCmdPsql {
	return UpdateDatatypeCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

// DeleteDatatypeCmdPsql is an audited command for deleting a datatype in PostgreSQL.
type DeleteDatatypeCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.DatatypeID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c DeleteDatatypeCmdPsql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteDatatypeCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteDatatypeCmdPsql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteDatatypeCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteDatatypeCmdPsql) TableName() string                     { return "datatypes" }

// GetID returns the ID to delete.
func (c DeleteDatatypeCmdPsql) GetID() string                         { return string(c.id) }

// GetBefore retrieves the record before deleting.
func (c DeleteDatatypeCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Datatypes, error) {
	queries := mdbp.New(tx)
	return queries.GetDatatype(ctx, mdbp.GetDatatypeParams{DatatypeID: c.id})
}

// Execute executes the delete command.
func (c DeleteDatatypeCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteDatatype(ctx, mdbp.DeleteDatatypeParams{DatatypeID: c.id})
}

// DeleteDatatypeCmd creates a new delete command.
func (d PsqlDatabase) DeleteDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, id types.DatatypeID) DeleteDatatypeCmdPsql {
	return DeleteDatatypeCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}

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

// Locale represents a locale configuration.
type Locale struct {
	LocaleID     types.LocaleID  `json:"locale_id"`
	Code         string          `json:"code"`
	Label        string          `json:"label"`
	IsDefault    bool            `json:"is_default"`
	IsEnabled    bool            `json:"is_enabled"`
	FallbackCode string          `json:"fallback_code"`
	SortOrder    int64           `json:"sort_order"`
	DateCreated  types.Timestamp `json:"date_created"`
}

// CreateLocaleParams specifies parameters for creating a locale.
type CreateLocaleParams struct {
	Code         string          `json:"code"`
	Label        string          `json:"label"`
	IsDefault    bool            `json:"is_default"`
	IsEnabled    bool            `json:"is_enabled"`
	FallbackCode string          `json:"fallback_code"`
	SortOrder    int64           `json:"sort_order"`
	DateCreated  types.Timestamp `json:"date_created"`
}

// UpdateLocaleParams specifies parameters for updating a locale.
type UpdateLocaleParams struct {
	Code         string          `json:"code"`
	Label        string          `json:"label"`
	IsDefault    bool            `json:"is_default"`
	IsEnabled    bool            `json:"is_enabled"`
	FallbackCode string          `json:"fallback_code"`
	SortOrder    int64           `json:"sort_order"`
	DateCreated  types.Timestamp `json:"date_created"`
	LocaleID     types.LocaleID  `json:"locale_id"`
}

// StringLocale is the string representation of Locale for TUI table display.
type StringLocale struct {
	LocaleID     string `json:"locale_id"`
	Code         string `json:"code"`
	Label        string `json:"label"`
	IsDefault    string `json:"is_default"`
	IsEnabled    string `json:"is_enabled"`
	FallbackCode string `json:"fallback_code"`
	SortOrder    string `json:"sort_order"`
	DateCreated  string `json:"date_created"`
}

// MapStringLocale converts Locale to StringLocale for table display.
func MapStringLocale(a Locale) StringLocale {
	return StringLocale{
		LocaleID:     a.LocaleID.String(),
		Code:         a.Code,
		Label:        a.Label,
		IsDefault:    fmt.Sprintf("%t", a.IsDefault),
		IsEnabled:    fmt.Sprintf("%t", a.IsEnabled),
		FallbackCode: a.FallbackCode,
		SortOrder:    fmt.Sprintf("%d", a.SortOrder),
		DateCreated:  a.DateCreated.String(),
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

// MapLocale converts a sqlc-generated SQLite type to the wrapper type.
func (d Database) MapLocale(a mdb.Locale) Locale {
	return Locale{
		LocaleID:     a.LocaleID,
		Code:         a.Code,
		Label:        a.Label,
		IsDefault:    a.IsDefault.Bool(),
		IsEnabled:    a.IsEnabled.Bool(),
		FallbackCode: ReadNullString(a.FallbackCode),
		SortOrder:    a.SortOrder,
		DateCreated:  a.DateCreated,
	}
}

// MapCreateLocaleParams converts wrapper params to a sqlc-generated SQLite type.
func (d Database) MapCreateLocaleParams(a CreateLocaleParams) mdb.CreateLocaleParams {
	return mdb.CreateLocaleParams{
		LocaleID:     types.NewLocaleID(),
		Code:         a.Code,
		Label:        a.Label,
		IsDefault:    types.NewSafeBool(a.IsDefault),
		IsEnabled:    types.NewSafeBool(a.IsEnabled),
		FallbackCode: StringToNullString(a.FallbackCode),
		SortOrder:    a.SortOrder,
		DateCreated:  a.DateCreated,
	}
}

// MapUpdateLocaleParams converts wrapper params to a sqlc-generated SQLite type.
func (d Database) MapUpdateLocaleParams(a UpdateLocaleParams) mdb.UpdateLocaleParams {
	return mdb.UpdateLocaleParams{
		Code:         a.Code,
		Label:        a.Label,
		IsDefault:    types.NewSafeBool(a.IsDefault),
		IsEnabled:    types.NewSafeBool(a.IsEnabled),
		FallbackCode: StringToNullString(a.FallbackCode),
		SortOrder:    a.SortOrder,
		DateCreated:  a.DateCreated,
		LocaleID:     a.LocaleID,
	}
}

// QUERIES

// CountLocales returns the total count of locales.
func (d Database) CountLocales() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountLocale(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count locales: %w", err)
	}
	return &c, nil
}

// CreateLocaleTable creates the locales table.
func (d Database) CreateLocaleTable() error {
	queries := mdb.New(d.Connection)
	return queries.CreateLocaleTable(d.Context)
}

// DropLocaleTable drops the locales table.
func (d Database) DropLocaleTable() error {
	queries := mdb.New(d.Connection)
	return queries.DropLocaleTable(d.Context)
}

// CreateLocale creates a new locale with audit trail.
func (d Database) CreateLocale(ctx context.Context, ac audited.AuditContext, s CreateLocaleParams) (*Locale, error) {
	cmd := d.NewLocaleCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create locale: %w", err)
	}
	r := d.MapLocale(result)
	return &r, nil
}

// GetLocale retrieves a locale by ID.
func (d Database) GetLocale(id types.LocaleID) (*Locale, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetLocale(d.Context, mdb.GetLocaleParams{LocaleID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get locale: %w", err)
	}
	res := d.MapLocale(row)
	return &res, nil
}

// GetLocaleByCode retrieves a locale by its code.
func (d Database) GetLocaleByCode(code string) (*Locale, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetLocaleByCode(d.Context, mdb.GetLocaleByCodeParams{Code: code})
	if err != nil {
		return nil, fmt.Errorf("failed to get locale by code: %w", err)
	}
	res := d.MapLocale(row)
	return &res, nil
}

// GetDefaultLocale retrieves the default locale.
func (d Database) GetDefaultLocale() (*Locale, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetDefaultLocale(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get default locale: %w", err)
	}
	res := d.MapLocale(row)
	return &res, nil
}

// ListLocales retrieves all locales.
func (d Database) ListLocales() (*[]Locale, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListLocales(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list locales: %w", err)
	}
	res := []Locale{}
	for _, v := range rows {
		res = append(res, d.MapLocale(v))
	}
	return &res, nil
}

// ListEnabledLocales retrieves all enabled locales.
func (d Database) ListEnabledLocales() (*[]Locale, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListEnabledLocales(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled locales: %w", err)
	}
	res := []Locale{}
	for _, v := range rows {
		res = append(res, d.MapLocale(v))
	}
	return &res, nil
}

// ListLocalesPaginated retrieves locales with pagination.
func (d Database) ListLocalesPaginated(params PaginationParams) (*[]Locale, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListLocalesPaginated(d.Context, mdb.ListLocalesPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list locales paginated: %w", err)
	}
	res := []Locale{}
	for _, v := range rows {
		res = append(res, d.MapLocale(v))
	}
	return &res, nil
}

// UpdateLocale updates a locale with audit trail.
func (d Database) UpdateLocale(ctx context.Context, ac audited.AuditContext, s UpdateLocaleParams) error {
	cmd := d.UpdateLocaleCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// DeleteLocale deletes a locale with audit trail.
func (d Database) DeleteLocale(ctx context.Context, ac audited.AuditContext, id types.LocaleID) error {
	cmd := d.DeleteLocaleCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// ClearDefaultLocale clears the default flag on all locales.
func (d Database) ClearDefaultLocale(ctx context.Context) error {
	queries := mdb.New(d.Connection)
	return queries.ClearDefaultLocale(ctx)
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

// MapLocale converts a sqlc-generated MySQL type to the wrapper type.
func (d MysqlDatabase) MapLocale(a mdbm.Locale) Locale {
	return Locale{
		LocaleID:     a.LocaleID,
		Code:         a.Code,
		Label:        a.Label,
		IsDefault:    a.IsDefault.Bool(),
		IsEnabled:    a.IsEnabled.Bool(),
		FallbackCode: ReadNullString(a.FallbackCode),
		SortOrder:    a.SortOrder,
		DateCreated:  a.DateCreated,
	}
}

// MapCreateLocaleParams converts wrapper params to a sqlc-generated MySQL type.
func (d MysqlDatabase) MapCreateLocaleParams(a CreateLocaleParams) mdbm.CreateLocaleParams {
	return mdbm.CreateLocaleParams{
		LocaleID:     types.NewLocaleID(),
		Code:         a.Code,
		Label:        a.Label,
		IsDefault:    types.NewSafeBool(a.IsDefault),
		IsEnabled:    types.NewSafeBool(a.IsEnabled),
		FallbackCode: StringToNullString(a.FallbackCode),
		SortOrder:    a.SortOrder,
		DateCreated:  a.DateCreated,
	}
}

// MapUpdateLocaleParams converts wrapper params to a sqlc-generated MySQL type.
func (d MysqlDatabase) MapUpdateLocaleParams(a UpdateLocaleParams) mdbm.UpdateLocaleParams {
	return mdbm.UpdateLocaleParams{
		Code:         a.Code,
		Label:        a.Label,
		IsDefault:    types.NewSafeBool(a.IsDefault),
		IsEnabled:    types.NewSafeBool(a.IsEnabled),
		FallbackCode: StringToNullString(a.FallbackCode),
		SortOrder:    a.SortOrder,
		DateCreated:  a.DateCreated,
		LocaleID:     a.LocaleID,
	}
}

// QUERIES

// CountLocales returns the total count of locales.
func (d MysqlDatabase) CountLocales() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountLocale(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count locales: %w", err)
	}
	return &c, nil
}

// CreateLocaleTable creates the locales table.
func (d MysqlDatabase) CreateLocaleTable() error {
	queries := mdbm.New(d.Connection)
	return queries.CreateLocaleTable(d.Context)
}

// DropLocaleTable drops the locales table.
func (d MysqlDatabase) DropLocaleTable() error {
	queries := mdbm.New(d.Connection)
	return queries.DropLocaleTable(d.Context)
}

// CreateLocale creates a new locale with audit trail.
func (d MysqlDatabase) CreateLocale(ctx context.Context, ac audited.AuditContext, s CreateLocaleParams) (*Locale, error) {
	cmd := d.NewLocaleCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create locale: %w", err)
	}
	r := d.MapLocale(result)
	return &r, nil
}

// GetLocale retrieves a locale by ID.
func (d MysqlDatabase) GetLocale(id types.LocaleID) (*Locale, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetLocale(d.Context, mdbm.GetLocaleParams{LocaleID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get locale: %w", err)
	}
	res := d.MapLocale(row)
	return &res, nil
}

// GetLocaleByCode retrieves a locale by its code.
func (d MysqlDatabase) GetLocaleByCode(code string) (*Locale, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetLocaleByCode(d.Context, mdbm.GetLocaleByCodeParams{Code: code})
	if err != nil {
		return nil, fmt.Errorf("failed to get locale by code: %w", err)
	}
	res := d.MapLocale(row)
	return &res, nil
}

// GetDefaultLocale retrieves the default locale.
func (d MysqlDatabase) GetDefaultLocale() (*Locale, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetDefaultLocale(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get default locale: %w", err)
	}
	res := d.MapLocale(row)
	return &res, nil
}

// ListLocales retrieves all locales.
func (d MysqlDatabase) ListLocales() (*[]Locale, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListLocales(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list locales: %w", err)
	}
	res := []Locale{}
	for _, v := range rows {
		res = append(res, d.MapLocale(v))
	}
	return &res, nil
}

// ListEnabledLocales retrieves all enabled locales.
func (d MysqlDatabase) ListEnabledLocales() (*[]Locale, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListEnabledLocales(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled locales: %w", err)
	}
	res := []Locale{}
	for _, v := range rows {
		res = append(res, d.MapLocale(v))
	}
	return &res, nil
}

// ListLocalesPaginated retrieves locales with pagination.
func (d MysqlDatabase) ListLocalesPaginated(params PaginationParams) (*[]Locale, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListLocalesPaginated(d.Context, mdbm.ListLocalesPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list locales paginated: %w", err)
	}
	res := []Locale{}
	for _, v := range rows {
		res = append(res, d.MapLocale(v))
	}
	return &res, nil
}

// UpdateLocale updates a locale with audit trail.
func (d MysqlDatabase) UpdateLocale(ctx context.Context, ac audited.AuditContext, s UpdateLocaleParams) error {
	cmd := d.UpdateLocaleCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// DeleteLocale deletes a locale with audit trail.
func (d MysqlDatabase) DeleteLocale(ctx context.Context, ac audited.AuditContext, id types.LocaleID) error {
	cmd := d.DeleteLocaleCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// ClearDefaultLocale clears the default flag on all locales.
func (d MysqlDatabase) ClearDefaultLocale(ctx context.Context) error {
	queries := mdbm.New(d.Connection)
	return queries.ClearDefaultLocale(ctx)
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

// MapLocale converts a sqlc-generated PostgreSQL type to the wrapper type.
func (d PsqlDatabase) MapLocale(a mdbp.Locale) Locale {
	return Locale{
		LocaleID:     a.LocaleID,
		Code:         a.Code,
		Label:        a.Label,
		IsDefault:    a.IsDefault.Bool(),
		IsEnabled:    a.IsEnabled.Bool(),
		FallbackCode: ReadNullString(a.FallbackCode),
		SortOrder:    a.SortOrder,
		DateCreated:  a.DateCreated,
	}
}

// MapCreateLocaleParams converts wrapper params to a sqlc-generated PostgreSQL type.
func (d PsqlDatabase) MapCreateLocaleParams(a CreateLocaleParams) mdbp.CreateLocaleParams {
	return mdbp.CreateLocaleParams{
		LocaleID:     types.NewLocaleID(),
		Code:         a.Code,
		Label:        a.Label,
		IsDefault:    types.NewSafeBool(a.IsDefault),
		IsEnabled:    types.NewSafeBool(a.IsEnabled),
		FallbackCode: StringToNullString(a.FallbackCode),
		SortOrder:    a.SortOrder,
		DateCreated:  a.DateCreated,
	}
}

// MapUpdateLocaleParams converts wrapper params to a sqlc-generated PostgreSQL type.
func (d PsqlDatabase) MapUpdateLocaleParams(a UpdateLocaleParams) mdbp.UpdateLocaleParams {
	return mdbp.UpdateLocaleParams{
		Code:         a.Code,
		Label:        a.Label,
		IsDefault:    types.NewSafeBool(a.IsDefault),
		IsEnabled:    types.NewSafeBool(a.IsEnabled),
		FallbackCode: StringToNullString(a.FallbackCode),
		SortOrder:    a.SortOrder,
		DateCreated:  a.DateCreated,
		LocaleID:     a.LocaleID,
	}
}

// QUERIES

// CountLocales returns the total count of locales.
func (d PsqlDatabase) CountLocales() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountLocale(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count locales: %w", err)
	}
	return &c, nil
}

// CreateLocaleTable creates the locales table.
func (d PsqlDatabase) CreateLocaleTable() error {
	queries := mdbp.New(d.Connection)
	return queries.CreateLocaleTable(d.Context)
}

// DropLocaleTable drops the locales table.
func (d PsqlDatabase) DropLocaleTable() error {
	queries := mdbp.New(d.Connection)
	return queries.DropLocaleTable(d.Context)
}

// CreateLocale creates a new locale with audit trail.
func (d PsqlDatabase) CreateLocale(ctx context.Context, ac audited.AuditContext, s CreateLocaleParams) (*Locale, error) {
	cmd := d.NewLocaleCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create locale: %w", err)
	}
	r := d.MapLocale(result)
	return &r, nil
}

// GetLocale retrieves a locale by ID.
func (d PsqlDatabase) GetLocale(id types.LocaleID) (*Locale, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetLocale(d.Context, mdbp.GetLocaleParams{LocaleID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get locale: %w", err)
	}
	res := d.MapLocale(row)
	return &res, nil
}

// GetLocaleByCode retrieves a locale by its code.
func (d PsqlDatabase) GetLocaleByCode(code string) (*Locale, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetLocaleByCode(d.Context, mdbp.GetLocaleByCodeParams{Code: code})
	if err != nil {
		return nil, fmt.Errorf("failed to get locale by code: %w", err)
	}
	res := d.MapLocale(row)
	return &res, nil
}

// GetDefaultLocale retrieves the default locale.
func (d PsqlDatabase) GetDefaultLocale() (*Locale, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetDefaultLocale(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get default locale: %w", err)
	}
	res := d.MapLocale(row)
	return &res, nil
}

// ListLocales retrieves all locales.
func (d PsqlDatabase) ListLocales() (*[]Locale, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListLocales(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list locales: %w", err)
	}
	res := []Locale{}
	for _, v := range rows {
		res = append(res, d.MapLocale(v))
	}
	return &res, nil
}

// ListEnabledLocales retrieves all enabled locales.
func (d PsqlDatabase) ListEnabledLocales() (*[]Locale, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListEnabledLocales(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled locales: %w", err)
	}
	res := []Locale{}
	for _, v := range rows {
		res = append(res, d.MapLocale(v))
	}
	return &res, nil
}

// ListLocalesPaginated retrieves locales with pagination.
func (d PsqlDatabase) ListLocalesPaginated(params PaginationParams) (*[]Locale, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListLocalesPaginated(d.Context, mdbp.ListLocalesPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list locales paginated: %w", err)
	}
	res := []Locale{}
	for _, v := range rows {
		res = append(res, d.MapLocale(v))
	}
	return &res, nil
}

// UpdateLocale updates a locale with audit trail.
func (d PsqlDatabase) UpdateLocale(ctx context.Context, ac audited.AuditContext, s UpdateLocaleParams) error {
	cmd := d.UpdateLocaleCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// DeleteLocale deletes a locale with audit trail.
func (d PsqlDatabase) DeleteLocale(ctx context.Context, ac audited.AuditContext, id types.LocaleID) error {
	cmd := d.DeleteLocaleCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// ClearDefaultLocale clears the default flag on all locales.
func (d PsqlDatabase) ClearDefaultLocale(ctx context.Context) error {
	queries := mdbp.New(d.Connection)
	return queries.ClearDefaultLocale(ctx)
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ----- SQLite CREATE -----

// NewLocaleCmd is an audited command for creating a locale.
type NewLocaleCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateLocaleParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c NewLocaleCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewLocaleCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewLocaleCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewLocaleCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewLocaleCmd) TableName() string { return "locales" }

// Params returns the command parameters.
func (c NewLocaleCmd) Params() any { return c.params }

// GetID returns the ID from a locale.
func (c NewLocaleCmd) GetID(r mdb.Locale) string {
	return string(r.LocaleID)
}

// Execute creates the locale in the database.
func (c NewLocaleCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Locale, error) {
	queries := mdb.New(tx)
	return queries.CreateLocale(ctx, mdb.CreateLocaleParams{
		LocaleID:     types.NewLocaleID(),
		Code:         c.params.Code,
		Label:        c.params.Label,
		IsDefault:    types.NewSafeBool(c.params.IsDefault),
		IsEnabled:    types.NewSafeBool(c.params.IsEnabled),
		FallbackCode: StringToNullString(c.params.FallbackCode),
		SortOrder:    c.params.SortOrder,
		DateCreated:  c.params.DateCreated,
	})
}

// NewLocaleCmd creates a new create command for a locale.
func (d Database) NewLocaleCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateLocaleParams) NewLocaleCmd {
	return NewLocaleCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE -----

// UpdateLocaleCmd is an audited command for updating a locale.
type UpdateLocaleCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateLocaleParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c UpdateLocaleCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateLocaleCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateLocaleCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateLocaleCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c UpdateLocaleCmd) TableName() string { return "locales" }

// Params returns the command parameters.
func (c UpdateLocaleCmd) Params() any { return c.params }

// GetID returns the locale ID.
func (c UpdateLocaleCmd) GetID() string { return string(c.params.LocaleID) }

// GetBefore retrieves the locale before the update.
func (c UpdateLocaleCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Locale, error) {
	queries := mdb.New(tx)
	return queries.GetLocale(ctx, mdb.GetLocaleParams{LocaleID: c.params.LocaleID})
}

// Execute updates the locale in the database.
func (c UpdateLocaleCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateLocale(ctx, mdb.UpdateLocaleParams{
		Code:         c.params.Code,
		Label:        c.params.Label,
		IsDefault:    types.NewSafeBool(c.params.IsDefault),
		IsEnabled:    types.NewSafeBool(c.params.IsEnabled),
		FallbackCode: StringToNullString(c.params.FallbackCode),
		SortOrder:    c.params.SortOrder,
		DateCreated:  c.params.DateCreated,
		LocaleID:     c.params.LocaleID,
	})
}

// UpdateLocaleCmd creates a new update command for a locale.
func (d Database) UpdateLocaleCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateLocaleParams) UpdateLocaleCmd {
	return UpdateLocaleCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

// DeleteLocaleCmd is an audited command for deleting a locale.
type DeleteLocaleCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.LocaleID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c DeleteLocaleCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteLocaleCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteLocaleCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteLocaleCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteLocaleCmd) TableName() string { return "locales" }

// GetID returns the locale ID.
func (c DeleteLocaleCmd) GetID() string { return string(c.id) }

// GetBefore retrieves the locale before deletion.
func (c DeleteLocaleCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Locale, error) {
	queries := mdb.New(tx)
	return queries.GetLocale(ctx, mdb.GetLocaleParams{LocaleID: c.id})
}

// Execute deletes the locale from the database.
func (c DeleteLocaleCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteLocale(ctx, mdb.DeleteLocaleParams{LocaleID: c.id})
}

// DeleteLocaleCmd creates a new delete command for a locale.
func (d Database) DeleteLocaleCmd(ctx context.Context, auditCtx audited.AuditContext, id types.LocaleID) DeleteLocaleCmd {
	return DeleteLocaleCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

// NewLocaleCmdMysql is an audited command for creating a locale on MySQL.
type NewLocaleCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateLocaleParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c NewLocaleCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewLocaleCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewLocaleCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewLocaleCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewLocaleCmdMysql) TableName() string { return "locales" }

// Params returns the command parameters.
func (c NewLocaleCmdMysql) Params() any { return c.params }

// GetID returns the ID from a locale.
func (c NewLocaleCmdMysql) GetID(r mdbm.Locale) string {
	return string(r.LocaleID)
}

// Execute creates the locale in the database.
func (c NewLocaleCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Locale, error) {
	id := types.NewLocaleID()
	queries := mdbm.New(tx)
	params := mdbm.CreateLocaleParams{
		LocaleID:     id,
		Code:         c.params.Code,
		Label:        c.params.Label,
		IsDefault:    types.NewSafeBool(c.params.IsDefault),
		IsEnabled:    types.NewSafeBool(c.params.IsEnabled),
		FallbackCode: StringToNullString(c.params.FallbackCode),
		SortOrder:    c.params.SortOrder,
		DateCreated:  c.params.DateCreated,
	}
	if err := queries.CreateLocale(ctx, params); err != nil {
		return mdbm.Locale{}, err
	}
	return queries.GetLocale(ctx, mdbm.GetLocaleParams{LocaleID: id})
}

// NewLocaleCmd creates a new create command for a locale.
func (d MysqlDatabase) NewLocaleCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateLocaleParams) NewLocaleCmdMysql {
	return NewLocaleCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE -----

// UpdateLocaleCmdMysql is an audited command for updating a locale on MySQL.
type UpdateLocaleCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateLocaleParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c UpdateLocaleCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateLocaleCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateLocaleCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateLocaleCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c UpdateLocaleCmdMysql) TableName() string { return "locales" }

// Params returns the command parameters.
func (c UpdateLocaleCmdMysql) Params() any { return c.params }

// GetID returns the locale ID.
func (c UpdateLocaleCmdMysql) GetID() string { return string(c.params.LocaleID) }

// GetBefore retrieves the locale before the update.
func (c UpdateLocaleCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Locale, error) {
	queries := mdbm.New(tx)
	return queries.GetLocale(ctx, mdbm.GetLocaleParams{LocaleID: c.params.LocaleID})
}

// Execute updates the locale in the database.
func (c UpdateLocaleCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateLocale(ctx, mdbm.UpdateLocaleParams{
		Code:         c.params.Code,
		Label:        c.params.Label,
		IsDefault:    types.NewSafeBool(c.params.IsDefault),
		IsEnabled:    types.NewSafeBool(c.params.IsEnabled),
		FallbackCode: StringToNullString(c.params.FallbackCode),
		SortOrder:    c.params.SortOrder,
		DateCreated:  c.params.DateCreated,
		LocaleID:     c.params.LocaleID,
	})
}

// UpdateLocaleCmd creates a new update command for a locale.
func (d MysqlDatabase) UpdateLocaleCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateLocaleParams) UpdateLocaleCmdMysql {
	return UpdateLocaleCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

// DeleteLocaleCmdMysql is an audited command for deleting a locale on MySQL.
type DeleteLocaleCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.LocaleID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c DeleteLocaleCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteLocaleCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteLocaleCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteLocaleCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteLocaleCmdMysql) TableName() string { return "locales" }

// GetID returns the locale ID.
func (c DeleteLocaleCmdMysql) GetID() string { return string(c.id) }

// GetBefore retrieves the locale before deletion.
func (c DeleteLocaleCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Locale, error) {
	queries := mdbm.New(tx)
	return queries.GetLocale(ctx, mdbm.GetLocaleParams{LocaleID: c.id})
}

// Execute deletes the locale from the database.
func (c DeleteLocaleCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteLocale(ctx, mdbm.DeleteLocaleParams{LocaleID: c.id})
}

// DeleteLocaleCmd creates a new delete command for a locale.
func (d MysqlDatabase) DeleteLocaleCmd(ctx context.Context, auditCtx audited.AuditContext, id types.LocaleID) DeleteLocaleCmdMysql {
	return DeleteLocaleCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

// NewLocaleCmdPsql is an audited command for creating a locale on PostgreSQL.
type NewLocaleCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateLocaleParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c NewLocaleCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewLocaleCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewLocaleCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewLocaleCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewLocaleCmdPsql) TableName() string { return "locales" }

// Params returns the command parameters.
func (c NewLocaleCmdPsql) Params() any { return c.params }

// GetID returns the ID from a locale.
func (c NewLocaleCmdPsql) GetID(r mdbp.Locale) string {
	return string(r.LocaleID)
}

// Execute creates the locale in the database.
func (c NewLocaleCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Locale, error) {
	queries := mdbp.New(tx)
	return queries.CreateLocale(ctx, mdbp.CreateLocaleParams{
		LocaleID:     types.NewLocaleID(),
		Code:         c.params.Code,
		Label:        c.params.Label,
		IsDefault:    types.NewSafeBool(c.params.IsDefault),
		IsEnabled:    types.NewSafeBool(c.params.IsEnabled),
		FallbackCode: StringToNullString(c.params.FallbackCode),
		SortOrder:    c.params.SortOrder,
		DateCreated:  c.params.DateCreated,
	})
}

// NewLocaleCmd creates a new create command for a locale.
func (d PsqlDatabase) NewLocaleCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateLocaleParams) NewLocaleCmdPsql {
	return NewLocaleCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE -----

// UpdateLocaleCmdPsql is an audited command for updating a locale on PostgreSQL.
type UpdateLocaleCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateLocaleParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c UpdateLocaleCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateLocaleCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateLocaleCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateLocaleCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c UpdateLocaleCmdPsql) TableName() string { return "locales" }

// Params returns the command parameters.
func (c UpdateLocaleCmdPsql) Params() any { return c.params }

// GetID returns the locale ID.
func (c UpdateLocaleCmdPsql) GetID() string { return string(c.params.LocaleID) }

// GetBefore retrieves the locale before the update.
func (c UpdateLocaleCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Locale, error) {
	queries := mdbp.New(tx)
	return queries.GetLocale(ctx, mdbp.GetLocaleParams{LocaleID: c.params.LocaleID})
}

// Execute updates the locale in the database.
func (c UpdateLocaleCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateLocale(ctx, mdbp.UpdateLocaleParams{
		Code:         c.params.Code,
		Label:        c.params.Label,
		IsDefault:    types.NewSafeBool(c.params.IsDefault),
		IsEnabled:    types.NewSafeBool(c.params.IsEnabled),
		FallbackCode: StringToNullString(c.params.FallbackCode),
		SortOrder:    c.params.SortOrder,
		DateCreated:  c.params.DateCreated,
		LocaleID:     c.params.LocaleID,
	})
}

// UpdateLocaleCmd creates a new update command for a locale.
func (d PsqlDatabase) UpdateLocaleCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateLocaleParams) UpdateLocaleCmdPsql {
	return UpdateLocaleCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

// DeleteLocaleCmdPsql is an audited command for deleting a locale on PostgreSQL.
type DeleteLocaleCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.LocaleID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c DeleteLocaleCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteLocaleCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteLocaleCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteLocaleCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteLocaleCmdPsql) TableName() string { return "locales" }

// GetID returns the locale ID.
func (c DeleteLocaleCmdPsql) GetID() string { return string(c.id) }

// GetBefore retrieves the locale before deletion.
func (c DeleteLocaleCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Locale, error) {
	queries := mdbp.New(tx)
	return queries.GetLocale(ctx, mdbp.GetLocaleParams{LocaleID: c.id})
}

// Execute deletes the locale from the database.
func (c DeleteLocaleCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteLocale(ctx, mdbp.DeleteLocaleParams{LocaleID: c.id})
}

// DeleteLocaleCmd creates a new delete command for a locale.
func (d PsqlDatabase) DeleteLocaleCmd(ctx context.Context, auditCtx audited.AuditContext, id types.LocaleID) DeleteLocaleCmdPsql {
	return DeleteLocaleCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}

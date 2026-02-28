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

// AdminContentVersion represents a version snapshot of admin content.
type AdminContentVersion struct {
	AdminContentVersionID types.AdminContentVersionID `json:"admin_content_version_id"`
	AdminContentDataID    types.AdminContentID        `json:"admin_content_data_id"`
	VersionNumber         int64                       `json:"version_number"`
	Locale                string                      `json:"locale"`
	Snapshot              string                      `json:"snapshot"`
	Trigger               string                      `json:"trigger"`
	Label                 string                      `json:"label"`
	Published             bool                        `json:"published"`
	PublishedBy           types.NullableUserID        `json:"published_by"`
	DateCreated           types.Timestamp             `json:"date_created"`
}

// CreateAdminContentVersionParams contains fields for creating a new admin content version.
type CreateAdminContentVersionParams struct {
	AdminContentDataID types.AdminContentID `json:"admin_content_data_id"`
	VersionNumber      int64                `json:"version_number"`
	Locale             string               `json:"locale"`
	Snapshot           string               `json:"snapshot"`
	Trigger            string               `json:"trigger"`
	Label              string               `json:"label"`
	Published          bool                 `json:"published"`
	PublishedBy        types.NullableUserID `json:"published_by"`
	DateCreated        types.Timestamp      `json:"date_created"`
}

// StringAdminContentVersion is the string representation for TUI table display.
type StringAdminContentVersion struct {
	AdminContentVersionID string `json:"admin_content_version_id"`
	AdminContentDataID    string `json:"admin_content_data_id"`
	VersionNumber         string `json:"version_number"`
	Locale                string `json:"locale"`
	Snapshot              string `json:"snapshot"`
	Trigger               string `json:"trigger"`
	Label                 string `json:"label"`
	Published             string `json:"published"`
	PublishedBy           string `json:"published_by"`
	DateCreated           string `json:"date_created"`
}

// MapStringAdminContentVersion converts AdminContentVersion to StringAdminContentVersion for table display.
func MapStringAdminContentVersion(a AdminContentVersion) StringAdminContentVersion {
	return StringAdminContentVersion{
		AdminContentVersionID: a.AdminContentVersionID.String(),
		AdminContentDataID:    a.AdminContentDataID.String(),
		VersionNumber:         fmt.Sprintf("%d", a.VersionNumber),
		Locale:                a.Locale,
		Snapshot:              a.Snapshot,
		Trigger:               a.Trigger,
		Label:                 a.Label,
		Published:             fmt.Sprintf("%t", a.Published),
		PublishedBy:           a.PublishedBy.String(),
		DateCreated:           a.DateCreated.String(),
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

// MapAdminContentVersion converts a sqlc-generated SQLite type to the wrapper type.
func (d Database) MapAdminContentVersion(a mdb.AdminContentVersions) AdminContentVersion {
	return AdminContentVersion{
		AdminContentVersionID: a.AdminContentVersionID,
		AdminContentDataID:    a.AdminContentDataID,
		VersionNumber:         a.VersionNumber,
		Locale:                a.Locale,
		Snapshot:              a.Snapshot,
		Trigger:               a.Trigger,
		Label:                 a.Label,
		Published:             a.Published != 0,
		PublishedBy:           a.PublishedBy,
		DateCreated:           a.DateCreated,
	}
}

// MapCreateAdminContentVersionParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapCreateAdminContentVersionParams(a CreateAdminContentVersionParams) mdb.CreateAdminContentVersionParams {
	return mdb.CreateAdminContentVersionParams{
		AdminContentVersionID: types.NewAdminContentVersionID(),
		AdminContentDataID:    a.AdminContentDataID,
		VersionNumber:         a.VersionNumber,
		Locale:                a.Locale,
		Snapshot:              a.Snapshot,
		Trigger:               a.Trigger,
		Label:                 a.Label,
		Published:             boolToInt64(a.Published),
		PublishedBy:           a.PublishedBy,
		DateCreated:           a.DateCreated,
	}
}

// QUERIES

// CountAdminContentVersions returns the total count of admin content versions.
func (d Database) CountAdminContentVersions() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminContentVersions(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count admin content versions: %w", err)
	}
	return &c, nil
}

// CountAdminContentVersionsByContent returns the count of admin content versions for a content item.
func (d Database) CountAdminContentVersionsByContent(adminContentDataID types.AdminContentID) (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminContentVersionsByContent(d.Context, mdb.CountAdminContentVersionsByContentParams{
		AdminContentDataID: adminContentDataID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count admin content versions by content: %w", err)
	}
	return &c, nil
}

// CreateAdminContentVersionTable creates the admin_content_versions table.
func (d Database) CreateAdminContentVersionTable() error {
	queries := mdb.New(d.Connection)
	return queries.CreateAdminContentVersionTable(d.Context)
}

// DropAdminContentVersionTable drops the admin_content_versions table.
func (d Database) DropAdminContentVersionTable() error {
	queries := mdb.New(d.Connection)
	return queries.DropAdminContentVersionTable(d.Context)
}

// CreateAdminContentVersion inserts a new admin content version.
func (d Database) CreateAdminContentVersion(ctx context.Context, ac audited.AuditContext, s CreateAdminContentVersionParams) (*AdminContentVersion, error) {
	cmd := d.NewAdminContentVersionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin content version: %w", err)
	}
	r := d.MapAdminContentVersion(result)
	return &r, nil
}

// DeleteAdminContentVersion removes a record.
func (d Database) DeleteAdminContentVersion(ctx context.Context, ac audited.AuditContext, id types.AdminContentVersionID) error {
	cmd := d.DeleteAdminContentVersionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetAdminContentVersion retrieves a record by ID.
func (d Database) GetAdminContentVersion(id types.AdminContentVersionID) (*AdminContentVersion, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminContentVersion(d.Context, mdb.GetAdminContentVersionParams{AdminContentVersionID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin content version: %w", err)
	}
	res := d.MapAdminContentVersion(row)
	return &res, nil
}

// GetAdminPublishedSnapshot retrieves the published version for a content item and locale.
func (d Database) GetAdminPublishedSnapshot(adminContentDataID types.AdminContentID, locale string) (*AdminContentVersion, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminPublishedSnapshot(d.Context, mdb.GetAdminPublishedSnapshotParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin published snapshot: %w", err)
	}
	res := d.MapAdminContentVersion(row)
	return &res, nil
}

// ListAdminContentVersionsByContent retrieves all versions for a content item.
func (d Database) ListAdminContentVersionsByContent(adminContentDataID types.AdminContentID) (*[]AdminContentVersion, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentVersionsByContent(d.Context, mdb.ListAdminContentVersionsByContentParams{
		AdminContentDataID: adminContentDataID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content versions by content: %w", err)
	}
	res := []AdminContentVersion{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentVersion(v))
	}
	return &res, nil
}

// ListAdminContentVersionsByContentLocale retrieves all versions for a content item and locale.
func (d Database) ListAdminContentVersionsByContentLocale(adminContentDataID types.AdminContentID, locale string) (*[]AdminContentVersion, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentVersionsByContentLocale(d.Context, mdb.ListAdminContentVersionsByContentLocaleParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content versions by content locale: %w", err)
	}
	res := []AdminContentVersion{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentVersion(v))
	}
	return &res, nil
}

// ClearAdminPublishedFlag clears the published flag for all versions of a content item and locale.
func (d Database) ClearAdminPublishedFlag(adminContentDataID types.AdminContentID, locale string) error {
	queries := mdb.New(d.Connection)
	return queries.ClearAdminPublishedFlag(d.Context, mdb.ClearAdminPublishedFlagParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
	})
}

// GetAdminMaxVersionNumber returns the maximum version number for a content item and locale.
func (d Database) GetAdminMaxVersionNumber(adminContentDataID types.AdminContentID, locale string) (int64, error) {
	queries := mdb.New(d.Connection)
	result, err := queries.GetAdminMaxVersionNumber(d.Context, mdb.GetAdminMaxVersionNumberParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get admin max version number: %w", err)
	}
	v, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("failed to get admin max version number: unexpected type %T", result)
	}
	return v, nil
}

// PruneAdminOldVersions removes old unlabeled unpublished versions for a content item and locale.
func (d Database) PruneAdminOldVersions(adminContentDataID types.AdminContentID, locale string, limit int64) error {
	queries := mdb.New(d.Connection)
	return queries.PruneAdminOldVersions(d.Context, mdb.PruneAdminOldVersionsParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
		Limit:              limit,
	})
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

// MapAdminContentVersion converts a sqlc-generated MySQL type to the wrapper type.
func (d MysqlDatabase) MapAdminContentVersion(a mdbm.AdminContentVersions) AdminContentVersion {
	return AdminContentVersion{
		AdminContentVersionID: a.AdminContentVersionID,
		AdminContentDataID:    a.AdminContentDataID,
		VersionNumber:         int64(a.VersionNumber),
		Locale:                a.Locale,
		Snapshot:              a.Snapshot,
		Trigger:               a.Trigger,
		Label:                 a.Label,
		Published:             a.Published != 0,
		PublishedBy:           a.PublishedBy,
		DateCreated:           a.DateCreated,
	}
}

// MapCreateAdminContentVersionParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapCreateAdminContentVersionParams(a CreateAdminContentVersionParams) mdbm.CreateAdminContentVersionParams {
	return mdbm.CreateAdminContentVersionParams{
		AdminContentVersionID: types.NewAdminContentVersionID(),
		AdminContentDataID:    a.AdminContentDataID,
		VersionNumber:         int32(a.VersionNumber),
		Locale:                a.Locale,
		Snapshot:              a.Snapshot,
		Trigger:               a.Trigger,
		Label:                 a.Label,
		Published:             boolToInt8(a.Published),
		PublishedBy:           a.PublishedBy,
		DateCreated:           a.DateCreated,
	}
}

// QUERIES

// CountAdminContentVersions returns the total count of admin content versions.
func (d MysqlDatabase) CountAdminContentVersions() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminContentVersions(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count admin content versions: %w", err)
	}
	return &c, nil
}

// CountAdminContentVersionsByContent returns the count of admin content versions for a content item.
func (d MysqlDatabase) CountAdminContentVersionsByContent(adminContentDataID types.AdminContentID) (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminContentVersionsByContent(d.Context, mdbm.CountAdminContentVersionsByContentParams{
		AdminContentDataID: adminContentDataID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count admin content versions by content: %w", err)
	}
	return &c, nil
}

// CreateAdminContentVersionTable creates the admin_content_versions table.
func (d MysqlDatabase) CreateAdminContentVersionTable() error {
	queries := mdbm.New(d.Connection)
	return queries.CreateAdminContentVersionTable(d.Context)
}

// DropAdminContentVersionTable drops the admin_content_versions table.
func (d MysqlDatabase) DropAdminContentVersionTable() error {
	queries := mdbm.New(d.Connection)
	return queries.DropAdminContentVersionTable(d.Context)
}

// CreateAdminContentVersion inserts a new admin content version.
func (d MysqlDatabase) CreateAdminContentVersion(ctx context.Context, ac audited.AuditContext, s CreateAdminContentVersionParams) (*AdminContentVersion, error) {
	cmd := d.NewAdminContentVersionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin content version: %w", err)
	}
	r := d.MapAdminContentVersion(result)
	return &r, nil
}

// DeleteAdminContentVersion removes a record.
func (d MysqlDatabase) DeleteAdminContentVersion(ctx context.Context, ac audited.AuditContext, id types.AdminContentVersionID) error {
	cmd := d.DeleteAdminContentVersionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetAdminContentVersion retrieves a record by ID.
func (d MysqlDatabase) GetAdminContentVersion(id types.AdminContentVersionID) (*AdminContentVersion, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminContentVersion(d.Context, mdbm.GetAdminContentVersionParams{AdminContentVersionID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin content version: %w", err)
	}
	res := d.MapAdminContentVersion(row)
	return &res, nil
}

// GetAdminPublishedSnapshot retrieves the published version for a content item and locale.
func (d MysqlDatabase) GetAdminPublishedSnapshot(adminContentDataID types.AdminContentID, locale string) (*AdminContentVersion, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminPublishedSnapshot(d.Context, mdbm.GetAdminPublishedSnapshotParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin published snapshot: %w", err)
	}
	res := d.MapAdminContentVersion(row)
	return &res, nil
}

// ListAdminContentVersionsByContent retrieves all versions for a content item.
func (d MysqlDatabase) ListAdminContentVersionsByContent(adminContentDataID types.AdminContentID) (*[]AdminContentVersion, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentVersionsByContent(d.Context, mdbm.ListAdminContentVersionsByContentParams{
		AdminContentDataID: adminContentDataID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content versions by content: %w", err)
	}
	res := []AdminContentVersion{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentVersion(v))
	}
	return &res, nil
}

// ListAdminContentVersionsByContentLocale retrieves all versions for a content item and locale.
func (d MysqlDatabase) ListAdminContentVersionsByContentLocale(adminContentDataID types.AdminContentID, locale string) (*[]AdminContentVersion, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentVersionsByContentLocale(d.Context, mdbm.ListAdminContentVersionsByContentLocaleParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content versions by content locale: %w", err)
	}
	res := []AdminContentVersion{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentVersion(v))
	}
	return &res, nil
}

// ClearAdminPublishedFlag clears the published flag for all versions of a content item and locale.
func (d MysqlDatabase) ClearAdminPublishedFlag(adminContentDataID types.AdminContentID, locale string) error {
	queries := mdbm.New(d.Connection)
	return queries.ClearAdminPublishedFlag(d.Context, mdbm.ClearAdminPublishedFlagParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
	})
}

// GetAdminMaxVersionNumber returns the maximum version number for a content item and locale.
func (d MysqlDatabase) GetAdminMaxVersionNumber(adminContentDataID types.AdminContentID, locale string) (int64, error) {
	queries := mdbm.New(d.Connection)
	result, err := queries.GetAdminMaxVersionNumber(d.Context, mdbm.GetAdminMaxVersionNumberParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get admin max version number: %w", err)
	}
	v, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("failed to get admin max version number: unexpected type %T", result)
	}
	return v, nil
}

// PruneAdminOldVersions removes old unlabeled unpublished versions for a content item and locale.
func (d MysqlDatabase) PruneAdminOldVersions(adminContentDataID types.AdminContentID, locale string, limit int64) error {
	queries := mdbm.New(d.Connection)
	return queries.PruneAdminOldVersions(d.Context, mdbm.PruneAdminOldVersionsParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
		Limit:              int32(limit),
	})
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

// MapAdminContentVersion converts a sqlc-generated PostgreSQL type to the wrapper type.
func (d PsqlDatabase) MapAdminContentVersion(a mdbp.AdminContentVersions) AdminContentVersion {
	return AdminContentVersion{
		AdminContentVersionID: a.AdminContentVersionID,
		AdminContentDataID:    a.AdminContentDataID,
		VersionNumber:         int64(a.VersionNumber),
		Locale:                a.Locale,
		Snapshot:              a.Snapshot,
		Trigger:               a.Trigger,
		Label:                 a.Label,
		Published:             a.Published,
		PublishedBy:           a.PublishedBy,
		DateCreated:           a.DateCreated,
	}
}

// MapCreateAdminContentVersionParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapCreateAdminContentVersionParams(a CreateAdminContentVersionParams) mdbp.CreateAdminContentVersionParams {
	return mdbp.CreateAdminContentVersionParams{
		AdminContentVersionID: types.NewAdminContentVersionID(),
		AdminContentDataID:    a.AdminContentDataID,
		VersionNumber:         int32(a.VersionNumber),
		Locale:                a.Locale,
		Snapshot:              a.Snapshot,
		Trigger:               a.Trigger,
		Label:                 a.Label,
		Published:             a.Published,
		PublishedBy:           a.PublishedBy,
		DateCreated:           a.DateCreated,
	}
}

// QUERIES

// CountAdminContentVersions returns the total count of admin content versions.
func (d PsqlDatabase) CountAdminContentVersions() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminContentVersions(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count admin content versions: %w", err)
	}
	return &c, nil
}

// CountAdminContentVersionsByContent returns the count of admin content versions for a content item.
func (d PsqlDatabase) CountAdminContentVersionsByContent(adminContentDataID types.AdminContentID) (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminContentVersionsByContent(d.Context, mdbp.CountAdminContentVersionsByContentParams{
		AdminContentDataID: adminContentDataID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count admin content versions by content: %w", err)
	}
	return &c, nil
}

// CreateAdminContentVersionTable creates the admin_content_versions table.
func (d PsqlDatabase) CreateAdminContentVersionTable() error {
	queries := mdbp.New(d.Connection)
	return queries.CreateAdminContentVersionTable(d.Context)
}

// DropAdminContentVersionTable drops the admin_content_versions table.
func (d PsqlDatabase) DropAdminContentVersionTable() error {
	queries := mdbp.New(d.Connection)
	return queries.DropAdminContentVersionTable(d.Context)
}

// CreateAdminContentVersion inserts a new admin content version.
func (d PsqlDatabase) CreateAdminContentVersion(ctx context.Context, ac audited.AuditContext, s CreateAdminContentVersionParams) (*AdminContentVersion, error) {
	cmd := d.NewAdminContentVersionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin content version: %w", err)
	}
	r := d.MapAdminContentVersion(result)
	return &r, nil
}

// DeleteAdminContentVersion removes a record.
func (d PsqlDatabase) DeleteAdminContentVersion(ctx context.Context, ac audited.AuditContext, id types.AdminContentVersionID) error {
	cmd := d.DeleteAdminContentVersionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetAdminContentVersion retrieves a record by ID.
func (d PsqlDatabase) GetAdminContentVersion(id types.AdminContentVersionID) (*AdminContentVersion, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminContentVersion(d.Context, mdbp.GetAdminContentVersionParams{AdminContentVersionID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin content version: %w", err)
	}
	res := d.MapAdminContentVersion(row)
	return &res, nil
}

// GetAdminPublishedSnapshot retrieves the published version for a content item and locale.
func (d PsqlDatabase) GetAdminPublishedSnapshot(adminContentDataID types.AdminContentID, locale string) (*AdminContentVersion, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminPublishedSnapshot(d.Context, mdbp.GetAdminPublishedSnapshotParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin published snapshot: %w", err)
	}
	res := d.MapAdminContentVersion(row)
	return &res, nil
}

// ListAdminContentVersionsByContent retrieves all versions for a content item.
func (d PsqlDatabase) ListAdminContentVersionsByContent(adminContentDataID types.AdminContentID) (*[]AdminContentVersion, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentVersionsByContent(d.Context, mdbp.ListAdminContentVersionsByContentParams{
		AdminContentDataID: adminContentDataID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content versions by content: %w", err)
	}
	res := []AdminContentVersion{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentVersion(v))
	}
	return &res, nil
}

// ListAdminContentVersionsByContentLocale retrieves all versions for a content item and locale.
func (d PsqlDatabase) ListAdminContentVersionsByContentLocale(adminContentDataID types.AdminContentID, locale string) (*[]AdminContentVersion, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentVersionsByContentLocale(d.Context, mdbp.ListAdminContentVersionsByContentLocaleParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content versions by content locale: %w", err)
	}
	res := []AdminContentVersion{}
	for _, v := range rows {
		res = append(res, d.MapAdminContentVersion(v))
	}
	return &res, nil
}

// ClearAdminPublishedFlag clears the published flag for all versions of a content item and locale.
func (d PsqlDatabase) ClearAdminPublishedFlag(adminContentDataID types.AdminContentID, locale string) error {
	queries := mdbp.New(d.Connection)
	return queries.ClearAdminPublishedFlag(d.Context, mdbp.ClearAdminPublishedFlagParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
	})
}

// GetAdminMaxVersionNumber returns the maximum version number for a content item and locale.
func (d PsqlDatabase) GetAdminMaxVersionNumber(adminContentDataID types.AdminContentID, locale string) (int64, error) {
	queries := mdbp.New(d.Connection)
	result, err := queries.GetAdminMaxVersionNumber(d.Context, mdbp.GetAdminMaxVersionNumberParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get admin max version number: %w", err)
	}
	v, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("failed to get admin max version number: unexpected type %T", result)
	}
	return v, nil
}

// PruneAdminOldVersions removes old unlabeled unpublished versions for a content item and locale.
func (d PsqlDatabase) PruneAdminOldVersions(adminContentDataID types.AdminContentID, locale string, limit int64) error {
	queries := mdbp.New(d.Connection)
	return queries.PruneAdminOldVersions(d.Context, mdbp.PruneAdminOldVersionsParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
		Limit:              int32(limit),
	})
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ----- SQLite CREATE -----

// NewAdminContentVersionCmd is an audited command for create on admin_content_versions.
type NewAdminContentVersionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentVersionParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context.
func (c NewAdminContentVersionCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewAdminContentVersionCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewAdminContentVersionCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewAdminContentVersionCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewAdminContentVersionCmd) TableName() string { return "admin_content_versions" }

// Params returns the command parameters.
func (c NewAdminContentVersionCmd) Params() any { return c.params }

// GetID returns the record ID.
func (c NewAdminContentVersionCmd) GetID(r mdb.AdminContentVersions) string {
	return string(r.AdminContentVersionID)
}

// Execute performs the create operation.
func (c NewAdminContentVersionCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.AdminContentVersions, error) {
	queries := mdb.New(tx)
	return queries.CreateAdminContentVersion(ctx, mdb.CreateAdminContentVersionParams{
		AdminContentVersionID: types.NewAdminContentVersionID(),
		AdminContentDataID:    c.params.AdminContentDataID,
		VersionNumber:         c.params.VersionNumber,
		Locale:                c.params.Locale,
		Snapshot:              c.params.Snapshot,
		Trigger:               c.params.Trigger,
		Label:                 c.params.Label,
		Published:             boolToInt64(c.params.Published),
		PublishedBy:           c.params.PublishedBy,
		DateCreated:           c.params.DateCreated,
	})
}

// NewAdminContentVersionCmd creates a new audited command for create.
func (d Database) NewAdminContentVersionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentVersionParams) NewAdminContentVersionCmd {
	return NewAdminContentVersionCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

// DeleteAdminContentVersionCmd is an audited command for delete on admin_content_versions.
type DeleteAdminContentVersionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentVersionID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context.
func (c DeleteAdminContentVersionCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteAdminContentVersionCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteAdminContentVersionCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteAdminContentVersionCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteAdminContentVersionCmd) TableName() string { return "admin_content_versions" }

// GetID returns the record ID.
func (c DeleteAdminContentVersionCmd) GetID() string { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteAdminContentVersionCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminContentVersions, error) {
	queries := mdb.New(tx)
	return queries.GetAdminContentVersion(ctx, mdb.GetAdminContentVersionParams{AdminContentVersionID: c.id})
}

// Execute performs the delete operation.
func (c DeleteAdminContentVersionCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteAdminContentVersion(ctx, mdb.DeleteAdminContentVersionParams{AdminContentVersionID: c.id})
}

// DeleteAdminContentVersionCmd creates a new audited command for delete.
func (d Database) DeleteAdminContentVersionCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentVersionID) DeleteAdminContentVersionCmd {
	return DeleteAdminContentVersionCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

// NewAdminContentVersionCmdMysql is an audited command for create on admin_content_versions (MySQL).
type NewAdminContentVersionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentVersionParams
	id       types.AdminContentVersionID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context.
func (c NewAdminContentVersionCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewAdminContentVersionCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewAdminContentVersionCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewAdminContentVersionCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewAdminContentVersionCmdMysql) TableName() string { return "admin_content_versions" }

// Params returns the command parameters.
func (c NewAdminContentVersionCmdMysql) Params() any { return c.params }

// GetID returns the record ID.
func (c NewAdminContentVersionCmdMysql) GetID(r mdbm.AdminContentVersions) string {
	return string(r.AdminContentVersionID)
}

// Execute performs the create operation.
func (c NewAdminContentVersionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.AdminContentVersions, error) {
	queries := mdbm.New(tx)
	params := mdbm.CreateAdminContentVersionParams{
		AdminContentVersionID: c.id,
		AdminContentDataID:    c.params.AdminContentDataID,
		VersionNumber:         int32(c.params.VersionNumber),
		Locale:                c.params.Locale,
		Snapshot:              c.params.Snapshot,
		Trigger:               c.params.Trigger,
		Label:                 c.params.Label,
		Published:             boolToInt8(c.params.Published),
		PublishedBy:           c.params.PublishedBy,
		DateCreated:           c.params.DateCreated,
	}
	if err := queries.CreateAdminContentVersion(ctx, params); err != nil {
		return mdbm.AdminContentVersions{}, err
	}
	return queries.GetAdminContentVersion(ctx, mdbm.GetAdminContentVersionParams{AdminContentVersionID: c.id})
}

// NewAdminContentVersionCmd creates a new audited command for create.
func (d MysqlDatabase) NewAdminContentVersionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentVersionParams) NewAdminContentVersionCmdMysql {
	id := types.NewAdminContentVersionID()
	return NewAdminContentVersionCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

// DeleteAdminContentVersionCmdMysql is an audited command for delete on admin_content_versions (MySQL).
type DeleteAdminContentVersionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentVersionID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context.
func (c DeleteAdminContentVersionCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteAdminContentVersionCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteAdminContentVersionCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteAdminContentVersionCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteAdminContentVersionCmdMysql) TableName() string { return "admin_content_versions" }

// GetID returns the record ID.
func (c DeleteAdminContentVersionCmdMysql) GetID() string { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteAdminContentVersionCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminContentVersions, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminContentVersion(ctx, mdbm.GetAdminContentVersionParams{AdminContentVersionID: c.id})
}

// Execute performs the delete operation.
func (c DeleteAdminContentVersionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteAdminContentVersion(ctx, mdbm.DeleteAdminContentVersionParams{AdminContentVersionID: c.id})
}

// DeleteAdminContentVersionCmd creates a new audited command for delete.
func (d MysqlDatabase) DeleteAdminContentVersionCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentVersionID) DeleteAdminContentVersionCmdMysql {
	return DeleteAdminContentVersionCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

// NewAdminContentVersionCmdPsql is an audited command for create on admin_content_versions (PostgreSQL).
type NewAdminContentVersionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentVersionParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context.
func (c NewAdminContentVersionCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewAdminContentVersionCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewAdminContentVersionCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewAdminContentVersionCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewAdminContentVersionCmdPsql) TableName() string { return "admin_content_versions" }

// Params returns the command parameters.
func (c NewAdminContentVersionCmdPsql) Params() any { return c.params }

// GetID returns the record ID.
func (c NewAdminContentVersionCmdPsql) GetID(r mdbp.AdminContentVersions) string {
	return string(r.AdminContentVersionID)
}

// Execute performs the create operation.
func (c NewAdminContentVersionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.AdminContentVersions, error) {
	queries := mdbp.New(tx)
	return queries.CreateAdminContentVersion(ctx, mdbp.CreateAdminContentVersionParams{
		AdminContentVersionID: types.NewAdminContentVersionID(),
		AdminContentDataID:    c.params.AdminContentDataID,
		VersionNumber:         int32(c.params.VersionNumber),
		Locale:                c.params.Locale,
		Snapshot:              c.params.Snapshot,
		Trigger:               c.params.Trigger,
		Label:                 c.params.Label,
		Published:             c.params.Published,
		PublishedBy:           c.params.PublishedBy,
		DateCreated:           c.params.DateCreated,
	})
}

// NewAdminContentVersionCmd creates a new audited command for create.
func (d PsqlDatabase) NewAdminContentVersionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentVersionParams) NewAdminContentVersionCmdPsql {
	return NewAdminContentVersionCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

// DeleteAdminContentVersionCmdPsql is an audited command for delete on admin_content_versions (PostgreSQL).
type DeleteAdminContentVersionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentVersionID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context.
func (c DeleteAdminContentVersionCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteAdminContentVersionCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteAdminContentVersionCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteAdminContentVersionCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteAdminContentVersionCmdPsql) TableName() string { return "admin_content_versions" }

// GetID returns the record ID.
func (c DeleteAdminContentVersionCmdPsql) GetID() string { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteAdminContentVersionCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminContentVersions, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminContentVersion(ctx, mdbp.GetAdminContentVersionParams{AdminContentVersionID: c.id})
}

// Execute performs the delete operation.
func (c DeleteAdminContentVersionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteAdminContentVersion(ctx, mdbp.DeleteAdminContentVersionParams{AdminContentVersionID: c.id})
}

// DeleteAdminContentVersionCmd creates a new audited command for delete.
func (d PsqlDatabase) DeleteAdminContentVersionCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentVersionID) DeleteAdminContentVersionCmdPsql {
	return DeleteAdminContentVersionCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}

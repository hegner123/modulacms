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

// ContentVersion represents a versioned snapshot of content data.
type ContentVersion struct {
	ContentVersionID types.ContentVersionID `json:"content_version_id"`
	ContentDataID    types.ContentID        `json:"content_data_id"`
	VersionNumber    int64                  `json:"version_number"`
	Locale           string                 `json:"locale"`
	Snapshot         string                 `json:"snapshot"`
	Trigger          string                 `json:"trigger"`
	Label            string                 `json:"label"`
	Published        bool                   `json:"published"`
	PublishedBy      types.NullableUserID   `json:"published_by"`
	DateCreated      types.Timestamp        `json:"date_created"`
}

// CreateContentVersionParams specifies parameters for creating a content version.
type CreateContentVersionParams struct {
	ContentDataID types.ContentID      `json:"content_data_id"`
	VersionNumber int64                `json:"version_number"`
	Locale        string               `json:"locale"`
	Snapshot      string               `json:"snapshot"`
	Trigger       string               `json:"trigger"`
	Label         string               `json:"label"`
	Published     bool                 `json:"published"`
	PublishedBy   types.NullableUserID `json:"published_by"`
	DateCreated   types.Timestamp      `json:"date_created"`
}

// StringContentVersion is the string representation of ContentVersion for TUI table display.
type StringContentVersion struct {
	ContentVersionID string `json:"content_version_id"`
	ContentDataID    string `json:"content_data_id"`
	VersionNumber    string `json:"version_number"`
	Locale           string `json:"locale"`
	Snapshot         string `json:"snapshot"`
	Trigger          string `json:"trigger"`
	Label            string `json:"label"`
	Published        string `json:"published"`
	PublishedBy      string `json:"published_by"`
	DateCreated      string `json:"date_created"`
}

// MapStringContentVersion converts ContentVersion to StringContentVersion for table display.
func MapStringContentVersion(a ContentVersion) StringContentVersion {
	return StringContentVersion{
		ContentVersionID: a.ContentVersionID.String(),
		ContentDataID:    a.ContentDataID.String(),
		VersionNumber:    fmt.Sprintf("%d", a.VersionNumber),
		Locale:           a.Locale,
		Snapshot:         a.Snapshot,
		Trigger:          a.Trigger,
		Label:            a.Label,
		Published:        fmt.Sprintf("%t", a.Published),
		PublishedBy:      a.PublishedBy.String(),
		DateCreated:      a.DateCreated.String(),
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

// MapContentVersion converts a sqlc-generated SQLite type to the wrapper type.
func (d Database) MapContentVersion(a mdb.ContentVersions) ContentVersion {
	return ContentVersion{
		ContentVersionID: a.ContentVersionID,
		ContentDataID:    a.ContentDataID,
		VersionNumber:    a.VersionNumber,
		Locale:           a.Locale,
		Snapshot:         a.Snapshot,
		Trigger:          a.Trigger,
		Label:            a.Label,
		Published:        a.Published != 0,
		PublishedBy:      a.PublishedBy,
		DateCreated:      a.DateCreated,
	}
}

// MapCreateContentVersionParams converts wrapper params to a sqlc-generated SQLite type.
func (d Database) MapCreateContentVersionParams(a CreateContentVersionParams) mdb.CreateContentVersionParams {
	return mdb.CreateContentVersionParams{
		ContentVersionID: types.NewContentVersionID(),
		ContentDataID:    a.ContentDataID,
		VersionNumber:    a.VersionNumber,
		Locale:           a.Locale,
		Snapshot:         a.Snapshot,
		Trigger:          a.Trigger,
		Label:            a.Label,
		Published:        boolToInt64(a.Published),
		PublishedBy:      a.PublishedBy,
		DateCreated:      a.DateCreated,
	}
}

// QUERIES

// CountContentVersions returns the total count of content versions.
func (d Database) CountContentVersions() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountContentVersions(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count content versions: %w", err)
	}
	return &c, nil
}

// CountContentVersionsByContent returns the count of content versions for a given content data ID.
func (d Database) CountContentVersionsByContent(contentDataID types.ContentID) (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountContentVersionsByContent(d.Context, mdb.CountContentVersionsByContentParams{
		ContentDataID: contentDataID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count content versions by content: %w", err)
	}
	return &c, nil
}

// CreateContentVersionTable creates the content_versions table.
func (d Database) CreateContentVersionTable() error {
	queries := mdb.New(d.Connection)
	return queries.CreateContentVersionTable(d.Context)
}

// DropContentVersionTable drops the content_versions table.
func (d Database) DropContentVersionTable() error {
	queries := mdb.New(d.Connection)
	return queries.DropContentVersionTable(d.Context)
}

// CreateContentVersion creates a new content version with audit trail.
func (d Database) CreateContentVersion(ctx context.Context, ac audited.AuditContext, s CreateContentVersionParams) (*ContentVersion, error) {
	cmd := d.NewContentVersionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create content version: %w", err)
	}
	r := d.MapContentVersion(result)
	return &r, nil
}

// DeleteContentVersion deletes a content version with audit trail.
func (d Database) DeleteContentVersion(ctx context.Context, ac audited.AuditContext, id types.ContentVersionID) error {
	cmd := d.DeleteContentVersionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetContentVersion retrieves a content version by ID.
func (d Database) GetContentVersion(id types.ContentVersionID) (*ContentVersion, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetContentVersion(d.Context, mdb.GetContentVersionParams{ContentVersionID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get content version: %w", err)
	}
	res := d.MapContentVersion(row)
	return &res, nil
}

// GetPublishedSnapshot retrieves the published content version for a content data ID and locale.
func (d Database) GetPublishedSnapshot(contentDataID types.ContentID, locale string) (*ContentVersion, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetPublishedSnapshot(d.Context, mdb.GetPublishedSnapshotParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get published snapshot: %w", err)
	}
	res := d.MapContentVersion(row)
	return &res, nil
}

// ListContentVersionsByContent retrieves all content versions for a given content data ID.
func (d Database) ListContentVersionsByContent(contentDataID types.ContentID) (*[]ContentVersion, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentVersionsByContent(d.Context, mdb.ListContentVersionsByContentParams{
		ContentDataID: contentDataID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list content versions by content: %w", err)
	}
	res := []ContentVersion{}
	for _, v := range rows {
		res = append(res, d.MapContentVersion(v))
	}
	return &res, nil
}

// ListContentVersionsByContentLocale retrieves content versions for a given content data ID and locale.
func (d Database) ListContentVersionsByContentLocale(contentDataID types.ContentID, locale string) (*[]ContentVersion, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentVersionsByContentLocale(d.Context, mdb.ListContentVersionsByContentLocaleParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list content versions by content locale: %w", err)
	}
	res := []ContentVersion{}
	for _, v := range rows {
		res = append(res, d.MapContentVersion(v))
	}
	return &res, nil
}

// ClearPublishedFlag clears the published flag for all versions of a content data ID and locale.
func (d Database) ClearPublishedFlag(contentDataID types.ContentID, locale string) error {
	queries := mdb.New(d.Connection)
	return queries.ClearPublishedFlag(d.Context, mdb.ClearPublishedFlagParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
}

// GetMaxVersionNumber returns the highest version number for a content data ID and locale.
func (d Database) GetMaxVersionNumber(contentDataID types.ContentID, locale string) (int64, error) {
	queries := mdb.New(d.Connection)
	result, err := queries.GetMaxVersionNumber(d.Context, mdb.GetMaxVersionNumberParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get max version number: %w", err)
	}
	// COALESCE always returns a value, assert to int64
	return result.(int64), nil
}

// PruneOldVersions removes the oldest unlabeled, unpublished versions for a content data ID and locale.
func (d Database) PruneOldVersions(contentDataID types.ContentID, locale string, limit int64) error {
	queries := mdb.New(d.Connection)
	return queries.PruneOldVersions(d.Context, mdb.PruneOldVersionsParams{
		ContentDataID: contentDataID,
		Locale:        locale,
		Limit:         limit,
	})
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

// MapContentVersion converts a sqlc-generated MySQL type to the wrapper type.
func (d MysqlDatabase) MapContentVersion(a mdbm.ContentVersions) ContentVersion {
	return ContentVersion{
		ContentVersionID: a.ContentVersionID,
		ContentDataID:    a.ContentDataID,
		VersionNumber:    int64(a.VersionNumber),
		Locale:           a.Locale,
		Snapshot:         a.Snapshot,
		Trigger:          a.Trigger,
		Label:            a.Label,
		Published:        a.Published != 0,
		PublishedBy:      a.PublishedBy,
		DateCreated:      a.DateCreated,
	}
}

// MapCreateContentVersionParams converts wrapper params to a sqlc-generated MySQL type.
func (d MysqlDatabase) MapCreateContentVersionParams(a CreateContentVersionParams) mdbm.CreateContentVersionParams {
	return mdbm.CreateContentVersionParams{
		ContentVersionID: types.NewContentVersionID(),
		ContentDataID:    a.ContentDataID,
		VersionNumber:    int32(a.VersionNumber),
		Locale:           a.Locale,
		Snapshot:         a.Snapshot,
		Trigger:          a.Trigger,
		Label:            a.Label,
		Published:        boolToInt8(a.Published),
		PublishedBy:      a.PublishedBy,
		DateCreated:      a.DateCreated,
	}
}

// QUERIES

// CountContentVersions returns the total count of content versions.
func (d MysqlDatabase) CountContentVersions() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountContentVersions(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count content versions: %w", err)
	}
	return &c, nil
}

// CountContentVersionsByContent returns the count of content versions for a given content data ID.
func (d MysqlDatabase) CountContentVersionsByContent(contentDataID types.ContentID) (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountContentVersionsByContent(d.Context, mdbm.CountContentVersionsByContentParams{
		ContentDataID: contentDataID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count content versions by content: %w", err)
	}
	return &c, nil
}

// CreateContentVersionTable creates the content_versions table.
func (d MysqlDatabase) CreateContentVersionTable() error {
	queries := mdbm.New(d.Connection)
	return queries.CreateContentVersionTable(d.Context)
}

// DropContentVersionTable drops the content_versions table.
func (d MysqlDatabase) DropContentVersionTable() error {
	queries := mdbm.New(d.Connection)
	return queries.DropContentVersionTable(d.Context)
}

// CreateContentVersion creates a new content version with audit trail.
func (d MysqlDatabase) CreateContentVersion(ctx context.Context, ac audited.AuditContext, s CreateContentVersionParams) (*ContentVersion, error) {
	cmd := d.NewContentVersionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create content version: %w", err)
	}
	r := d.MapContentVersion(result)
	return &r, nil
}

// DeleteContentVersion deletes a content version with audit trail.
func (d MysqlDatabase) DeleteContentVersion(ctx context.Context, ac audited.AuditContext, id types.ContentVersionID) error {
	cmd := d.DeleteContentVersionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetContentVersion retrieves a content version by ID.
func (d MysqlDatabase) GetContentVersion(id types.ContentVersionID) (*ContentVersion, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetContentVersion(d.Context, mdbm.GetContentVersionParams{ContentVersionID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get content version: %w", err)
	}
	res := d.MapContentVersion(row)
	return &res, nil
}

// GetPublishedSnapshot retrieves the published content version for a content data ID and locale.
func (d MysqlDatabase) GetPublishedSnapshot(contentDataID types.ContentID, locale string) (*ContentVersion, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetPublishedSnapshot(d.Context, mdbm.GetPublishedSnapshotParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get published snapshot: %w", err)
	}
	res := d.MapContentVersion(row)
	return &res, nil
}

// ListContentVersionsByContent retrieves all content versions for a given content data ID.
func (d MysqlDatabase) ListContentVersionsByContent(contentDataID types.ContentID) (*[]ContentVersion, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentVersionsByContent(d.Context, mdbm.ListContentVersionsByContentParams{
		ContentDataID: contentDataID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list content versions by content: %w", err)
	}
	res := []ContentVersion{}
	for _, v := range rows {
		res = append(res, d.MapContentVersion(v))
	}
	return &res, nil
}

// ListContentVersionsByContentLocale retrieves content versions for a given content data ID and locale.
func (d MysqlDatabase) ListContentVersionsByContentLocale(contentDataID types.ContentID, locale string) (*[]ContentVersion, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentVersionsByContentLocale(d.Context, mdbm.ListContentVersionsByContentLocaleParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list content versions by content locale: %w", err)
	}
	res := []ContentVersion{}
	for _, v := range rows {
		res = append(res, d.MapContentVersion(v))
	}
	return &res, nil
}

// ClearPublishedFlag clears the published flag for all versions of a content data ID and locale.
func (d MysqlDatabase) ClearPublishedFlag(contentDataID types.ContentID, locale string) error {
	queries := mdbm.New(d.Connection)
	return queries.ClearPublishedFlag(d.Context, mdbm.ClearPublishedFlagParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
}

// GetMaxVersionNumber returns the highest version number for a content data ID and locale.
func (d MysqlDatabase) GetMaxVersionNumber(contentDataID types.ContentID, locale string) (int64, error) {
	queries := mdbm.New(d.Connection)
	result, err := queries.GetMaxVersionNumber(d.Context, mdbm.GetMaxVersionNumberParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get max version number: %w", err)
	}
	// COALESCE always returns a value, assert to int64
	return result.(int64), nil
}

// PruneOldVersions removes the oldest unlabeled, unpublished versions for a content data ID and locale.
func (d MysqlDatabase) PruneOldVersions(contentDataID types.ContentID, locale string, limit int64) error {
	queries := mdbm.New(d.Connection)
	return queries.PruneOldVersions(d.Context, mdbm.PruneOldVersionsParams{
		ContentDataID: contentDataID,
		Locale:        locale,
		Limit:         int32(limit),
	})
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

// MapContentVersion converts a sqlc-generated PostgreSQL type to the wrapper type.
func (d PsqlDatabase) MapContentVersion(a mdbp.ContentVersions) ContentVersion {
	return ContentVersion{
		ContentVersionID: a.ContentVersionID,
		ContentDataID:    a.ContentDataID,
		VersionNumber:    int64(a.VersionNumber),
		Locale:           a.Locale,
		Snapshot:         a.Snapshot,
		Trigger:          a.Trigger,
		Label:            a.Label,
		Published:        a.Published,
		PublishedBy:      a.PublishedBy,
		DateCreated:      a.DateCreated,
	}
}

// MapCreateContentVersionParams converts wrapper params to a sqlc-generated PostgreSQL type.
func (d PsqlDatabase) MapCreateContentVersionParams(a CreateContentVersionParams) mdbp.CreateContentVersionParams {
	return mdbp.CreateContentVersionParams{
		ContentVersionID: types.NewContentVersionID(),
		ContentDataID:    a.ContentDataID,
		VersionNumber:    int32(a.VersionNumber),
		Locale:           a.Locale,
		Snapshot:         a.Snapshot,
		Trigger:          a.Trigger,
		Label:            a.Label,
		Published:        a.Published,
		PublishedBy:      a.PublishedBy,
		DateCreated:      a.DateCreated,
	}
}

// QUERIES

// CountContentVersions returns the total count of content versions.
func (d PsqlDatabase) CountContentVersions() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountContentVersions(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count content versions: %w", err)
	}
	return &c, nil
}

// CountContentVersionsByContent returns the count of content versions for a given content data ID.
func (d PsqlDatabase) CountContentVersionsByContent(contentDataID types.ContentID) (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountContentVersionsByContent(d.Context, mdbp.CountContentVersionsByContentParams{
		ContentDataID: contentDataID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count content versions by content: %w", err)
	}
	return &c, nil
}

// CreateContentVersionTable creates the content_versions table.
func (d PsqlDatabase) CreateContentVersionTable() error {
	queries := mdbp.New(d.Connection)
	return queries.CreateContentVersionTable(d.Context)
}

// DropContentVersionTable drops the content_versions table.
func (d PsqlDatabase) DropContentVersionTable() error {
	queries := mdbp.New(d.Connection)
	return queries.DropContentVersionTable(d.Context)
}

// CreateContentVersion creates a new content version with audit trail.
func (d PsqlDatabase) CreateContentVersion(ctx context.Context, ac audited.AuditContext, s CreateContentVersionParams) (*ContentVersion, error) {
	cmd := d.NewContentVersionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create content version: %w", err)
	}
	r := d.MapContentVersion(result)
	return &r, nil
}

// DeleteContentVersion deletes a content version with audit trail.
func (d PsqlDatabase) DeleteContentVersion(ctx context.Context, ac audited.AuditContext, id types.ContentVersionID) error {
	cmd := d.DeleteContentVersionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetContentVersion retrieves a content version by ID.
func (d PsqlDatabase) GetContentVersion(id types.ContentVersionID) (*ContentVersion, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetContentVersion(d.Context, mdbp.GetContentVersionParams{ContentVersionID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get content version: %w", err)
	}
	res := d.MapContentVersion(row)
	return &res, nil
}

// GetPublishedSnapshot retrieves the published content version for a content data ID and locale.
func (d PsqlDatabase) GetPublishedSnapshot(contentDataID types.ContentID, locale string) (*ContentVersion, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetPublishedSnapshot(d.Context, mdbp.GetPublishedSnapshotParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get published snapshot: %w", err)
	}
	res := d.MapContentVersion(row)
	return &res, nil
}

// ListContentVersionsByContent retrieves all content versions for a given content data ID.
func (d PsqlDatabase) ListContentVersionsByContent(contentDataID types.ContentID) (*[]ContentVersion, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentVersionsByContent(d.Context, mdbp.ListContentVersionsByContentParams{
		ContentDataID: contentDataID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list content versions by content: %w", err)
	}
	res := []ContentVersion{}
	for _, v := range rows {
		res = append(res, d.MapContentVersion(v))
	}
	return &res, nil
}

// ListContentVersionsByContentLocale retrieves content versions for a given content data ID and locale.
func (d PsqlDatabase) ListContentVersionsByContentLocale(contentDataID types.ContentID, locale string) (*[]ContentVersion, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentVersionsByContentLocale(d.Context, mdbp.ListContentVersionsByContentLocaleParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list content versions by content locale: %w", err)
	}
	res := []ContentVersion{}
	for _, v := range rows {
		res = append(res, d.MapContentVersion(v))
	}
	return &res, nil
}

// ClearPublishedFlag clears the published flag for all versions of a content data ID and locale.
func (d PsqlDatabase) ClearPublishedFlag(contentDataID types.ContentID, locale string) error {
	queries := mdbp.New(d.Connection)
	return queries.ClearPublishedFlag(d.Context, mdbp.ClearPublishedFlagParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
}

// GetMaxVersionNumber returns the highest version number for a content data ID and locale.
func (d PsqlDatabase) GetMaxVersionNumber(contentDataID types.ContentID, locale string) (int64, error) {
	queries := mdbp.New(d.Connection)
	result, err := queries.GetMaxVersionNumber(d.Context, mdbp.GetMaxVersionNumberParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get max version number: %w", err)
	}
	// COALESCE always returns a value, assert to int64
	return result.(int64), nil
}

// PruneOldVersions removes the oldest unlabeled, unpublished versions for a content data ID and locale.
func (d PsqlDatabase) PruneOldVersions(contentDataID types.ContentID, locale string, limit int64) error {
	queries := mdbp.New(d.Connection)
	return queries.PruneOldVersions(d.Context, mdbp.PruneOldVersionsParams{
		ContentDataID: contentDataID,
		Locale:        locale,
		Limit:         int32(limit),
	})
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ----- SQLite CREATE -----

// NewContentVersionCmd is an audited command for creating a content version.
type NewContentVersionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateContentVersionParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c NewContentVersionCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewContentVersionCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewContentVersionCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewContentVersionCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewContentVersionCmd) TableName() string { return "content_versions" }

// Params returns the command parameters.
func (c NewContentVersionCmd) Params() any { return c.params }

// GetID returns the ID from a content version.
func (c NewContentVersionCmd) GetID(r mdb.ContentVersions) string {
	return string(r.ContentVersionID)
}

// Execute creates the content version in the database.
func (c NewContentVersionCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.ContentVersions, error) {
	queries := mdb.New(tx)
	return queries.CreateContentVersion(ctx, mdb.CreateContentVersionParams{
		ContentVersionID: types.NewContentVersionID(),
		ContentDataID:    c.params.ContentDataID,
		VersionNumber:    c.params.VersionNumber,
		Locale:           c.params.Locale,
		Snapshot:         c.params.Snapshot,
		Trigger:          c.params.Trigger,
		Label:            c.params.Label,
		Published:        boolToInt64(c.params.Published),
		PublishedBy:      c.params.PublishedBy,
		DateCreated:      c.params.DateCreated,
	})
}

// NewContentVersionCmd creates a new create command for a content version.
func (d Database) NewContentVersionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateContentVersionParams) NewContentVersionCmd {
	return NewContentVersionCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

// DeleteContentVersionCmd is an audited command for deleting a content version.
type DeleteContentVersionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.ContentVersionID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c DeleteContentVersionCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteContentVersionCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteContentVersionCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteContentVersionCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteContentVersionCmd) TableName() string { return "content_versions" }

// GetID returns the content version ID.
func (c DeleteContentVersionCmd) GetID() string { return string(c.id) }

// GetBefore retrieves the content version before deletion.
func (c DeleteContentVersionCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.ContentVersions, error) {
	queries := mdb.New(tx)
	return queries.GetContentVersion(ctx, mdb.GetContentVersionParams{ContentVersionID: c.id})
}

// Execute deletes the content version from the database.
func (c DeleteContentVersionCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteContentVersion(ctx, mdb.DeleteContentVersionParams{ContentVersionID: c.id})
}

// DeleteContentVersionCmd creates a new delete command for a content version.
func (d Database) DeleteContentVersionCmd(ctx context.Context, auditCtx audited.AuditContext, id types.ContentVersionID) DeleteContentVersionCmd {
	return DeleteContentVersionCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

// NewContentVersionCmdMysql is an audited command for creating a content version on MySQL.
type NewContentVersionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateContentVersionParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c NewContentVersionCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewContentVersionCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewContentVersionCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewContentVersionCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewContentVersionCmdMysql) TableName() string { return "content_versions" }

// Params returns the command parameters.
func (c NewContentVersionCmdMysql) Params() any { return c.params }

// GetID returns the ID from a content version.
func (c NewContentVersionCmdMysql) GetID(r mdbm.ContentVersions) string {
	return string(r.ContentVersionID)
}

// Execute creates the content version in the database.
func (c NewContentVersionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.ContentVersions, error) {
	id := types.NewContentVersionID()
	queries := mdbm.New(tx)
	params := mdbm.CreateContentVersionParams{
		ContentVersionID: id,
		ContentDataID:    c.params.ContentDataID,
		VersionNumber:    int32(c.params.VersionNumber),
		Locale:           c.params.Locale,
		Snapshot:         c.params.Snapshot,
		Trigger:          c.params.Trigger,
		Label:            c.params.Label,
		Published:        boolToInt8(c.params.Published),
		PublishedBy:      c.params.PublishedBy,
		DateCreated:      c.params.DateCreated,
	}
	if err := queries.CreateContentVersion(ctx, params); err != nil {
		return mdbm.ContentVersions{}, err
	}
	return queries.GetContentVersion(ctx, mdbm.GetContentVersionParams{ContentVersionID: id})
}

// NewContentVersionCmd creates a new create command for a content version.
func (d MysqlDatabase) NewContentVersionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateContentVersionParams) NewContentVersionCmdMysql {
	return NewContentVersionCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

// DeleteContentVersionCmdMysql is an audited command for deleting a content version on MySQL.
type DeleteContentVersionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.ContentVersionID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c DeleteContentVersionCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteContentVersionCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteContentVersionCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteContentVersionCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteContentVersionCmdMysql) TableName() string { return "content_versions" }

// GetID returns the content version ID.
func (c DeleteContentVersionCmdMysql) GetID() string { return string(c.id) }

// GetBefore retrieves the content version before deletion.
func (c DeleteContentVersionCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.ContentVersions, error) {
	queries := mdbm.New(tx)
	return queries.GetContentVersion(ctx, mdbm.GetContentVersionParams{ContentVersionID: c.id})
}

// Execute deletes the content version from the database.
func (c DeleteContentVersionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteContentVersion(ctx, mdbm.DeleteContentVersionParams{ContentVersionID: c.id})
}

// DeleteContentVersionCmd creates a new delete command for a content version.
func (d MysqlDatabase) DeleteContentVersionCmd(ctx context.Context, auditCtx audited.AuditContext, id types.ContentVersionID) DeleteContentVersionCmdMysql {
	return DeleteContentVersionCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

// NewContentVersionCmdPsql is an audited command for creating a content version on PostgreSQL.
type NewContentVersionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateContentVersionParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c NewContentVersionCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewContentVersionCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewContentVersionCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewContentVersionCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewContentVersionCmdPsql) TableName() string { return "content_versions" }

// Params returns the command parameters.
func (c NewContentVersionCmdPsql) Params() any { return c.params }

// GetID returns the ID from a content version.
func (c NewContentVersionCmdPsql) GetID(r mdbp.ContentVersions) string {
	return string(r.ContentVersionID)
}

// Execute creates the content version in the database.
func (c NewContentVersionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.ContentVersions, error) {
	queries := mdbp.New(tx)
	return queries.CreateContentVersion(ctx, mdbp.CreateContentVersionParams{
		ContentVersionID: types.NewContentVersionID(),
		ContentDataID:    c.params.ContentDataID,
		VersionNumber:    int32(c.params.VersionNumber),
		Locale:           c.params.Locale,
		Snapshot:         c.params.Snapshot,
		Trigger:          c.params.Trigger,
		Label:            c.params.Label,
		Published:        c.params.Published,
		PublishedBy:      c.params.PublishedBy,
		DateCreated:      c.params.DateCreated,
	})
}

// NewContentVersionCmd creates a new create command for a content version.
func (d PsqlDatabase) NewContentVersionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateContentVersionParams) NewContentVersionCmdPsql {
	return NewContentVersionCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

// DeleteContentVersionCmdPsql is an audited command for deleting a content version on PostgreSQL.
type DeleteContentVersionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.ContentVersionID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c DeleteContentVersionCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteContentVersionCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteContentVersionCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteContentVersionCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteContentVersionCmdPsql) TableName() string { return "content_versions" }

// GetID returns the content version ID.
func (c DeleteContentVersionCmdPsql) GetID() string { return string(c.id) }

// GetBefore retrieves the content version before deletion.
func (c DeleteContentVersionCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.ContentVersions, error) {
	queries := mdbp.New(tx)
	return queries.GetContentVersion(ctx, mdbp.GetContentVersionParams{ContentVersionID: c.id})
}

// Execute deletes the content version from the database.
func (c DeleteContentVersionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteContentVersion(ctx, mdbp.DeleteContentVersionParams{ContentVersionID: c.id})
}

// DeleteContentVersionCmd creates a new delete command for a content version.
func (d PsqlDatabase) DeleteContentVersionCmd(ctx context.Context, auditCtx audited.AuditContext, id types.ContentVersionID) DeleteContentVersionCmdPsql {
	return DeleteContentVersionCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}

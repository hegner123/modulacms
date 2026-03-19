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

// ListMediaByFolderPaginatedParams contains parameters for paginated media listing by folder.
type ListMediaByFolderPaginatedParams struct {
	FolderID types.NullableMediaFolderID `json:"folder_id"`
	Limit    int64                       `json:"limit"`
	Offset   int64                       `json:"offset"`
}

// MoveMediaToFolderParams contains parameters for moving a media item to a folder.
type MoveMediaToFolderParams struct {
	FolderID     types.NullableMediaFolderID `json:"folder_id"`
	DateModified types.Timestamp             `json:"date_modified"`
	MediaID      types.MediaID               `json:"media_id"`
}

///////////////////////////////
// SQLITE
//////////////////////////////

// ListMediaByFolder retrieves all media in a given folder (SQLite).
func (d Database) ListMediaByFolder(folderID types.NullableMediaFolderID) (*[]Media, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListMediaByFolder(d.Context, mdb.ListMediaByFolderParams{FolderID: folderID})
	if err != nil {
		return nil, fmt.Errorf("failed to list media by folder: %w", err)
	}
	res := make([]Media, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapMedia(v))
	}
	return &res, nil
}

// ListMediaByFolderPaginated retrieves a paginated list of media in a given folder (SQLite).
func (d Database) ListMediaByFolderPaginated(params ListMediaByFolderPaginatedParams) (*[]Media, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListMediaByFolderPaginated(d.Context, mdb.ListMediaByFolderPaginatedParams{
		FolderID: params.FolderID,
		Limit:    params.Limit,
		Offset:   params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list media by folder paginated: %w", err)
	}
	res := make([]Media, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapMedia(v))
	}
	return &res, nil
}

// ListMediaUnfiled retrieves all media with no folder assignment (SQLite).
func (d Database) ListMediaUnfiled() (*[]Media, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListMediaUnfiled(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list unfiled media: %w", err)
	}
	res := make([]Media, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapMedia(v))
	}
	return &res, nil
}

// ListMediaUnfiledPaginated retrieves a paginated list of media with no folder assignment (SQLite).
func (d Database) ListMediaUnfiledPaginated(params PaginationParams) (*[]Media, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListMediaUnfiledPaginated(d.Context, mdb.ListMediaUnfiledPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list unfiled media paginated: %w", err)
	}
	res := make([]Media, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapMedia(v))
	}
	return &res, nil
}

// CountMediaByFolder returns the number of media in a given folder (SQLite).
func (d Database) CountMediaByFolder(folderID types.NullableMediaFolderID) (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountMediaByFolder(d.Context, mdb.CountMediaByFolderParams{FolderID: folderID})
	if err != nil {
		return nil, fmt.Errorf("failed to count media by folder: %w", err)
	}
	return &c, nil
}

// CountMediaUnfiled returns the number of media with no folder assignment (SQLite).
func (d Database) CountMediaUnfiled() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountMediaUnfiled(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count unfiled media: %w", err)
	}
	return &c, nil
}

// MoveMediaToFolder moves a media item to a folder and records an audit event (SQLite).
func (d Database) MoveMediaToFolder(ctx context.Context, ac audited.AuditContext, params MoveMediaToFolderParams) error {
	cmd := d.MoveMediaToFolderCmd(ctx, ac, params)
	return audited.Update(cmd)
}

// ----- SQLite MoveMediaToFolder UPDATE command -----

// MoveMediaToFolderCmd is an audited command for moving media to a folder.
type MoveMediaToFolderCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   MoveMediaToFolderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c MoveMediaToFolderCmd) Context() context.Context              { return c.ctx }
func (c MoveMediaToFolderCmd) AuditContext() audited.AuditContext    { return c.auditCtx }
func (c MoveMediaToFolderCmd) Connection() *sql.DB                   { return c.conn }
func (c MoveMediaToFolderCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c MoveMediaToFolderCmd) TableName() string                     { return "media" }
func (c MoveMediaToFolderCmd) GetID() string                         { return string(c.params.MediaID) }
func (c MoveMediaToFolderCmd) Params() any                           { return c.params }

func (c MoveMediaToFolderCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Media, error) {
	queries := mdb.New(tx)
	return queries.GetMedia(ctx, mdb.GetMediaParams{MediaID: c.params.MediaID})
}

func (c MoveMediaToFolderCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.MoveMediaToFolder(ctx, mdb.MoveMediaToFolderParams{
		FolderID:     c.params.FolderID,
		DateModified: c.params.DateModified,
		MediaID:      c.params.MediaID,
	})
}

func (d Database) MoveMediaToFolderCmd(ctx context.Context, auditCtx audited.AuditContext, params MoveMediaToFolderParams) MoveMediaToFolderCmd {
	return MoveMediaToFolderCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL MoveMediaToFolder UPDATE command -----

// ----- PostgreSQL MoveMediaToFolder UPDATE command -----

// ===== MySQL =====

// ===== PostgreSQL =====

// MYSQL

// ListMediaByFolder retrieves all media in a given folder (MySQL).
func (d MysqlDatabase) ListMediaByFolder(folderID types.NullableMediaFolderID) (*[]Media, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListMediaByFolder(d.Context, mdbm.ListMediaByFolderParams{FolderID: folderID})
	if err != nil {
		return nil, fmt.Errorf("failed to list media by folder: %w", err)
	}
	res := make([]Media, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapMedia(v))
	}
	return &res, nil
}

// ListMediaUnfiled retrieves all media with no folder assignment (MySQL).
func (d MysqlDatabase) ListMediaUnfiled() (*[]Media, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListMediaUnfiled(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list unfiled media: %w", err)
	}
	res := make([]Media, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapMedia(v))
	}
	return &res, nil
}

// CountMediaByFolder returns the number of media in a given folder (MySQL).
func (d MysqlDatabase) CountMediaByFolder(folderID types.NullableMediaFolderID) (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountMediaByFolder(d.Context, mdbm.CountMediaByFolderParams{FolderID: folderID})
	if err != nil {
		return nil, fmt.Errorf("failed to count media by folder: %w", err)
	}
	return &c, nil
}

// CountMediaUnfiled returns the number of media with no folder assignment (MySQL).
func (d MysqlDatabase) CountMediaUnfiled() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountMediaUnfiled(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count unfiled media: %w", err)
	}
	return &c, nil
}

type MoveMediaToFolderCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   MoveMediaToFolderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c MoveMediaToFolderCmdMysql) Context() context.Context              { return c.ctx }
func (c MoveMediaToFolderCmdMysql) AuditContext() audited.AuditContext    { return c.auditCtx }
func (c MoveMediaToFolderCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c MoveMediaToFolderCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c MoveMediaToFolderCmdMysql) TableName() string                     { return "media" }
func (c MoveMediaToFolderCmdMysql) GetID() string                         { return string(c.params.MediaID) }
func (c MoveMediaToFolderCmdMysql) Params() any                           { return c.params }
func (c MoveMediaToFolderCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Media, error) {
	queries := mdbm.New(tx)
	return queries.GetMedia(ctx, mdbm.GetMediaParams{MediaID: c.params.MediaID})
}
func (c MoveMediaToFolderCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.MoveMediaToFolder(ctx, mdbm.MoveMediaToFolderParams{
		FolderID:     c.params.FolderID,
		DateModified: c.params.DateModified,
		MediaID:      c.params.MediaID,
	})
}
func (d MysqlDatabase) MoveMediaToFolderCmd(ctx context.Context, auditCtx audited.AuditContext, params MoveMediaToFolderParams) MoveMediaToFolderCmdMysql {
	return MoveMediaToFolderCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ListMediaByFolderPaginated retrieves a paginated list of media in a given folder (MySQL).
func (d MysqlDatabase) ListMediaByFolderPaginated(params ListMediaByFolderPaginatedParams) (*[]Media, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListMediaByFolderPaginated(d.Context, mdbm.ListMediaByFolderPaginatedParams{
		FolderID: params.FolderID,
		Limit:    int32(params.Limit),
		Offset:   int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list media by folder paginated: %w", err)
	}
	res := make([]Media, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapMedia(v))
	}
	return &res, nil
}

// ListMediaUnfiledPaginated retrieves a paginated list of media with no folder assignment (MySQL).
func (d MysqlDatabase) ListMediaUnfiledPaginated(params PaginationParams) (*[]Media, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListMediaUnfiledPaginated(d.Context, mdbm.ListMediaUnfiledPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list unfiled media paginated: %w", err)
	}
	res := make([]Media, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapMedia(v))
	}
	return &res, nil
}

// MoveMediaToFolder moves a media item to a folder and records an audit event (MySQL).
func (d MysqlDatabase) MoveMediaToFolder(ctx context.Context, ac audited.AuditContext, params MoveMediaToFolderParams) error {
	cmd := d.MoveMediaToFolderCmd(ctx, ac, params)
	return audited.Update(cmd)
}

// PSQL

// ListMediaByFolder retrieves all media in a given folder (PostgreSQL).
func (d PsqlDatabase) ListMediaByFolder(folderID types.NullableMediaFolderID) (*[]Media, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListMediaByFolder(d.Context, mdbp.ListMediaByFolderParams{FolderID: folderID})
	if err != nil {
		return nil, fmt.Errorf("failed to list media by folder: %w", err)
	}
	res := make([]Media, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapMedia(v))
	}
	return &res, nil
}

// ListMediaUnfiled retrieves all media with no folder assignment (PostgreSQL).
func (d PsqlDatabase) ListMediaUnfiled() (*[]Media, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListMediaUnfiled(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list unfiled media: %w", err)
	}
	res := make([]Media, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapMedia(v))
	}
	return &res, nil
}

// CountMediaByFolder returns the number of media in a given folder (PostgreSQL).
func (d PsqlDatabase) CountMediaByFolder(folderID types.NullableMediaFolderID) (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountMediaByFolder(d.Context, mdbp.CountMediaByFolderParams{FolderID: folderID})
	if err != nil {
		return nil, fmt.Errorf("failed to count media by folder: %w", err)
	}
	return &c, nil
}

// CountMediaUnfiled returns the number of media with no folder assignment (PostgreSQL).
func (d PsqlDatabase) CountMediaUnfiled() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountMediaUnfiled(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count unfiled media: %w", err)
	}
	return &c, nil
}

type MoveMediaToFolderCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   MoveMediaToFolderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c MoveMediaToFolderCmdPsql) Context() context.Context              { return c.ctx }
func (c MoveMediaToFolderCmdPsql) AuditContext() audited.AuditContext    { return c.auditCtx }
func (c MoveMediaToFolderCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c MoveMediaToFolderCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c MoveMediaToFolderCmdPsql) TableName() string                     { return "media" }
func (c MoveMediaToFolderCmdPsql) GetID() string                         { return string(c.params.MediaID) }
func (c MoveMediaToFolderCmdPsql) Params() any                           { return c.params }
func (c MoveMediaToFolderCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Media, error) {
	queries := mdbp.New(tx)
	return queries.GetMedia(ctx, mdbp.GetMediaParams{MediaID: c.params.MediaID})
}
func (c MoveMediaToFolderCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.MoveMediaToFolder(ctx, mdbp.MoveMediaToFolderParams{
		FolderID:     c.params.FolderID,
		DateModified: c.params.DateModified,
		MediaID:      c.params.MediaID,
	})
}
func (d PsqlDatabase) MoveMediaToFolderCmd(ctx context.Context, auditCtx audited.AuditContext, params MoveMediaToFolderParams) MoveMediaToFolderCmdPsql {
	return MoveMediaToFolderCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ListMediaByFolderPaginated retrieves a paginated list of media in a given folder (PostgreSQL).
func (d PsqlDatabase) ListMediaByFolderPaginated(params ListMediaByFolderPaginatedParams) (*[]Media, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListMediaByFolderPaginated(d.Context, mdbp.ListMediaByFolderPaginatedParams{
		FolderID: params.FolderID,
		Limit:    int32(params.Limit),
		Offset:   int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list media by folder paginated: %w", err)
	}
	res := make([]Media, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapMedia(v))
	}
	return &res, nil
}

// ListMediaUnfiledPaginated retrieves a paginated list of media with no folder assignment (PostgreSQL).
func (d PsqlDatabase) ListMediaUnfiledPaginated(params PaginationParams) (*[]Media, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListMediaUnfiledPaginated(d.Context, mdbp.ListMediaUnfiledPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list unfiled media paginated: %w", err)
	}
	res := make([]Media, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapMedia(v))
	}
	return &res, nil
}

// MoveMediaToFolder moves a media item to a folder and records an audit event (PostgreSQL).
func (d PsqlDatabase) MoveMediaToFolder(ctx context.Context, ac audited.AuditContext, params MoveMediaToFolderParams) error {
	cmd := d.MoveMediaToFolderCmd(ctx, ac, params)
	return audited.Update(cmd)
}

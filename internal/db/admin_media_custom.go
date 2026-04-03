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

// ListAdminMediaByFolderPaginatedParams contains parameters for paginated admin media listing by folder.
type ListAdminMediaByFolderPaginatedParams struct {
	FolderID types.NullableAdminMediaFolderID `json:"folder_id"`
	Limit    int64                            `json:"limit"`
	Offset   int64                            `json:"offset"`
}

// MoveAdminMediaToFolderParams contains parameters for moving an admin media item to a folder.
type MoveAdminMediaToFolderParams struct {
	FolderID     types.NullableAdminMediaFolderID `json:"folder_id"`
	DateModified types.Timestamp                  `json:"date_modified"`
	AdminMediaID types.AdminMediaID               `json:"admin_media_id"`
}

///////////////////////////////
// SQLITE
//////////////////////////////

// ListAdminMediaByFolder retrieves all admin media in a given folder (SQLite).
func (d Database) ListAdminMediaByFolder(folderID types.NullableAdminMediaFolderID) (*[]AdminMedia, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminMediaByFolder(d.Context, mdb.ListAdminMediaByFolderParams{FolderID: folderID})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin media by folder: %w", err)
	}
	res := make([]AdminMedia, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminMedia(v))
	}
	return &res, nil
}

// ListAdminMediaByFolderPaginated retrieves a paginated list of admin media in a given folder (SQLite).
func (d Database) ListAdminMediaByFolderPaginated(params ListAdminMediaByFolderPaginatedParams) (*[]AdminMedia, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminMediaByFolderPaginated(d.Context, mdb.ListAdminMediaByFolderPaginatedParams{
		FolderID: params.FolderID,
		Limit:    params.Limit,
		Offset:   params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin media by folder paginated: %w", err)
	}
	res := make([]AdminMedia, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminMedia(v))
	}
	return &res, nil
}

// ListAdminMediaUnfiled retrieves all admin media with no folder assignment (SQLite).
func (d Database) ListAdminMediaUnfiled() (*[]AdminMedia, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminMediaUnfiled(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list unfiled admin media: %w", err)
	}
	res := make([]AdminMedia, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminMedia(v))
	}
	return &res, nil
}

// ListAdminMediaUnfiledPaginated retrieves a paginated list of admin media with no folder assignment (SQLite).
func (d Database) ListAdminMediaUnfiledPaginated(params PaginationParams) (*[]AdminMedia, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminMediaUnfiledPaginated(d.Context, mdb.ListAdminMediaUnfiledPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list unfiled admin media paginated: %w", err)
	}
	res := make([]AdminMedia, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminMedia(v))
	}
	return &res, nil
}

// CountAdminMediaByFolder returns the number of admin media in a given folder (SQLite).
func (d Database) CountAdminMediaByFolder(folderID types.NullableAdminMediaFolderID) (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminMediaByFolder(d.Context, mdb.CountAdminMediaByFolderParams{FolderID: folderID})
	if err != nil {
		return nil, fmt.Errorf("failed to count admin media by folder: %w", err)
	}
	return &c, nil
}

// CountAdminMediaUnfiled returns the number of admin media with no folder assignment (SQLite).
func (d Database) CountAdminMediaUnfiled() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminMediaUnfiled(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count unfiled admin media: %w", err)
	}
	return &c, nil
}

// MoveAdminMediaToFolder moves an admin media item to a folder and records an audit event (SQLite).
func (d Database) MoveAdminMediaToFolder(ctx context.Context, ac audited.AuditContext, params MoveAdminMediaToFolderParams) error {
	cmd := d.MoveAdminMediaToFolderCmd(ctx, ac, params)
	return audited.Update(cmd)
}

// ----- SQLite MoveAdminMediaToFolder UPDATE command -----

// MoveAdminMediaToFolderCmd is an audited command for moving admin media to a folder.
type MoveAdminMediaToFolderCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   MoveAdminMediaToFolderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c MoveAdminMediaToFolderCmd) Context() context.Context              { return c.ctx }
func (c MoveAdminMediaToFolderCmd) AuditContext() audited.AuditContext    { return c.auditCtx }
func (c MoveAdminMediaToFolderCmd) Connection() *sql.DB                   { return c.conn }
func (c MoveAdminMediaToFolderCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c MoveAdminMediaToFolderCmd) TableName() string                     { return "admin_media" }
func (c MoveAdminMediaToFolderCmd) GetID() string {
	return string(c.params.AdminMediaID)
}
func (c MoveAdminMediaToFolderCmd) Params() any { return c.params }

func (c MoveAdminMediaToFolderCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminMedia, error) {
	queries := mdb.New(tx)
	return queries.GetAdminMedia(ctx, mdb.GetAdminMediaParams{AdminMediaID: c.params.AdminMediaID})
}

func (c MoveAdminMediaToFolderCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.MoveAdminMediaToFolder(ctx, mdb.MoveAdminMediaToFolderParams{
		FolderID:     c.params.FolderID,
		DateModified: c.params.DateModified,
		AdminMediaID: c.params.AdminMediaID,
	})
}

func (d Database) MoveAdminMediaToFolderCmd(ctx context.Context, auditCtx audited.AuditContext, params MoveAdminMediaToFolderParams) MoveAdminMediaToFolderCmd {
	return MoveAdminMediaToFolderCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL MoveAdminMediaToFolder UPDATE command -----

// ----- PostgreSQL MoveAdminMediaToFolder UPDATE command -----

// ===== MySQL =====

// ===== PostgreSQL =====

// MYSQL

// ListAdminMediaByFolder retrieves all admin media in a given folder (MySQL).
func (d MysqlDatabase) ListAdminMediaByFolder(folderID types.NullableAdminMediaFolderID) (*[]AdminMedia, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminMediaByFolder(d.Context, mdbm.ListAdminMediaByFolderParams{FolderID: folderID})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin media by folder: %w", err)
	}
	res := make([]AdminMedia, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminMedia(v))
	}
	return &res, nil
}

// ListAdminMediaUnfiled retrieves all admin media with no folder assignment (MySQL).
func (d MysqlDatabase) ListAdminMediaUnfiled() (*[]AdminMedia, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminMediaUnfiled(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list unfiled admin media: %w", err)
	}
	res := make([]AdminMedia, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminMedia(v))
	}
	return &res, nil
}

// CountAdminMediaByFolder returns the number of admin media in a given folder (MySQL).
func (d MysqlDatabase) CountAdminMediaByFolder(folderID types.NullableAdminMediaFolderID) (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminMediaByFolder(d.Context, mdbm.CountAdminMediaByFolderParams{FolderID: folderID})
	if err != nil {
		return nil, fmt.Errorf("failed to count admin media by folder: %w", err)
	}
	return &c, nil
}

// CountAdminMediaUnfiled returns the number of admin media with no folder assignment (MySQL).
func (d MysqlDatabase) CountAdminMediaUnfiled() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminMediaUnfiled(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count unfiled admin media: %w", err)
	}
	return &c, nil
}

func (d MysqlDatabase) MoveAdminMediaToFolderCmd(ctx context.Context, auditCtx audited.AuditContext, params MoveAdminMediaToFolderParams) MoveAdminMediaToFolderCmd {
	return MoveAdminMediaToFolderCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ListAdminMediaByFolderPaginated retrieves a paginated list of admin media in a given folder (MySQL).
func (d MysqlDatabase) ListAdminMediaByFolderPaginated(params ListAdminMediaByFolderPaginatedParams) (*[]AdminMedia, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminMediaByFolderPaginated(d.Context, mdbm.ListAdminMediaByFolderPaginatedParams{
		FolderID: params.FolderID,
		Limit:    int32(params.Limit),
		Offset:   int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin media by folder paginated: %w", err)
	}
	res := make([]AdminMedia, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminMedia(v))
	}
	return &res, nil
}

// ListAdminMediaUnfiledPaginated retrieves a paginated list of admin media with no folder assignment (MySQL).
func (d MysqlDatabase) ListAdminMediaUnfiledPaginated(params PaginationParams) (*[]AdminMedia, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminMediaUnfiledPaginated(d.Context, mdbm.ListAdminMediaUnfiledPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list unfiled admin media paginated: %w", err)
	}
	res := make([]AdminMedia, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminMedia(v))
	}
	return &res, nil
}

// MoveAdminMediaToFolder moves an admin media item to a folder and records an audit event (MySQL).
func (d MysqlDatabase) MoveAdminMediaToFolder(ctx context.Context, ac audited.AuditContext, params MoveAdminMediaToFolderParams) error {
	cmd := d.MoveAdminMediaToFolderCmd(ctx, ac, params)
	return audited.Update(cmd)
}

// PSQL

// ListAdminMediaByFolder retrieves all admin media in a given folder (PostgreSQL).
func (d PsqlDatabase) ListAdminMediaByFolder(folderID types.NullableAdminMediaFolderID) (*[]AdminMedia, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminMediaByFolder(d.Context, mdbp.ListAdminMediaByFolderParams{FolderID: folderID})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin media by folder: %w", err)
	}
	res := make([]AdminMedia, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminMedia(v))
	}
	return &res, nil
}

// ListAdminMediaUnfiled retrieves all admin media with no folder assignment (PostgreSQL).
func (d PsqlDatabase) ListAdminMediaUnfiled() (*[]AdminMedia, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminMediaUnfiled(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list unfiled admin media: %w", err)
	}
	res := make([]AdminMedia, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminMedia(v))
	}
	return &res, nil
}

// CountAdminMediaByFolder returns the number of admin media in a given folder (PostgreSQL).
func (d PsqlDatabase) CountAdminMediaByFolder(folderID types.NullableAdminMediaFolderID) (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminMediaByFolder(d.Context, mdbp.CountAdminMediaByFolderParams{FolderID: folderID})
	if err != nil {
		return nil, fmt.Errorf("failed to count admin media by folder: %w", err)
	}
	return &c, nil
}

// CountAdminMediaUnfiled returns the number of admin media with no folder assignment (PostgreSQL).
func (d PsqlDatabase) CountAdminMediaUnfiled() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminMediaUnfiled(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count unfiled admin media: %w", err)
	}
	return &c, nil
}

func (d PsqlDatabase) MoveAdminMediaToFolderCmd(ctx context.Context, auditCtx audited.AuditContext, params MoveAdminMediaToFolderParams) MoveAdminMediaToFolderCmd {
	return MoveAdminMediaToFolderCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ListAdminMediaByFolderPaginated retrieves a paginated list of admin media in a given folder (PostgreSQL).
func (d PsqlDatabase) ListAdminMediaByFolderPaginated(params ListAdminMediaByFolderPaginatedParams) (*[]AdminMedia, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminMediaByFolderPaginated(d.Context, mdbp.ListAdminMediaByFolderPaginatedParams{
		FolderID: params.FolderID,
		Limit:    int32(params.Limit),
		Offset:   int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin media by folder paginated: %w", err)
	}
	res := make([]AdminMedia, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminMedia(v))
	}
	return &res, nil
}

// ListAdminMediaUnfiledPaginated retrieves a paginated list of admin media with no folder assignment (PostgreSQL).
func (d PsqlDatabase) ListAdminMediaUnfiledPaginated(params PaginationParams) (*[]AdminMedia, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminMediaUnfiledPaginated(d.Context, mdbp.ListAdminMediaUnfiledPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list unfiled admin media paginated: %w", err)
	}
	res := make([]AdminMedia, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminMedia(v))
	}
	return &res, nil
}

// MoveAdminMediaToFolder moves an admin media item to a folder and records an audit event (PostgreSQL).
func (d PsqlDatabase) MoveAdminMediaToFolder(ctx context.Context, ac audited.AuditContext, params MoveAdminMediaToFolderParams) error {
	cmd := d.MoveAdminMediaToFolderCmd(ctx, ac, params)
	return audited.Update(cmd)
}

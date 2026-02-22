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

// MediaDimensionsHistoryEntry represents a media dimension history record.
type MediaDimensionsHistoryEntry struct {
	MdID        string     `json:"md_id"`
	Label       NullString `json:"label"`
	Width       NullInt64  `json:"width"`
	Height      NullInt64  `json:"height"`
	AspectRatio NullString `json:"aspect_ratio"`
}

// CreateMediaDimensionFormParams contains form parameters for creating a media dimension.
type CreateMediaDimensionFormParams struct {
	Label       string `json:"label"`
	Width       string `json:"width"`
	Height      string `json:"height"`
	AspectRatio string `json:"aspect_ratio"`
}

// UpdateMediaDimensionFormParams contains form parameters for updating a media dimension.
type UpdateMediaDimensionFormParams struct {
	Label       string `json:"label"`
	Width       string `json:"width"`
	Height      string `json:"height"`
	AspectRatio string `json:"aspect_ratio"`
	MdID        string `json:"md_id"`
}

// MediaDimensionsJSON represents a media dimension as JSON.
type MediaDimensionsJSON struct {
	MdID        string     `json:"md_id"`
	Label       NullString `json:"label"`
	Width       NullInt64  `json:"width"`
	Height      NullInt64  `json:"height"`
	AspectRatio NullString `json:"aspect_ratio"`
}

// CreateMediaDimensionParamsJSON contains JSON parameters for creating a media dimension.
type CreateMediaDimensionParamsJSON struct {
	Label       NullString `json:"label"`
	Width       NullInt64  `json:"width"`
	Height      NullInt64  `json:"height"`
	AspectRatio NullString `json:"aspect_ratio"`
}

// UpdateMediaDimensionParamsJSON contains JSON parameters for updating a media dimension.
type UpdateMediaDimensionParamsJSON struct {
	Label       NullString `json:"label"`
	Width       NullInt64  `json:"width"`
	Height      NullInt64  `json:"height"`
	AspectRatio NullString `json:"aspect_ratio"`
	MdID        string     `json:"md_id"`
}

// MapCreateMediaDimensionParams converts form parameters to database parameters.
func MapCreateMediaDimensionParams(a CreateMediaDimensionFormParams) CreateMediaDimensionParams {
	return CreateMediaDimensionParams{
		Label:       NewNullString(a.Label),
		Width:       NullInt64{StringToNullInt64(a.Width)},
		Height:      NullInt64{StringToNullInt64(a.Height)},
		AspectRatio: NewNullString(a.AspectRatio),
	}
}

// MapUpdateMediaDimensionParams converts form parameters to database parameters.
func MapUpdateMediaDimensionParams(a UpdateMediaDimensionFormParams) UpdateMediaDimensionParams {
	return UpdateMediaDimensionParams{
		Label:       NewNullString(a.Label),
		Width:       NullInt64{StringToNullInt64(a.Width)},
		Height:      NullInt64{StringToNullInt64(a.Height)},
		AspectRatio: NewNullString(a.AspectRatio),
		MdID:        a.MdID,
	}
}

// MapCreateMediaDimensionJSONParams converts JSON parameters to database parameters.
func MapCreateMediaDimensionJSONParams(a CreateMediaDimensionParamsJSON) CreateMediaDimensionParams {
	return CreateMediaDimensionParams{
		Label:       a.Label,
		Width:       a.Width,
		Height:      a.Height,
		AspectRatio: a.AspectRatio,
	}
}

// MapUpdateMediaDimensionJSONParams converts JSON parameters to database parameters.
func MapUpdateMediaDimensionJSONParams(a UpdateMediaDimensionParamsJSON) UpdateMediaDimensionParams {
	return UpdateMediaDimensionParams{
		Label:       a.Label,
		Width:       a.Width,
		Height:      a.Height,
		AspectRatio: a.AspectRatio,
		MdID:        a.MdID,
	}
}

///////////////////////////////
// SQLITE MAPPERS
//////////////////////////////

// MapMediaDimension converts a sqlc-generated SQLite type to the wrapper type.
func (d Database) MapMediaDimension(a mdb.MediaDimensions) MediaDimensions {
	return MediaDimensions{
		MdID:        a.MdID,
		Label:       NullString{a.Label},
		Width:       NullInt64{a.Width},
		Height:      NullInt64{a.Height},
		AspectRatio: NullString{a.AspectRatio},
	}
}

// MapCreateMediaDimensionParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapCreateMediaDimensionParams(a CreateMediaDimensionParams) mdb.CreateMediaDimensionParams {
	return mdb.CreateMediaDimensionParams{
		MdID:        string(types.NewMediaDimensionID()),
		Label:       a.Label.NullString,
		Width:       a.Width.NullInt64,
		Height:      a.Height.NullInt64,
		AspectRatio: a.AspectRatio.NullString,
	}
}

// MapUpdateMediaDimensionParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapUpdateMediaDimensionParams(a UpdateMediaDimensionParams) mdb.UpdateMediaDimensionParams {
	return mdb.UpdateMediaDimensionParams{
		Label:       a.Label.NullString,
		Width:       a.Width.NullInt64,
		Height:      a.Height.NullInt64,
		AspectRatio: a.AspectRatio.NullString,
		MdID:        a.MdID,
	}
}

///////////////////////////////
// MYSQL MAPPERS
//////////////////////////////

// MapMediaDimension converts a sqlc-generated MySQL type to the wrapper type.
func (d MysqlDatabase) MapMediaDimension(a mdbm.MediaDimensions) MediaDimensions {
	return MediaDimensions{
		MdID:        a.MdID,
		Label:       NullString{a.Label},
		Width:       NullInt64{Int64ToNullInt64(int64(a.Width.Int32))},
		Height:      NullInt64{Int64ToNullInt64(int64(a.Height.Int32))},
		AspectRatio: NullString{a.AspectRatio},
	}
}

// MapCreateMediaDimensionParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapCreateMediaDimensionParams(a CreateMediaDimensionParams) mdbm.CreateMediaDimensionParams {
	return mdbm.CreateMediaDimensionParams{
		MdID:        string(types.NewMediaDimensionID()),
		Label:       a.Label.NullString,
		Width:       Int64ToNullInt32(a.Width.Int64),
		Height:      Int64ToNullInt32(a.Height.Int64),
		AspectRatio: a.AspectRatio.NullString,
	}
}

// MapUpdateMediaDimensionParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapUpdateMediaDimensionParams(a UpdateMediaDimensionParams) mdbm.UpdateMediaDimensionParams {
	return mdbm.UpdateMediaDimensionParams{
		Label:       a.Label.NullString,
		Width:       Int64ToNullInt32(a.Width.Int64),
		Height:      Int64ToNullInt32(a.Height.Int64),
		AspectRatio: a.AspectRatio.NullString,
		MdID:        a.MdID,
	}
}

///////////////////////////////
// POSTGRESQL MAPPERS
//////////////////////////////

// MapMediaDimension converts a sqlc-generated PostgreSQL type to the wrapper type.
func (d PsqlDatabase) MapMediaDimension(a mdbp.MediaDimensions) MediaDimensions {
	return MediaDimensions{
		MdID:        a.MdID,
		Label:       NullString{a.Label},
		Width:       NullInt64{Int64ToNullInt64(int64(a.Width.Int32))},
		Height:      NullInt64{Int64ToNullInt64(int64(a.Height.Int32))},
		AspectRatio: NullString{a.AspectRatio},
	}
}

// MapCreateMediaDimensionParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapCreateMediaDimensionParams(a CreateMediaDimensionParams) mdbp.CreateMediaDimensionParams {
	return mdbp.CreateMediaDimensionParams{
		MdID:        string(types.NewMediaDimensionID()),
		Label:       a.Label.NullString,
		Width:       Int64ToNullInt32(a.Width.Int64),
		Height:      Int64ToNullInt32(a.Height.Int64),
		AspectRatio: a.AspectRatio.NullString,
	}
}

// MapUpdateMediaDimensionParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapUpdateMediaDimensionParams(a UpdateMediaDimensionParams) mdbp.UpdateMediaDimensionParams {
	return mdbp.UpdateMediaDimensionParams{
		Label:       a.Label.NullString,
		Width:       Int64ToNullInt32(a.Width.Int64),
		Height:      Int64ToNullInt32(a.Height.Int64),
		AspectRatio: a.AspectRatio.NullString,
		MdID:        a.MdID,
	}
}

///////////////////////////////
// SQLITE QUERIES
//////////////////////////////

// CreateMediaDimension creates a new media dimension record (SQLite).
func (d Database) CreateMediaDimension(ctx context.Context, ac audited.AuditContext, s CreateMediaDimensionParams) (*MediaDimensions, error) {
	cmd := d.NewMediaDimensionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create media dimension: %w", err)
	}
	r := d.MapMediaDimension(result)
	return &r, nil
}

// UpdateMediaDimension updates an existing media dimension record (SQLite).
func (d Database) UpdateMediaDimension(ctx context.Context, ac audited.AuditContext, s UpdateMediaDimensionParams) (*string, error) {
	cmd := d.UpdateMediaDimensionCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update media dimension: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.MdID)
	return &msg, nil
}

// DeleteMediaDimension deletes a media dimension record (SQLite).
func (d Database) DeleteMediaDimension(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteMediaDimensionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

///////////////////////////////
// MYSQL QUERIES
//////////////////////////////

// CreateMediaDimension creates a new media dimension record (MySQL).
func (d MysqlDatabase) CreateMediaDimension(ctx context.Context, ac audited.AuditContext, s CreateMediaDimensionParams) (*MediaDimensions, error) {
	cmd := d.NewMediaDimensionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create media dimension: %w", err)
	}
	r := d.MapMediaDimension(result)
	return &r, nil
}

// UpdateMediaDimension updates an existing media dimension record (MySQL).
func (d MysqlDatabase) UpdateMediaDimension(ctx context.Context, ac audited.AuditContext, s UpdateMediaDimensionParams) (*string, error) {
	cmd := d.UpdateMediaDimensionCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update media dimension: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.MdID)
	return &msg, nil
}

// DeleteMediaDimension deletes a media dimension record (MySQL).
func (d MysqlDatabase) DeleteMediaDimension(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteMediaDimensionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

///////////////////////////////
// POSTGRES QUERIES
//////////////////////////////

// CreateMediaDimension creates a new media dimension record (PostgreSQL).
func (d PsqlDatabase) CreateMediaDimension(ctx context.Context, ac audited.AuditContext, s CreateMediaDimensionParams) (*MediaDimensions, error) {
	cmd := d.NewMediaDimensionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create media dimension: %w", err)
	}
	r := d.MapMediaDimension(result)
	return &r, nil
}

// UpdateMediaDimension updates an existing media dimension record (PostgreSQL).
func (d PsqlDatabase) UpdateMediaDimension(ctx context.Context, ac audited.AuditContext, s UpdateMediaDimensionParams) (*string, error) {
	cmd := d.UpdateMediaDimensionCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update media dimension: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.MdID)
	return &msg, nil
}

// DeleteMediaDimension deletes a media dimension record (PostgreSQL).
func (d PsqlDatabase) DeleteMediaDimension(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteMediaDimensionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

///////////////////////////////
// AUDITED COMMANDS -- SQLITE
//////////////////////////////

// NewMediaDimensionCmd is an audited create command for media_dimensions (SQLite).
type NewMediaDimensionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateMediaDimensionParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewMediaDimensionCmd) Context() context.Context              { return c.ctx }
func (c NewMediaDimensionCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewMediaDimensionCmd) Connection() *sql.DB                   { return c.conn }
func (c NewMediaDimensionCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewMediaDimensionCmd) TableName() string                     { return "media_dimensions" }
func (c NewMediaDimensionCmd) Params() any                           { return c.params }
func (c NewMediaDimensionCmd) GetID(row mdb.MediaDimensions) string  { return row.MdID }

func (c NewMediaDimensionCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.MediaDimensions, error) {
	queries := mdb.New(tx)
	return queries.CreateMediaDimension(ctx, mdb.CreateMediaDimensionParams{
		MdID:        string(types.NewMediaDimensionID()),
		Label:       c.params.Label.NullString,
		Width:       c.params.Width.NullInt64,
		Height:      c.params.Height.NullInt64,
		AspectRatio: c.params.AspectRatio.NullString,
	})
}

func (d Database) NewMediaDimensionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateMediaDimensionParams) NewMediaDimensionCmd {
	return NewMediaDimensionCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// UpdateMediaDimensionCmd is an audited update command for media_dimensions (SQLite).
type UpdateMediaDimensionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateMediaDimensionParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateMediaDimensionCmd) Context() context.Context              { return c.ctx }
func (c UpdateMediaDimensionCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateMediaDimensionCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateMediaDimensionCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateMediaDimensionCmd) TableName() string                     { return "media_dimensions" }
func (c UpdateMediaDimensionCmd) Params() any                           { return c.params }
func (c UpdateMediaDimensionCmd) GetID() string                         { return c.params.MdID }

func (c UpdateMediaDimensionCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.MediaDimensions, error) {
	queries := mdb.New(tx)
	return queries.GetMediaDimension(ctx, mdb.GetMediaDimensionParams{MdID: c.params.MdID})
}

func (c UpdateMediaDimensionCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateMediaDimension(ctx, mdb.UpdateMediaDimensionParams{
		Label:       c.params.Label.NullString,
		Width:       c.params.Width.NullInt64,
		Height:      c.params.Height.NullInt64,
		AspectRatio: c.params.AspectRatio.NullString,
		MdID:        c.params.MdID,
	})
}

func (d Database) UpdateMediaDimensionCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateMediaDimensionParams) UpdateMediaDimensionCmd {
	return UpdateMediaDimensionCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// DeleteMediaDimensionCmd is an audited delete command for media_dimensions (SQLite).
type DeleteMediaDimensionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteMediaDimensionCmd) Context() context.Context              { return c.ctx }
func (c DeleteMediaDimensionCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteMediaDimensionCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteMediaDimensionCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteMediaDimensionCmd) TableName() string                     { return "media_dimensions" }
func (c DeleteMediaDimensionCmd) GetID() string                         { return c.id }

func (c DeleteMediaDimensionCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.MediaDimensions, error) {
	queries := mdb.New(tx)
	return queries.GetMediaDimension(ctx, mdb.GetMediaDimensionParams{MdID: c.id})
}

func (c DeleteMediaDimensionCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteMediaDimension(ctx, mdb.DeleteMediaDimensionParams{MdID: c.id})
}

func (d Database) DeleteMediaDimensionCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteMediaDimensionCmd {
	return DeleteMediaDimensionCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

///////////////////////////////
// AUDITED COMMANDS -- MYSQL
//////////////////////////////

// NewMediaDimensionCmdMysql is an audited create command for media_dimensions (MySQL).
type NewMediaDimensionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateMediaDimensionParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewMediaDimensionCmdMysql) Context() context.Context              { return c.ctx }
func (c NewMediaDimensionCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewMediaDimensionCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewMediaDimensionCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewMediaDimensionCmdMysql) TableName() string                     { return "media_dimensions" }
func (c NewMediaDimensionCmdMysql) Params() any                           { return c.params }
func (c NewMediaDimensionCmdMysql) GetID(row mdbm.MediaDimensions) string { return row.MdID }

func (c NewMediaDimensionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.MediaDimensions, error) {
	id := string(types.NewMediaDimensionID())
	queries := mdbm.New(tx)
	err := queries.CreateMediaDimension(ctx, mdbm.CreateMediaDimensionParams{
		MdID:        id,
		Label:       c.params.Label.NullString,
		Width:       Int64ToNullInt32(c.params.Width.Int64),
		Height:      Int64ToNullInt32(c.params.Height.Int64),
		AspectRatio: c.params.AspectRatio.NullString,
	})
	if err != nil {
		return mdbm.MediaDimensions{}, fmt.Errorf("execute create media_dimensions: %w", err)
	}
	return queries.GetMediaDimension(ctx, mdbm.GetMediaDimensionParams{MdID: id})
}

func (d MysqlDatabase) NewMediaDimensionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateMediaDimensionParams) NewMediaDimensionCmdMysql {
	return NewMediaDimensionCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// UpdateMediaDimensionCmdMysql is an audited update command for media_dimensions (MySQL).
type UpdateMediaDimensionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateMediaDimensionParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateMediaDimensionCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateMediaDimensionCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateMediaDimensionCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateMediaDimensionCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateMediaDimensionCmdMysql) TableName() string                     { return "media_dimensions" }
func (c UpdateMediaDimensionCmdMysql) Params() any                           { return c.params }
func (c UpdateMediaDimensionCmdMysql) GetID() string                         { return c.params.MdID }

func (c UpdateMediaDimensionCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.MediaDimensions, error) {
	queries := mdbm.New(tx)
	return queries.GetMediaDimension(ctx, mdbm.GetMediaDimensionParams{MdID: c.params.MdID})
}

func (c UpdateMediaDimensionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateMediaDimension(ctx, mdbm.UpdateMediaDimensionParams{
		Label:       c.params.Label.NullString,
		Width:       Int64ToNullInt32(c.params.Width.Int64),
		Height:      Int64ToNullInt32(c.params.Height.Int64),
		AspectRatio: c.params.AspectRatio.NullString,
		MdID:        c.params.MdID,
	})
}

func (d MysqlDatabase) UpdateMediaDimensionCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateMediaDimensionParams) UpdateMediaDimensionCmdMysql {
	return UpdateMediaDimensionCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// DeleteMediaDimensionCmdMysql is an audited delete command for media_dimensions (MySQL).
type DeleteMediaDimensionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteMediaDimensionCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteMediaDimensionCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteMediaDimensionCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteMediaDimensionCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteMediaDimensionCmdMysql) TableName() string                     { return "media_dimensions" }
func (c DeleteMediaDimensionCmdMysql) GetID() string                         { return c.id }

func (c DeleteMediaDimensionCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.MediaDimensions, error) {
	queries := mdbm.New(tx)
	return queries.GetMediaDimension(ctx, mdbm.GetMediaDimensionParams{MdID: c.id})
}

func (c DeleteMediaDimensionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteMediaDimension(ctx, mdbm.DeleteMediaDimensionParams{MdID: c.id})
}

func (d MysqlDatabase) DeleteMediaDimensionCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteMediaDimensionCmdMysql {
	return DeleteMediaDimensionCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

///////////////////////////////
// AUDITED COMMANDS -- POSTGRES
//////////////////////////////

// NewMediaDimensionCmdPsql is an audited create command for media_dimensions (PostgreSQL).
type NewMediaDimensionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateMediaDimensionParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewMediaDimensionCmdPsql) Context() context.Context              { return c.ctx }
func (c NewMediaDimensionCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewMediaDimensionCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewMediaDimensionCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewMediaDimensionCmdPsql) TableName() string                     { return "media_dimensions" }
func (c NewMediaDimensionCmdPsql) Params() any                           { return c.params }
func (c NewMediaDimensionCmdPsql) GetID(row mdbp.MediaDimensions) string { return row.MdID }

func (c NewMediaDimensionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.MediaDimensions, error) {
	queries := mdbp.New(tx)
	return queries.CreateMediaDimension(ctx, mdbp.CreateMediaDimensionParams{
		MdID:        string(types.NewMediaDimensionID()),
		Label:       c.params.Label.NullString,
		Width:       Int64ToNullInt32(c.params.Width.Int64),
		Height:      Int64ToNullInt32(c.params.Height.Int64),
		AspectRatio: c.params.AspectRatio.NullString,
	})
}

func (d PsqlDatabase) NewMediaDimensionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateMediaDimensionParams) NewMediaDimensionCmdPsql {
	return NewMediaDimensionCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// UpdateMediaDimensionCmdPsql is an audited update command for media_dimensions (PostgreSQL).
type UpdateMediaDimensionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateMediaDimensionParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateMediaDimensionCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateMediaDimensionCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateMediaDimensionCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateMediaDimensionCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateMediaDimensionCmdPsql) TableName() string                     { return "media_dimensions" }
func (c UpdateMediaDimensionCmdPsql) Params() any                           { return c.params }
func (c UpdateMediaDimensionCmdPsql) GetID() string                         { return c.params.MdID }

func (c UpdateMediaDimensionCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.MediaDimensions, error) {
	queries := mdbp.New(tx)
	return queries.GetMediaDimension(ctx, mdbp.GetMediaDimensionParams{MdID: c.params.MdID})
}

func (c UpdateMediaDimensionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateMediaDimension(ctx, mdbp.UpdateMediaDimensionParams{
		Label:       c.params.Label.NullString,
		Width:       Int64ToNullInt32(c.params.Width.Int64),
		Height:      Int64ToNullInt32(c.params.Height.Int64),
		AspectRatio: c.params.AspectRatio.NullString,
		MdID:        c.params.MdID,
	})
}

func (d PsqlDatabase) UpdateMediaDimensionCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateMediaDimensionParams) UpdateMediaDimensionCmdPsql {
	return UpdateMediaDimensionCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// DeleteMediaDimensionCmdPsql is an audited delete command for media_dimensions (PostgreSQL).
type DeleteMediaDimensionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteMediaDimensionCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteMediaDimensionCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteMediaDimensionCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteMediaDimensionCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteMediaDimensionCmdPsql) TableName() string                     { return "media_dimensions" }
func (c DeleteMediaDimensionCmdPsql) GetID() string                         { return c.id }

func (c DeleteMediaDimensionCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.MediaDimensions, error) {
	queries := mdbp.New(tx)
	return queries.GetMediaDimension(ctx, mdbp.GetMediaDimensionParams{MdID: c.id})
}

func (c DeleteMediaDimensionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteMediaDimension(ctx, mdbp.DeleteMediaDimensionParams{MdID: c.id})
}

func (d PsqlDatabase) DeleteMediaDimensionCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteMediaDimensionCmdPsql {
	return DeleteMediaDimensionCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}

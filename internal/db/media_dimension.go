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
	"github.com/hegner123/modulacms/internal/utility"
)

// /////////////////////////////
// STRUCTS
// ////////////////////////////
type MediaDimensions struct {
	MdID        string         `json:"md_id"`
	Label       sql.NullString `json:"label"`
	Width       sql.NullInt64  `json:"width"`
	Height      sql.NullInt64  `json:"height"`
	AspectRatio sql.NullString `json:"aspect_ratio"`
}

type CreateMediaDimensionParams struct {
	Label       sql.NullString `json:"label"`
	Width       sql.NullInt64  `json:"width"`
	Height      sql.NullInt64  `json:"height"`
	AspectRatio sql.NullString `json:"aspect_ratio"`
}

type UpdateMediaDimensionParams struct {
	Label       sql.NullString `json:"label"`
	Width       sql.NullInt64  `json:"width"`
	Height      sql.NullInt64  `json:"height"`
	AspectRatio sql.NullString `json:"aspect_ratio"`
	MdID        string         `json:"md_id"`
}

type MediaDimensionsHistoryEntry struct {
	MdID        string         `json:"md_id"`
	Label       sql.NullString `json:"label"`
	Width       sql.NullInt64  `json:"width"`
	Height      sql.NullInt64  `json:"height"`
	AspectRatio sql.NullString `json:"aspect_ratio"`
}

type CreateMediaDimensionFormParams struct {
	Label       string `json:"label"`
	Width       string `json:"width"`
	Height      string `json:"height"`
	AspectRatio string `json:"aspect_ratio"`
}

type UpdateMediaDimensionFormParams struct {
	Label       string `json:"label"`
	Width       string `json:"width"`
	Height      string `json:"height"`
	AspectRatio string `json:"aspect_ratio"`
	MdID        string `json:"md_id"`
}
type MediaDimensionsJSON struct {
	MdID        string     `json:"md_id"`
	Label       NullString `json:"label"`
	Width       NullInt64  `json:"width"`
	Height      NullInt64  `json:"height"`
	AspectRatio NullString `json:"aspect_ratio"`
}

type CreateMediaDimensionParamsJSON struct {
	Label       NullString `json:"label"`
	Width       NullInt64  `json:"width"`
	Height      NullInt64  `json:"height"`
	AspectRatio NullString `json:"aspect_ratio"`
}

type UpdateMediaDimensionParamsJSON struct {
	Label       NullString `json:"label"`
	Width       NullInt64  `json:"width"`
	Height      NullInt64  `json:"height"`
	AspectRatio NullString `json:"aspect_ratio"`
	MdID        string     `json:"md_id"`
}

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreateMediaDimensionParams(a CreateMediaDimensionFormParams) CreateMediaDimensionParams {
	return CreateMediaDimensionParams{
		Label:       StringToNullString(a.Label),
		Width:       StringToNullInt64(a.Width),
		Height:      StringToNullInt64(a.Height),
		AspectRatio: StringToNullString(a.AspectRatio),
	}
}

func MapUpdateMediaDimensionParams(a UpdateMediaDimensionFormParams) UpdateMediaDimensionParams {
	return UpdateMediaDimensionParams{
		Label:       StringToNullString(a.Label),
		Width:       StringToNullInt64(a.Width),
		Height:      StringToNullInt64(a.Height),
		AspectRatio: StringToNullString(a.AspectRatio),
		MdID:        a.MdID,
	}
}

func MapStringMediaDimension(a MediaDimensions) StringMediaDimensions {
	return StringMediaDimensions{
		MdID:        a.MdID,
		Label:       utility.NullToString(a.Label),
		Width:       utility.NullToString(a.Width),
		Height:      utility.NullToString(a.Height),
		AspectRatio: utility.NullToString(a.AspectRatio),
	}
}
func MapCreateMediaDimensionJSONParams(a CreateMediaDimensionParamsJSON) CreateMediaDimensionParams {
	return CreateMediaDimensionParams{
		Label:       a.Label.NullString,
		Width:       a.Width.NullInt64,
		Height:      a.Height.NullInt64,
		AspectRatio: a.AspectRatio.NullString,
	}
}

func MapUpdateMediaDimensionJSONParams(a UpdateMediaDimensionParamsJSON) UpdateMediaDimensionParams {
	return UpdateMediaDimensionParams{
		Label:       a.Label.NullString,
		Width:       a.Width.NullInt64,
		Height:      a.Height.NullInt64,
		AspectRatio: a.AspectRatio.NullString,
		MdID:        a.MdID,
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

// /MAPS
func (d Database) MapMediaDimension(a mdb.MediaDimensions) MediaDimensions {
	return MediaDimensions{
		MdID:        a.MdID,
		Label:       a.Label,
		Width:       a.Width,
		Height:      a.Height,
		AspectRatio: a.AspectRatio,
	}
}

func (d Database) MapCreateMediaDimensionParams(a CreateMediaDimensionParams) mdb.CreateMediaDimensionParams {
	return mdb.CreateMediaDimensionParams{
		MdID:        string(types.NewMediaDimensionID()),
		Label:       a.Label,
		Width:       a.Width,
		Height:      a.Height,
		AspectRatio: a.AspectRatio,
	}
}

func (d Database) MapUpdateMediaDimensionParams(a UpdateMediaDimensionParams) mdb.UpdateMediaDimensionParams {
	return mdb.UpdateMediaDimensionParams{
		Label:       a.Label,
		Width:       a.Width,
		Height:      a.Height,
		AspectRatio: a.AspectRatio,
		MdID:        a.MdID,
	}
}

// /QUERIES
func (d Database) CountMediaDimensions() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountMediaDimension(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateMediaDimensionTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateMediaDimensionTable(d.Context)
	return err
}

func (d Database) CreateMediaDimension(ctx context.Context, ac audited.AuditContext, s CreateMediaDimensionParams) (*MediaDimensions, error) {
	cmd := d.NewMediaDimensionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create mediaDimension: %w", err)
	}
	r := d.MapMediaDimension(result)
	return &r, nil
}

func (d Database) DeleteMediaDimension(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteMediaDimensionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d Database) GetMediaDimension(id string) (*MediaDimensions, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetMediaDimension(d.Context, mdb.GetMediaDimensionParams{MdID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapMediaDimension(row)
	return &res, nil
}

func (d Database) ListMediaDimensions() (*[]MediaDimensions, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListMediaDimension(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get MediaDimensions: %v\n", err)
	}
	res := []MediaDimensions{}
	for _, v := range rows {
		m := d.MapMediaDimension(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateMediaDimension(ctx context.Context, ac audited.AuditContext, s UpdateMediaDimensionParams) (*string, error) {
	cmd := d.UpdateMediaDimensionCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update mediaDimension: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.MdID)
	return &msg, nil
}

///////////////////////////////
//MYSQL
//////////////////////////////

// /MAPS
func (d MysqlDatabase) MapMediaDimension(a mdbm.MediaDimensions) MediaDimensions {
	return MediaDimensions{
		MdID:        a.MdID,
		Label:       a.Label,
		Width:       Int64ToNullInt64(int64(a.Width.Int32)),
		Height:      Int64ToNullInt64(int64(a.Height.Int32)),
		AspectRatio: a.AspectRatio,
	}
}

func (d MysqlDatabase) MapCreateMediaDimensionParams(a CreateMediaDimensionParams) mdbm.CreateMediaDimensionParams {
	return mdbm.CreateMediaDimensionParams{
		MdID:        string(types.NewMediaDimensionID()),
		Label:       a.Label,
		Width:       Int64ToNullInt32(a.Width.Int64),
		Height:      Int64ToNullInt32(a.Height.Int64),
		AspectRatio: a.AspectRatio,
	}
}

func (d MysqlDatabase) MapUpdateMediaDimensionParams(a UpdateMediaDimensionParams) mdbm.UpdateMediaDimensionParams {
	return mdbm.UpdateMediaDimensionParams{
		Label:       a.Label,
		Width:       Int64ToNullInt32(a.Width.Int64),
		Height:      Int64ToNullInt32(a.Height.Int64),
		AspectRatio: a.AspectRatio,
		MdID:        a.MdID,
	}
}

// /QUERIES
func (d MysqlDatabase) CountMediaDimensions() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountMediaDimension(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateMediaDimensionTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateMediaDimensionTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateMediaDimension(ctx context.Context, ac audited.AuditContext, s CreateMediaDimensionParams) (*MediaDimensions, error) {
	cmd := d.NewMediaDimensionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create mediaDimension: %w", err)
	}
	r := d.MapMediaDimension(result)
	return &r, nil
}

func (d MysqlDatabase) DeleteMediaDimension(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteMediaDimensionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d MysqlDatabase) GetMediaDimension(id string) (*MediaDimensions, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetMediaDimension(d.Context, mdbm.GetMediaDimensionParams{MdID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapMediaDimension(row)
	return &res, nil
}

func (d MysqlDatabase) ListMediaDimensions() (*[]MediaDimensions, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListMediaDimension(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get MediaDimensions: %v\n", err)
	}
	res := []MediaDimensions{}
	for _, v := range rows {
		m := d.MapMediaDimension(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateMediaDimension(ctx context.Context, ac audited.AuditContext, s UpdateMediaDimensionParams) (*string, error) {
	cmd := d.UpdateMediaDimensionCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update mediaDimension: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.MdID)
	return &msg, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

// /MAPS
func (d PsqlDatabase) MapMediaDimension(a mdbp.MediaDimensions) MediaDimensions {
	return MediaDimensions{
		MdID:        a.MdID,
		Label:       a.Label,
		Width:       Int64ToNullInt64(int64(a.Width.Int32)),
		Height:      Int64ToNullInt64(int64(a.Height.Int32)),
		AspectRatio: a.AspectRatio,
	}
}

func (d PsqlDatabase) MapCreateMediaDimensionParams(a CreateMediaDimensionParams) mdbp.CreateMediaDimensionParams {
	return mdbp.CreateMediaDimensionParams{
		MdID:        string(types.NewMediaDimensionID()),
		Label:       a.Label,
		Width:       Int64ToNullInt32(a.Width.Int64),
		Height:      Int64ToNullInt32(a.Height.Int64),
		AspectRatio: a.AspectRatio,
	}
}

func (d PsqlDatabase) MapUpdateMediaDimensionParams(a UpdateMediaDimensionParams) mdbp.UpdateMediaDimensionParams {
	return mdbp.UpdateMediaDimensionParams{
		Label:       a.Label,
		Width:       Int64ToNullInt32(a.Width.Int64),
		Height:      Int64ToNullInt32(a.Height.Int64),
		AspectRatio: a.AspectRatio,
		MdID:        a.MdID,
	}
}

// /QUERIES
func (d PsqlDatabase) CountMediaDimensions() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountMediaDimension(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateMediaDimensionTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateMediaDimensionTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateMediaDimension(ctx context.Context, ac audited.AuditContext, s CreateMediaDimensionParams) (*MediaDimensions, error) {
	cmd := d.NewMediaDimensionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create mediaDimension: %w", err)
	}
	r := d.MapMediaDimension(result)
	return &r, nil
}

func (d PsqlDatabase) DeleteMediaDimension(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteMediaDimensionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d PsqlDatabase) GetMediaDimension(id string) (*MediaDimensions, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetMediaDimension(d.Context, mdbp.GetMediaDimensionParams{MdID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapMediaDimension(row)
	return &res, nil
}

func (d PsqlDatabase) ListMediaDimensions() (*[]MediaDimensions, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListMediaDimension(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get MediaDimensions: %v\n", err)
	}
	res := []MediaDimensions{}
	for _, v := range rows {
		m := d.MapMediaDimension(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateMediaDimension(ctx context.Context, ac audited.AuditContext, s UpdateMediaDimensionParams) (*string, error) {
	cmd := d.UpdateMediaDimensionCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update mediaDimension: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.MdID)
	return &msg, nil
}

///////////////////////////////
// AUDITED COMMANDS — SQLITE
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
		Label:       c.params.Label,
		Width:       c.params.Width,
		Height:      c.params.Height,
		AspectRatio: c.params.AspectRatio,
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
		Label:       c.params.Label,
		Width:       c.params.Width,
		Height:      c.params.Height,
		AspectRatio: c.params.AspectRatio,
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
// AUDITED COMMANDS — MYSQL
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
		Label:       c.params.Label,
		Width:       Int64ToNullInt32(c.params.Width.Int64),
		Height:      Int64ToNullInt32(c.params.Height.Int64),
		AspectRatio: c.params.AspectRatio,
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
		Label:       c.params.Label,
		Width:       Int64ToNullInt32(c.params.Width.Int64),
		Height:      Int64ToNullInt32(c.params.Height.Int64),
		AspectRatio: c.params.AspectRatio,
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
// AUDITED COMMANDS — POSTGRES
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
		Label:       c.params.Label,
		Width:       Int64ToNullInt32(c.params.Width.Int64),
		Height:      Int64ToNullInt32(c.params.Height.Int64),
		AspectRatio: c.params.AspectRatio,
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
		Label:       c.params.Label,
		Width:       Int64ToNullInt32(c.params.Width.Int64),
		Height:      Int64ToNullInt32(c.params.Height.Int64),
		AspectRatio: c.params.AspectRatio,
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

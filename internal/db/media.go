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

///////////////////////////////
// STRUCTS
//////////////////////////////

type Media struct {
	MediaID      types.MediaID        `json:"media_id"`
	Name         sql.NullString       `json:"name"`
	DisplayName  sql.NullString       `json:"display_name"`
	Alt          sql.NullString       `json:"alt"`
	Caption      sql.NullString       `json:"caption"`
	Description  sql.NullString       `json:"description"`
	Class        sql.NullString       `json:"class"`
	Mimetype     sql.NullString       `json:"mimetype"`
	Dimensions   sql.NullString       `json:"dimensions"`
	URL          types.URL            `json:"url"`
	Srcset       sql.NullString       `json:"srcset"`
	AuthorID     types.NullableUserID `json:"author_id"`
	DateCreated  types.Timestamp      `json:"date_created"`
	DateModified types.Timestamp      `json:"date_modified"`
}

type CreateMediaParams struct {
	Name         sql.NullString       `json:"name"`
	DisplayName  sql.NullString       `json:"display_name"`
	Alt          sql.NullString       `json:"alt"`
	Caption      sql.NullString       `json:"caption"`
	Description  sql.NullString       `json:"description"`
	Class        sql.NullString       `json:"class"`
	Mimetype     sql.NullString       `json:"mimetype"`
	Dimensions   sql.NullString       `json:"dimensions"`
	URL          types.URL            `json:"url"`
	Srcset       sql.NullString       `json:"srcset"`
	AuthorID     types.NullableUserID `json:"author_id"`
	DateCreated  types.Timestamp      `json:"date_created"`
	DateModified types.Timestamp      `json:"date_modified"`
}

type UpdateMediaParams struct {
	Name         sql.NullString       `json:"name"`
	DisplayName  sql.NullString       `json:"display_name"`
	Alt          sql.NullString       `json:"alt"`
	Caption      sql.NullString       `json:"caption"`
	Description  sql.NullString       `json:"description"`
	Class        sql.NullString       `json:"class"`
	Mimetype     sql.NullString       `json:"mimetype"`
	Dimensions   sql.NullString       `json:"dimensions"`
	URL          types.URL            `json:"url"`
	Srcset       sql.NullString       `json:"srcset"`
	AuthorID     types.NullableUserID `json:"author_id"`
	DateCreated  types.Timestamp      `json:"date_created"`
	DateModified types.Timestamp      `json:"date_modified"`
	MediaID      types.MediaID        `json:"media_id"`
}

// FormParams and JSON variants removed - use typed params directly

// MapStringMedia converts Media to StringMedia for display purposes
func MapStringMedia(a Media) StringMedia {
	return StringMedia{
		MediaID:      a.MediaID.String(),
		Name:         utility.NullToString(a.Name),
		DisplayName:  utility.NullToString(a.DisplayName),
		Alt:          utility.NullToString(a.Alt),
		Caption:      utility.NullToString(a.Caption),
		Description:  utility.NullToString(a.Description),
		Class:        utility.NullToString(a.Class),
		Mimetype:     utility.NullToString(a.Mimetype),
		Dimensions:   utility.NullToString(a.Dimensions),
		Url:          a.URL.String(),
		Srcset:       utility.NullToString(a.Srcset),
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapMedia(a mdb.Media) Media {
	return Media{
		MediaID:      a.MediaID,
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		URL:          a.URL,
		Srcset:       a.Srcset,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapCreateMediaParams(a CreateMediaParams) mdb.CreateMediaParams {
	return mdb.CreateMediaParams{
		MediaID:      types.NewMediaID(),
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		URL:          a.URL,
		Srcset:       a.Srcset,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapUpdateMediaParams(a UpdateMediaParams) mdb.UpdateMediaParams {
	return mdb.UpdateMediaParams{
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		URL:          a.URL,
		Srcset:       a.Srcset,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		MediaID:      a.MediaID,
	}
}

// QUERIES

func (d Database) CountMedia() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountMedia(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateMediaTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateMediaTable(d.Context)
	return err
}

func (d Database) CreateMedia(s CreateMediaParams) Media {
	params := d.MapCreateMediaParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateMedia(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateMedia: %v\n", err)
	}
	return d.MapMedia(row)
}

func (d Database) DeleteMedia(id types.MediaID) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteMedia(d.Context, mdb.DeleteMediaParams{MediaID: id})
	if err != nil {
		return fmt.Errorf("failed to delete Media: %v", id)
	}
	return nil
}

func (d Database) GetMedia(id types.MediaID) (*Media, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetMedia(d.Context, mdb.GetMediaParams{MediaID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d Database) GetMediaByName(name string) (*Media, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetMediaByName(d.Context, mdb.GetMediaByNameParams{Name: StringToNullString(name)})
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d Database) GetMediaByURL(url types.URL) (*Media, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetMediaByUrl(d.Context, mdb.GetMediaByUrlParams{URL: url})
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d Database) ListMedia() (*[]Media, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListMedia(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Media: %v\n", err)
	}
	res := []Media{}
	for _, v := range rows {
		m := d.MapMedia(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateMedia(s UpdateMediaParams) (*string, error) {
	params := d.MapUpdateMediaParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateMedia(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update media, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Name)
	return &u, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapMedia(a mdbm.Media) Media {
	return Media{
		MediaID:      a.MediaID,
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		URL:          a.URL,
		Srcset:       a.Srcset,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapCreateMediaParams(a CreateMediaParams) mdbm.CreateMediaParams {
	return mdbm.CreateMediaParams{
		MediaID:      types.NewMediaID(),
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		URL:          a.URL,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		Srcset:       a.Srcset,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapUpdateMediaParams(a UpdateMediaParams) mdbm.UpdateMediaParams {
	return mdbm.UpdateMediaParams{
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		URL:          a.URL,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		Srcset:       a.Srcset,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		MediaID:      a.MediaID,
	}
}

// QUERIES

func (d MysqlDatabase) CountMedia() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountMedia(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateMediaTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateMediaTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateMedia(s CreateMediaParams) Media {
	params := d.MapCreateMediaParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateMedia(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateMedia: %v\n", err)
	}
	row, err := queries.GetMedia(d.Context, mdbm.GetMediaParams{MediaID: params.MediaID})
	if err != nil {
		fmt.Printf("Failed to get last inserted Media: %v\n", err)
	}
	return d.MapMedia(row)
}

func (d MysqlDatabase) DeleteMedia(id types.MediaID) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteMedia(d.Context, mdbm.DeleteMediaParams{MediaID: id})
	if err != nil {
		return fmt.Errorf("failed to delete Media: %v", id)
	}
	return nil
}

func (d MysqlDatabase) GetMedia(id types.MediaID) (*Media, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetMedia(d.Context, mdbm.GetMediaParams{MediaID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d MysqlDatabase) GetMediaByName(name string) (*Media, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetMediaByName(d.Context, mdbm.GetMediaByNameParams{Name: StringToNullString(name)})
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d MysqlDatabase) GetMediaByURL(url types.URL) (*Media, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetMediaByUrl(d.Context, mdbm.GetMediaByUrlParams{URL: url})
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d MysqlDatabase) ListMedia() (*[]Media, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListMedia(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Media: %v\n", err)
	}
	res := []Media{}
	for _, v := range rows {
		m := d.MapMedia(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateMedia(s UpdateMediaParams) (*string, error) {
	params := d.MapUpdateMediaParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateMedia(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update media, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Name)
	return &u, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapMedia(a mdbp.Media) Media {
	return Media{
		MediaID:      a.MediaID,
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		URL:          a.URL,
		Srcset:       a.Srcset,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapCreateMediaParams(a CreateMediaParams) mdbp.CreateMediaParams {
	return mdbp.CreateMediaParams{
		MediaID:      types.NewMediaID(),
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		URL:          a.URL,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		Srcset:       a.Srcset,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapUpdateMediaParams(a UpdateMediaParams) mdbp.UpdateMediaParams {
	return mdbp.UpdateMediaParams{
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		URL:          a.URL,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		Srcset:       a.Srcset,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		MediaID:      a.MediaID,
	}
}

// QUERIES

func (d PsqlDatabase) CountMedia() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountMedia(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateMediaTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateMediaTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateMedia(s CreateMediaParams) Media {
	params := d.MapCreateMediaParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateMedia(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateMedia: %v\n", err)
	}
	return d.MapMedia(row)
}

func (d PsqlDatabase) DeleteMedia(id types.MediaID) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteMedia(d.Context, mdbp.DeleteMediaParams{MediaID: id})
	if err != nil {
		return fmt.Errorf("failed to delete Media: %v", id)
	}
	return nil
}

func (d PsqlDatabase) GetMedia(id types.MediaID) (*Media, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetMedia(d.Context, mdbp.GetMediaParams{MediaID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d PsqlDatabase) GetMediaByName(name string) (*Media, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetMediaByName(d.Context, mdbp.GetMediaByNameParams{Name: StringToNullString(name)})
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d PsqlDatabase) GetMediaByURL(url types.URL) (*Media, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetMediaByUrl(d.Context, mdbp.GetMediaByUrlParams{URL: url})
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d PsqlDatabase) ListMedia() (*[]Media, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListMedia(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Media: %v\n", err)
	}
	res := []Media{}
	for _, v := range rows {
		m := d.MapMedia(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateMedia(s UpdateMediaParams) (*string, error) {
	params := d.MapUpdateMediaParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateMedia(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update media, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Name)
	return &u, nil
}

///////////////////////////////
// AUDITED COMMANDS — SQLITE
//////////////////////////////

// NewMediaCmd is an audited create command for media (SQLite).
type NewMediaCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateMediaParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewMediaCmd) Context() context.Context              { return c.ctx }
func (c NewMediaCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewMediaCmd) Connection() *sql.DB                   { return c.conn }
func (c NewMediaCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewMediaCmd) TableName() string                     { return "media" }
func (c NewMediaCmd) Params() any                           { return c.params }
func (c NewMediaCmd) GetID(row mdb.Media) string            { return string(row.MediaID) }

func (c NewMediaCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Media, error) {
	queries := mdb.New(tx)
	return queries.CreateMedia(ctx, mdb.CreateMediaParams{
		MediaID:      types.NewMediaID(),
		Name:         c.params.Name,
		DisplayName:  c.params.DisplayName,
		Alt:          c.params.Alt,
		Caption:      c.params.Caption,
		Description:  c.params.Description,
		Class:        c.params.Class,
		Mimetype:     c.params.Mimetype,
		Dimensions:   c.params.Dimensions,
		URL:          c.params.URL,
		Srcset:       c.params.Srcset,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

func (d Database) NewMediaCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateMediaParams) NewMediaCmd {
	return NewMediaCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// UpdateMediaCmd is an audited update command for media (SQLite).
type UpdateMediaCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateMediaParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateMediaCmd) Context() context.Context              { return c.ctx }
func (c UpdateMediaCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateMediaCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateMediaCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateMediaCmd) TableName() string                     { return "media" }
func (c UpdateMediaCmd) Params() any                           { return c.params }
func (c UpdateMediaCmd) GetID() string                         { return string(c.params.MediaID) }

func (c UpdateMediaCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Media, error) {
	queries := mdb.New(tx)
	return queries.GetMedia(ctx, mdb.GetMediaParams{MediaID: c.params.MediaID})
}

func (c UpdateMediaCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateMedia(ctx, mdb.UpdateMediaParams{
		Name:         c.params.Name,
		DisplayName:  c.params.DisplayName,
		Alt:          c.params.Alt,
		Caption:      c.params.Caption,
		Description:  c.params.Description,
		Class:        c.params.Class,
		Mimetype:     c.params.Mimetype,
		Dimensions:   c.params.Dimensions,
		URL:          c.params.URL,
		Srcset:       c.params.Srcset,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		MediaID:      c.params.MediaID,
	})
}

func (d Database) UpdateMediaCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateMediaParams) UpdateMediaCmd {
	return UpdateMediaCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// DeleteMediaCmd is an audited delete command for media (SQLite).
type DeleteMediaCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.MediaID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteMediaCmd) Context() context.Context              { return c.ctx }
func (c DeleteMediaCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteMediaCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteMediaCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteMediaCmd) TableName() string                     { return "media" }
func (c DeleteMediaCmd) GetID() string                         { return string(c.id) }

func (c DeleteMediaCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Media, error) {
	queries := mdb.New(tx)
	return queries.GetMedia(ctx, mdb.GetMediaParams{MediaID: c.id})
}

func (c DeleteMediaCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteMedia(ctx, mdb.DeleteMediaParams{MediaID: c.id})
}

func (d Database) DeleteMediaCmd(ctx context.Context, auditCtx audited.AuditContext, id types.MediaID) DeleteMediaCmd {
	return DeleteMediaCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

///////////////////////////////
// AUDITED COMMANDS — MYSQL
//////////////////////////////

// NewMediaCmdMysql is an audited create command for media (MySQL).
type NewMediaCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateMediaParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewMediaCmdMysql) Context() context.Context              { return c.ctx }
func (c NewMediaCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewMediaCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewMediaCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewMediaCmdMysql) TableName() string                     { return "media" }
func (c NewMediaCmdMysql) Params() any                           { return c.params }
func (c NewMediaCmdMysql) GetID(row mdbm.Media) string           { return string(row.MediaID) }

func (c NewMediaCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Media, error) {
	id := types.NewMediaID()
	queries := mdbm.New(tx)
	err := queries.CreateMedia(ctx, mdbm.CreateMediaParams{
		MediaID:      id,
		Name:         c.params.Name,
		DisplayName:  c.params.DisplayName,
		Alt:          c.params.Alt,
		Caption:      c.params.Caption,
		Description:  c.params.Description,
		Class:        c.params.Class,
		URL:          c.params.URL,
		Mimetype:     c.params.Mimetype,
		Dimensions:   c.params.Dimensions,
		Srcset:       c.params.Srcset,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
	if err != nil {
		return mdbm.Media{}, fmt.Errorf("execute create media: %w", err)
	}
	return queries.GetMedia(ctx, mdbm.GetMediaParams{MediaID: id})
}

func (d MysqlDatabase) NewMediaCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateMediaParams) NewMediaCmdMysql {
	return NewMediaCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// UpdateMediaCmdMysql is an audited update command for media (MySQL).
type UpdateMediaCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateMediaParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateMediaCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateMediaCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateMediaCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateMediaCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateMediaCmdMysql) TableName() string                     { return "media" }
func (c UpdateMediaCmdMysql) Params() any                           { return c.params }
func (c UpdateMediaCmdMysql) GetID() string                         { return string(c.params.MediaID) }

func (c UpdateMediaCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Media, error) {
	queries := mdbm.New(tx)
	return queries.GetMedia(ctx, mdbm.GetMediaParams{MediaID: c.params.MediaID})
}

func (c UpdateMediaCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateMedia(ctx, mdbm.UpdateMediaParams{
		Name:         c.params.Name,
		DisplayName:  c.params.DisplayName,
		Alt:          c.params.Alt,
		Caption:      c.params.Caption,
		Description:  c.params.Description,
		Class:        c.params.Class,
		URL:          c.params.URL,
		Mimetype:     c.params.Mimetype,
		Dimensions:   c.params.Dimensions,
		Srcset:       c.params.Srcset,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		MediaID:      c.params.MediaID,
	})
}

func (d MysqlDatabase) UpdateMediaCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateMediaParams) UpdateMediaCmdMysql {
	return UpdateMediaCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// DeleteMediaCmdMysql is an audited delete command for media (MySQL).
type DeleteMediaCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.MediaID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteMediaCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteMediaCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteMediaCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteMediaCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteMediaCmdMysql) TableName() string                     { return "media" }
func (c DeleteMediaCmdMysql) GetID() string                         { return string(c.id) }

func (c DeleteMediaCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Media, error) {
	queries := mdbm.New(tx)
	return queries.GetMedia(ctx, mdbm.GetMediaParams{MediaID: c.id})
}

func (c DeleteMediaCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteMedia(ctx, mdbm.DeleteMediaParams{MediaID: c.id})
}

func (d MysqlDatabase) DeleteMediaCmd(ctx context.Context, auditCtx audited.AuditContext, id types.MediaID) DeleteMediaCmdMysql {
	return DeleteMediaCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

///////////////////////////////
// AUDITED COMMANDS — POSTGRES
//////////////////////////////

// NewMediaCmdPsql is an audited create command for media (PostgreSQL).
type NewMediaCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateMediaParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewMediaCmdPsql) Context() context.Context              { return c.ctx }
func (c NewMediaCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewMediaCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewMediaCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewMediaCmdPsql) TableName() string                     { return "media" }
func (c NewMediaCmdPsql) Params() any                           { return c.params }
func (c NewMediaCmdPsql) GetID(row mdbp.Media) string           { return string(row.MediaID) }

func (c NewMediaCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Media, error) {
	queries := mdbp.New(tx)
	return queries.CreateMedia(ctx, mdbp.CreateMediaParams{
		MediaID:      types.NewMediaID(),
		Name:         c.params.Name,
		DisplayName:  c.params.DisplayName,
		Alt:          c.params.Alt,
		Caption:      c.params.Caption,
		Description:  c.params.Description,
		Class:        c.params.Class,
		URL:          c.params.URL,
		Mimetype:     c.params.Mimetype,
		Dimensions:   c.params.Dimensions,
		Srcset:       c.params.Srcset,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

func (d PsqlDatabase) NewMediaCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateMediaParams) NewMediaCmdPsql {
	return NewMediaCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// UpdateMediaCmdPsql is an audited update command for media (PostgreSQL).
type UpdateMediaCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateMediaParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateMediaCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateMediaCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateMediaCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateMediaCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateMediaCmdPsql) TableName() string                     { return "media" }
func (c UpdateMediaCmdPsql) Params() any                           { return c.params }
func (c UpdateMediaCmdPsql) GetID() string                         { return string(c.params.MediaID) }

func (c UpdateMediaCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Media, error) {
	queries := mdbp.New(tx)
	return queries.GetMedia(ctx, mdbp.GetMediaParams{MediaID: c.params.MediaID})
}

func (c UpdateMediaCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateMedia(ctx, mdbp.UpdateMediaParams{
		Name:         c.params.Name,
		DisplayName:  c.params.DisplayName,
		Alt:          c.params.Alt,
		Caption:      c.params.Caption,
		Description:  c.params.Description,
		Class:        c.params.Class,
		URL:          c.params.URL,
		Mimetype:     c.params.Mimetype,
		Dimensions:   c.params.Dimensions,
		Srcset:       c.params.Srcset,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		MediaID:      c.params.MediaID,
	})
}

func (d PsqlDatabase) UpdateMediaCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateMediaParams) UpdateMediaCmdPsql {
	return UpdateMediaCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// DeleteMediaCmdPsql is an audited delete command for media (PostgreSQL).
type DeleteMediaCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.MediaID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteMediaCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteMediaCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteMediaCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteMediaCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteMediaCmdPsql) TableName() string                     { return "media" }
func (c DeleteMediaCmdPsql) GetID() string                         { return string(c.id) }

func (c DeleteMediaCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Media, error) {
	queries := mdbp.New(tx)
	return queries.GetMedia(ctx, mdbp.GetMediaParams{MediaID: c.id})
}

func (c DeleteMediaCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteMedia(ctx, mdbp.DeleteMediaParams{MediaID: c.id})
}

func (d PsqlDatabase) DeleteMediaCmd(ctx context.Context, auditCtx audited.AuditContext, id types.MediaID) DeleteMediaCmdPsql {
	return DeleteMediaCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}

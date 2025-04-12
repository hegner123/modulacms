package db

import (
	"database/sql"
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
)

///////////////////////////////
//STRUCTS
//////////////////////////////
type Media struct {
	MediaID      int64          `json:"media_id"`
	Name         sql.NullString `json:"name"`
	DisplayName  sql.NullString `json:"display_name"`
	Alt          sql.NullString `json:"alt"`
	Caption      sql.NullString `json:"caption"`
	Description  sql.NullString `json:"description"`
	Class        sql.NullString `json:"class"`
	Mimetype     sql.NullString `json:"mimetype"`
	Dimensions   sql.NullString `json:"dimensions"`
	Url          sql.NullString `json:"url"`
	Srcset       sql.NullString `json:"srcset"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type CreateMediaParams struct {
	Name         sql.NullString `json:"name"`
	DisplayName  sql.NullString `json:"display_name"`
	Alt          sql.NullString `json:"alt"`
	Caption      sql.NullString `json:"caption"`
	Description  sql.NullString `json:"description"`
	Class        sql.NullString `json:"class"`
	Url          sql.NullString `json:"url"`
	Mimetype     sql.NullString `json:"mimetype"`
	Dimensions   sql.NullString `json:"dimensions"`
	Srcset       sql.NullString `json:"srcset"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type UpdateMediaParams struct {
	Name         sql.NullString `json:"name"`
	DisplayName  sql.NullString `json:"display_name"`
	Alt          sql.NullString `json:"alt"`
	Caption      sql.NullString `json:"caption"`
	Description  sql.NullString `json:"description"`
	Class        sql.NullString `json:"class"`
	Url          sql.NullString `json:"url"`
	Mimetype     sql.NullString `json:"mimetype"`
	Dimensions   sql.NullString `json:"dimensions"`
	Srcset       sql.NullString `json:"srcset"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	MediaID      int64          `json:"media_id"`
}

type MediaHistoryEntry struct {
	MediaID      int64          `json:"media_id"`
	Name         sql.NullString `json:"name"`
	DisplayName  sql.NullString `json:"display_name"`
	Alt          sql.NullString `json:"alt"`
	Caption      sql.NullString `json:"caption"`
	Description  sql.NullString `json:"description"`
	Class        sql.NullString `json:"class"`
	Mimetype     sql.NullString `json:"mimetype"`
	Dimensions   sql.NullString `json:"dimensions"`
	Url          sql.NullString `json:"url"`
	Srcset       sql.NullString `json:"srcset"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type CreateMediaFormParams struct {
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	Alt          string `json:"alt"`
	Caption      string `json:"caption"`
	Description  string `json:"description"`
	Class        string `json:"class"`
	Url          string `json:"url"`
	Mimetype     string `json:"mimetype"`
	Dimensions   string `json:"dimensions"`
	Srcset       string `json:"srcset"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

type UpdateMediaFormParams struct {
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	Alt          string `json:"alt"`
	Caption      string `json:"caption"`
	Description  string `json:"description"`
	Class        string `json:"class"`
	Url          string `json:"url"`
	Mimetype     string `json:"mimetype"`
	Dimensions   string `json:"dimensions"`
	Srcset       string `json:"srcset"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	MediaID      string `json:"media_id"`
}
type MediaJSON struct {
	MediaID      int64          `json:"media_id"`
	Name         NullString `json:"name"`
	DisplayName  NullString `json:"display_name"`
	Alt          NullString `json:"alt"`
	Caption      NullString `json:"caption"`
	Description  NullString `json:"description"`
	Class        NullString `json:"class"`
	Mimetype     NullString `json:"mimetype"`
	Dimensions   NullString `json:"dimensions"`
	Url          NullString `json:"url"`
	Srcset       NullString `json:"srcset"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
}

type CreateMediaParamsJSON struct {
	Name         NullString `json:"name"`
	DisplayName  NullString `json:"display_name"`
	Alt          NullString `json:"alt"`
	Caption      NullString `json:"caption"`
	Description  NullString `json:"description"`
	Class        NullString `json:"class"`
	Url          NullString `json:"url"`
	Mimetype     NullString `json:"mimetype"`
	Dimensions   NullString `json:"dimensions"`
	Srcset       NullString `json:"srcset"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
}

type UpdateMediaParamsJSON struct {
	Name         NullString `json:"name"`
	DisplayName  NullString `json:"display_name"`
	Alt          NullString `json:"alt"`
	Caption      NullString `json:"caption"`
	Description  NullString `json:"description"`
	Class        NullString `json:"class"`
	Url          NullString `json:"url"`
	Mimetype     NullString `json:"mimetype"`
	Dimensions   NullString `json:"dimensions"`
	Srcset       NullString `json:"srcset"`
	AuthorID     int64          `json:"author_id"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
	MediaID      int64          `json:"media_id"`
}

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreateMediaParams(a CreateMediaFormParams) CreateMediaParams {
	return CreateMediaParams{
		Name:         StringToNullString(a.Name),
		DisplayName:  StringToNullString(a.DisplayName),
		Alt:          StringToNullString(a.Alt),
		Caption:      StringToNullString(a.Caption),
		Description:  StringToNullString(a.Description),
		Class:        StringToNullString(a.Class),
		Url:          StringToNullString(a.Url),
		Mimetype:     StringToNullString(a.Mimetype),
		Dimensions:   StringToNullString(a.Dimensions),
		Srcset:       StringToNullString(a.Srcset),
		AuthorID:     StringToInt64(a.AuthorID),
		DateCreated:  StringToNullString(a.DateCreated),
		DateModified: StringToNullString(a.DateModified),
	}
}

func MapUpdateMediaParams(a UpdateMediaFormParams) UpdateMediaParams {
	return UpdateMediaParams{
		Name:         StringToNullString(a.Name),
		DisplayName:  StringToNullString(a.DisplayName),
		Alt:          StringToNullString(a.Alt),
		Caption:      StringToNullString(a.Caption),
		Description:  StringToNullString(a.Description),
		Class:        StringToNullString(a.Class),
		Url:          StringToNullString(a.Url),
		Mimetype:     StringToNullString(a.Mimetype),
		Dimensions:   StringToNullString(a.Dimensions),
		Srcset:       StringToNullString(a.Srcset),
		AuthorID:     StringToInt64(a.AuthorID),
		DateCreated:  StringToNullString(a.DateCreated),
		DateModified: StringToNullString(a.DateModified),
		MediaID:      StringToInt64(a.MediaID),
	}
}

func MapStringMedia(a Media) StringMedia {
	return StringMedia{
		MediaID:      strconv.FormatInt(a.MediaID, 10),
		Name:         a.Name.String,
		DisplayName:  a.DisplayName.String,
		Alt:          a.Alt.String,
		Caption:      a.Caption.String,
		Description:  a.Description.String,
		Class:        a.Class.String,
		Mimetype:     a.Mimetype.String,
		Dimensions:   a.Dimensions.String,
		Url:          a.Url.String,
		Srcset:       a.Srcset.String,
		AuthorID:     strconv.FormatInt(a.AuthorID, 10),
		DateCreated:  a.DateCreated.String,
		DateModified: a.DateModified.String,
	}
}
func MapCreateMediaJSONParams(a CreateMediaParamsJSON) CreateMediaParams {
	return CreateMediaParams{
		Name:         a.Name.NullString,
		DisplayName:  a.DisplayName.NullString,
		Alt:          a.Alt.NullString,
		Caption:      a.Caption.NullString,
		Description:  a.Description.NullString,
		Class:        a.Class.NullString,
		Url:          a.Url.NullString,
		Mimetype:     a.Mimetype.NullString,
		Dimensions:   a.Dimensions.NullString,
		Srcset:       a.Srcset.NullString,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated.NullString,
		DateModified: a.DateModified.NullString,
	}
}

func MapUpdateMediaJSONParams(a UpdateMediaParamsJSON) UpdateMediaParams {
	return UpdateMediaParams{
		Name:         a.Name.NullString,
		DisplayName:  a.DisplayName.NullString,
		Alt:          a.Alt.NullString,
		Caption:      a.Caption.NullString,
		Description:  a.Description.NullString,
		Class:        a.Class.NullString,
		Url:          a.Url.NullString,
		Mimetype:     a.Mimetype.NullString,
		Dimensions:   a.Dimensions.NullString,
		Srcset:       a.Srcset.NullString,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated.NullString,
		DateModified: a.DateModified.NullString,
		MediaID:      a.MediaID,
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

///MAPS
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
		Url:          a.Url,
		Srcset:       a.Srcset,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapCreateMediaParams(a CreateMediaParams) mdb.CreateMediaParams {
	return mdb.CreateMediaParams{
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		Url:          a.Url,
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
		Url:          a.Url,
		Srcset:       a.Srcset,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		MediaID:      a.MediaID,
	}
}

///QUERIES
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

func (d Database) DeleteMedia(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteMedia(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Media: %v ", id)
	}
	return nil
}

func (d Database) GetMedia(id int64) (*Media, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetMedia(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d Database) GetMediaByName(name string) (*Media, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetMediaByName(d.Context, StringToNullString(name))
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d Database) GetMediaByURL(url string) (*Media, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetMediaByUrl(d.Context, StringToNullString(url))
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
//MYSQL
//////////////////////////////

///MAPS
func (d MysqlDatabase) MapMedia(a mdbm.Media) Media {
	return Media{
		MediaID:      int64(a.MediaID),
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		Url:          a.Url,
		Srcset:       a.Srcset,
		AuthorID:     int64(a.AuthorID),
		DateCreated:  StringToNullString(a.DateCreated.String()),
		DateModified: StringToNullString(a.DateModified.String()),
	}
}

func (d MysqlDatabase) MapCreateMediaParams(a CreateMediaParams) mdbm.CreateMediaParams {
	return mdbm.CreateMediaParams{
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		Url:          a.Url,
		Srcset:       a.Srcset,
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String).Time,
		DateModified: StringToNTime(a.DateModified.String).Time,
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
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		Url:          a.Url,
		Srcset:       a.Srcset,
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String).Time,
		DateModified: StringToNTime(a.DateModified.String).Time,
		MediaID:      int32(a.MediaID),
	}
}

///QUERIES
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
	row, err := queries.GetLastMedia(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted Media: %v\n", err)
	}
	return d.MapMedia(row)
}

func (d MysqlDatabase) DeleteMedia(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteMedia(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Media: %v ", id)
	}
	return nil
}

func (d MysqlDatabase) GetMedia(id int64) (*Media, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetMedia(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d MysqlDatabase) GetMediaByName(name string) (*Media, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetMediaByName(d.Context, StringToNullString(name))
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d MysqlDatabase) GetMediaByURL(url string) (*Media, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetMediaByUrl(d.Context, StringToNullString(url))
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
//POSTGRES
//////////////////////////////

///MAPS
func (d PsqlDatabase) MapMedia(a mdbp.Media) Media {
	return Media{
		MediaID:      int64(a.MediaID),
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		Url:          a.Url,
		Srcset:       a.Srcset,
		AuthorID:     int64(a.AuthorID),
		DateCreated:  StringToNullString(NullTimeToString(a.DateCreated)),
		DateModified: StringToNullString(NullTimeToString(a.DateModified)),
	}
}

func (d PsqlDatabase) MapCreateMediaParams(a CreateMediaParams) mdbp.CreateMediaParams {
	return mdbp.CreateMediaParams{
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		Url:          a.Url,
		Srcset:       a.Srcset,
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String),
		DateModified: StringToNTime(a.DateModified.String),
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
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		Url:          a.Url,
		Srcset:       a.Srcset,
		AuthorID:     int32(a.AuthorID),
		DateCreated:  StringToNTime(a.DateCreated.String),
		DateModified: StringToNTime(a.DateModified.String),
		MediaID:      int32(a.MediaID),
	}
}

///QUERIES
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

func (d PsqlDatabase) DeleteMedia(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteMedia(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Media: %v ", id)
	}
	return nil
}

func (d PsqlDatabase) GetMedia(id int64) (*Media, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetMedia(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d PsqlDatabase) GetMediaByName(name string) (*Media, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetMediaByName(d.Context, StringToNullString(name))
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d PsqlDatabase) GetMediaByURL(url string) (*Media, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetMediaByUrl(d.Context, StringToNullString(url))
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

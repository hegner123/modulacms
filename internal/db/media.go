package db

import (
	"database/sql"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
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
	row, err := queries.GetLastMedia(d.Context)
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

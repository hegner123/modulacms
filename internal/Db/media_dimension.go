package db

import (
	"database/sql"
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/db-mysql"
	mdbp "github.com/hegner123/modulacms/db-psql"
	mdb "github.com/hegner123/modulacms/db-sqlite"
)

///////////////////////////////
//STRUCTS
//////////////////////////////
type MediaDimensions struct {
	MdID        int64          `json:"md_id"`
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
	MdID        int64          `json:"md_id"`
}

type MediaDimensionsHistoryEntry struct {
	MdID        int64          `json:"md_id"`
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

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreateMediaDimensionParams(a CreateMediaDimensionFormParams) CreateMediaDimensionParams {
	return CreateMediaDimensionParams{
		Label:       Ns(a.Label),
		Width:       SNi64(a.Width),
		Height:      SNi64(a.Height),
		AspectRatio: Ns(a.AspectRatio),
	}
}

func MapUpdateMediaDimensionParams(a UpdateMediaDimensionFormParams) UpdateMediaDimensionParams {
	return UpdateMediaDimensionParams{
		Label:       Ns(a.Label),
		Width:       SNi64(a.Width),
		Height:      SNi64(a.Height),
		AspectRatio: Ns(a.AspectRatio),
		MdID:        Si(a.MdID),
	}
}

func MapStringMediaDimension(a MediaDimensions) StringMediaDimensions {
	return StringMediaDimensions{
		MdID:        strconv.FormatInt(a.MdID, 10),
		Label:       a.Label.String,
		Width:       strconv.FormatInt(a.Width.Int64, 10),
		Height:      strconv.FormatInt(a.Height.Int64, 10),
		AspectRatio: a.AspectRatio.String,
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

///MAPS
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

///QUERIES
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

func (d Database) CreateMediaDimension(s CreateMediaDimensionParams) MediaDimensions {
	params := d.MapCreateMediaDimensionParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateMediaDimension(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateMediaDimension: %v\n", err)
	}
	return d.MapMediaDimension(row)
}

func (d Database) DeleteMediaDimension(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteMediaDimension(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete MediaDimension: %v ", id)
	}
	return nil
}

func (d Database) GetMediaDimension(id int64) (*MediaDimensions, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetMediaDimension(d.Context, id)
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

func (d Database) UpdateMediaDimension(s UpdateMediaDimensionParams) (*string, error) {
	params := d.MapUpdateMediaDimensionParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateMediaDimension(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update mediadimenion, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
//MYSQL
//////////////////////////////

///MAPS
func (d MysqlDatabase) MapMediaDimension(a mdbm.MediaDimensions) MediaDimensions {
	return MediaDimensions{
		MdID:        int64(a.MdID),
		Label:       a.Label,
		Width:       Ni64(int64(a.Width.Int32)),
		Height:      Ni64(int64(a.Height.Int32)),
		AspectRatio: a.AspectRatio,
	}
}

func (d MysqlDatabase) MapCreateMediaDimensionParams(a CreateMediaDimensionParams) mdbm.CreateMediaDimensionParams {
	return mdbm.CreateMediaDimensionParams{
		Label:       a.Label,
		Width:       Ni32(a.Width.Int64),
		Height:      Ni32(a.Height.Int64),
		AspectRatio: a.AspectRatio,
	}
}

func (d MysqlDatabase) MapUpdateMediaDimensionParams(a UpdateMediaDimensionParams) mdbm.UpdateMediaDimensionParams {
	return mdbm.UpdateMediaDimensionParams{
		Label:       a.Label,
		Width:       Ni32(a.Width.Int64),
		Height:      Ni32(a.Height.Int64),
		AspectRatio: a.AspectRatio,
		MdID:        int32(a.MdID),
	}
}

///QUERIES
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

func (d MysqlDatabase) CreateMediaDimension(s CreateMediaDimensionParams) MediaDimensions {
	params := d.MapCreateMediaDimensionParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateMediaDimension(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateMediaDimension: %v\n", err)
	}
	row, err := queries.GetLastMediaDimension(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted MediaDimension: %v\n", err)
	}
	return d.MapMediaDimension(row)
}

func (d MysqlDatabase) DeleteMediaDimension(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteMediaDimension(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete MediaDimension: %v ", id)
	}
	return nil
}

func (d MysqlDatabase) GetMediaDimension(id int64) (*MediaDimensions, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetMediaDimension(d.Context, int32(id))
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

func (d MysqlDatabase) UpdateMediaDimension(s UpdateMediaDimensionParams) (*string, error) {
	params := d.MapUpdateMediaDimensionParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateMediaDimension(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update media dimension, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

///MAPS
func (d PsqlDatabase) MapMediaDimension(a mdbp.MediaDimensions) MediaDimensions {
	return MediaDimensions{
		MdID:        int64(a.MdID),
		Label:       a.Label,
		Width:       Ni64(int64(a.Width.Int32)),
		Height:      Ni64(int64(a.Height.Int32)),
		AspectRatio: a.AspectRatio,
	}
}

func (d PsqlDatabase) MapCreateMediaDimensionParams(a CreateMediaDimensionParams) mdbp.CreateMediaDimensionParams {
	return mdbp.CreateMediaDimensionParams{
		Label:       a.Label,
		Width:       Ni32(a.Width.Int64),
		Height:      Ni32(a.Height.Int64),
		AspectRatio: a.AspectRatio,
	}
}

func (d PsqlDatabase) MapUpdateMediaDimensionParams(a UpdateMediaDimensionParams) mdbp.UpdateMediaDimensionParams {
	return mdbp.UpdateMediaDimensionParams{
		Label:       a.Label,
		Width:       Ni32(a.Width.Int64),
		Height:      Ni32(a.Height.Int64),
		AspectRatio: a.AspectRatio,
		MdID:        int32(a.MdID),
	}
}

///QUERIES
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

func (d PsqlDatabase) CreateMediaDimension(s CreateMediaDimensionParams) MediaDimensions {
	params := d.MapCreateMediaDimensionParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateMediaDimension(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateMediaDimension: %v\n", err)
	}
	return d.MapMediaDimension(row)
}

func (d PsqlDatabase) DeleteMediaDimension(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteMediaDimension(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete MediaDimension: %v ", id)
	}
	return nil
}

func (d PsqlDatabase) GetMediaDimension(id int64) (*MediaDimensions, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetMediaDimension(d.Context, int32(id))
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

func (d PsqlDatabase) UpdateMediaDimension(s UpdateMediaDimensionParams) (*string, error) {
	params := d.MapUpdateMediaDimensionParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateMediaDimension(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update media dimension, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

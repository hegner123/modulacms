package db

import (
	"context"
	"database/sql"
)

type DbStatus string

const (
	open   DbStatus = "open"
	closed DbStatus = "closed"
	err    DbStatus = "error"
)

type Driver interface {
    GetDb() string
}

type Database struct {
	Src            string
	Status         DbStatus
	Connection     *sql.DB
	LastConnection string
	Err            error
	Context        context.Context
}

type MysqlDatabase struct {
	Src            string
	Status         DbStatus
	Connection     *sql.DB
	LastConnection string
	Err            error
	Context        context.Context
}
type PsqlDatabase struct {
	Src            string
	Status         DbStatus
	Connection     *sql.DB
	LastConnection string
	Err            error
	Context        context.Context
}
type CreateRoleParams struct {
	Label       string `json:"label"`
	Permissions string `json:"permissions"`
}

type CreateRouteParams struct {
	Author       string         `json:"author"`
	AuthorID     int64          `json:"author_id"`
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Status       int64          `json:"status"`
	History      sql.NullString `json:"history"`
	DateCreated  sql.NullTime   `json:"date_created"`
	DateModified sql.NullTime   `json:"date_modified"`
}

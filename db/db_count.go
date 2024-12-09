package db

import (
	"context"
	"database/sql"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func countAdminRoutes(db *sql.DB, ctx context.Context) int64 {
	queries := mdb.New(db)
	c, err := queries.CountAdminroute(ctx)
    if err != nil { 
        logError("error counting admin routes ", err)
    }
    return c
}
func countDatatypes(db *sql.DB, ctx context.Context) int64 {
	queries := mdb.New(db)
	c, err := queries.CountDatatype(ctx)
    if err != nil { 
        logError("error counting datatypes ", err)
    }
    return c
}
func countField(db *sql.DB, ctx context.Context) int64 {
	queries := mdb.New(db)
	c, err := queries.CountField(ctx)
    if err != nil { 
        logError("error counting field", err)
    }
    return c
}
func countMedia(db *sql.DB, ctx context.Context) int64 {
	queries := mdb.New(db)
	c, err := queries.CountMedia(ctx)
    if err != nil { 
        logError("error counting media", err)
    }
    return c
}
func countTables(db *sql.DB, ctx context.Context) int64 {
	queries := mdb.New(db)
	c, err := queries.CountTables(ctx)
    if err != nil { 
        logError("error counting tables", err)
    }
    return c
}
func countTokens(db *sql.DB, ctx context.Context) int64 {
	queries := mdb.New(db)
	c, err := queries.CountTokens(ctx)
    if err != nil { 
        logError("error counting tokens", err)
    }
    return c
}
func countUsers(db *sql.DB, ctx context.Context) int64 {
	queries := mdb.New(db)
	c, err := queries.CountUsers(ctx)
    if err != nil { 
        logError("error counting users", err)
    }
    return c
}

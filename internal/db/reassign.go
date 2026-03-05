package db

import (
	"context"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

///////////////////////////////
// CONTENT DATA REASSIGNMENT
//////////////////////////////

// ReassignContentDataAuthor bulk-reassigns all content_data from one author to another.
func (d Database) ReassignContentDataAuthor(ctx context.Context, newAuthor, oldAuthor types.UserID) error {
	queries := mdb.New(d.Connection)
	return queries.ReassignContentDataAuthor(ctx, mdb.ReassignContentDataAuthorParams{
		AuthorID:   newAuthor,
		AuthorID_2: oldAuthor,
	})
}

func (d MysqlDatabase) ReassignContentDataAuthor(ctx context.Context, newAuthor, oldAuthor types.UserID) error {
	queries := mdbm.New(d.Connection)
	return queries.ReassignContentDataAuthor(ctx, mdbm.ReassignContentDataAuthorParams{
		AuthorID:   newAuthor,
		AuthorID_2: oldAuthor,
	})
}

func (d PsqlDatabase) ReassignContentDataAuthor(ctx context.Context, newAuthor, oldAuthor types.UserID) error {
	queries := mdbp.New(d.Connection)
	return queries.ReassignContentDataAuthor(ctx, mdbp.ReassignContentDataAuthorParams{
		AuthorID:   newAuthor,
		AuthorID_2: oldAuthor,
	})
}

// CountContentDataByAuthor returns the number of content_data rows owned by a user.
func (d Database) CountContentDataByAuthor(ctx context.Context, authorID types.UserID) (int64, error) {
	queries := mdb.New(d.Connection)
	count, err := queries.CountContentDataByAuthor(ctx, mdb.CountContentDataByAuthorParams{AuthorID: authorID})
	if err != nil {
		return 0, fmt.Errorf("failed to count content data by author: %w", err)
	}
	return count, nil
}

func (d MysqlDatabase) CountContentDataByAuthor(ctx context.Context, authorID types.UserID) (int64, error) {
	queries := mdbm.New(d.Connection)
	count, err := queries.CountContentDataByAuthor(ctx, mdbm.CountContentDataByAuthorParams{AuthorID: authorID})
	if err != nil {
		return 0, fmt.Errorf("failed to count content data by author: %w", err)
	}
	return count, nil
}

func (d PsqlDatabase) CountContentDataByAuthor(ctx context.Context, authorID types.UserID) (int64, error) {
	queries := mdbp.New(d.Connection)
	count, err := queries.CountContentDataByAuthor(ctx, mdbp.CountContentDataByAuthorParams{AuthorID: authorID})
	if err != nil {
		return 0, fmt.Errorf("failed to count content data by author: %w", err)
	}
	return count, nil
}

///////////////////////////////
// DATATYPE REASSIGNMENT
//////////////////////////////

// ReassignDatatypeAuthor bulk-reassigns all datatypes from one author to another.
func (d Database) ReassignDatatypeAuthor(ctx context.Context, newAuthor, oldAuthor types.UserID) error {
	queries := mdb.New(d.Connection)
	return queries.ReassignDatatypeAuthor(ctx, mdb.ReassignDatatypeAuthorParams{
		AuthorID:   newAuthor,
		AuthorID_2: oldAuthor,
	})
}

func (d MysqlDatabase) ReassignDatatypeAuthor(ctx context.Context, newAuthor, oldAuthor types.UserID) error {
	queries := mdbm.New(d.Connection)
	return queries.ReassignDatatypeAuthor(ctx, mdbm.ReassignDatatypeAuthorParams{
		AuthorID:   newAuthor,
		AuthorID_2: oldAuthor,
	})
}

func (d PsqlDatabase) ReassignDatatypeAuthor(ctx context.Context, newAuthor, oldAuthor types.UserID) error {
	queries := mdbp.New(d.Connection)
	return queries.ReassignDatatypeAuthor(ctx, mdbp.ReassignDatatypeAuthorParams{
		AuthorID:   newAuthor,
		AuthorID_2: oldAuthor,
	})
}

// CountDatatypesByAuthor returns the number of datatypes owned by a user.
func (d Database) CountDatatypesByAuthor(ctx context.Context, authorID types.UserID) (int64, error) {
	queries := mdb.New(d.Connection)
	count, err := queries.CountDatatypesByAuthor(ctx, mdb.CountDatatypesByAuthorParams{AuthorID: authorID})
	if err != nil {
		return 0, fmt.Errorf("failed to count datatypes by author: %w", err)
	}
	return count, nil
}

func (d MysqlDatabase) CountDatatypesByAuthor(ctx context.Context, authorID types.UserID) (int64, error) {
	queries := mdbm.New(d.Connection)
	count, err := queries.CountDatatypesByAuthor(ctx, mdbm.CountDatatypesByAuthorParams{AuthorID: authorID})
	if err != nil {
		return 0, fmt.Errorf("failed to count datatypes by author: %w", err)
	}
	return count, nil
}

func (d PsqlDatabase) CountDatatypesByAuthor(ctx context.Context, authorID types.UserID) (int64, error) {
	queries := mdbp.New(d.Connection)
	count, err := queries.CountDatatypesByAuthor(ctx, mdbp.CountDatatypesByAuthorParams{AuthorID: authorID})
	if err != nil {
		return 0, fmt.Errorf("failed to count datatypes by author: %w", err)
	}
	return count, nil
}

///////////////////////////////
// ADMIN CONTENT DATA REASSIGNMENT
//////////////////////////////

// ReassignAdminContentDataAuthor bulk-reassigns all admin_content_data from one author to another.
func (d Database) ReassignAdminContentDataAuthor(ctx context.Context, newAuthor, oldAuthor types.UserID) error {
	queries := mdb.New(d.Connection)
	return queries.ReassignAdminContentDataAuthor(ctx, mdb.ReassignAdminContentDataAuthorParams{
		AuthorID:   newAuthor,
		AuthorID_2: oldAuthor,
	})
}

func (d MysqlDatabase) ReassignAdminContentDataAuthor(ctx context.Context, newAuthor, oldAuthor types.UserID) error {
	queries := mdbm.New(d.Connection)
	return queries.ReassignAdminContentDataAuthor(ctx, mdbm.ReassignAdminContentDataAuthorParams{
		AuthorID:   newAuthor,
		AuthorID_2: oldAuthor,
	})
}

func (d PsqlDatabase) ReassignAdminContentDataAuthor(ctx context.Context, newAuthor, oldAuthor types.UserID) error {
	queries := mdbp.New(d.Connection)
	return queries.ReassignAdminContentDataAuthor(ctx, mdbp.ReassignAdminContentDataAuthorParams{
		AuthorID:   newAuthor,
		AuthorID_2: oldAuthor,
	})
}

// CountAdminContentDataByAuthor returns the number of admin_content_data rows owned by a user.
func (d Database) CountAdminContentDataByAuthor(ctx context.Context, authorID types.UserID) (int64, error) {
	queries := mdb.New(d.Connection)
	count, err := queries.CountAdminContentDataByAuthor(ctx, mdb.CountAdminContentDataByAuthorParams{AuthorID: authorID})
	if err != nil {
		return 0, fmt.Errorf("failed to count admin content data by author: %w", err)
	}
	return count, nil
}

func (d MysqlDatabase) CountAdminContentDataByAuthor(ctx context.Context, authorID types.UserID) (int64, error) {
	queries := mdbm.New(d.Connection)
	count, err := queries.CountAdminContentDataByAuthor(ctx, mdbm.CountAdminContentDataByAuthorParams{AuthorID: authorID})
	if err != nil {
		return 0, fmt.Errorf("failed to count admin content data by author: %w", err)
	}
	return count, nil
}

func (d PsqlDatabase) CountAdminContentDataByAuthor(ctx context.Context, authorID types.UserID) (int64, error) {
	queries := mdbp.New(d.Connection)
	count, err := queries.CountAdminContentDataByAuthor(ctx, mdbp.CountAdminContentDataByAuthorParams{AuthorID: authorID})
	if err != nil {
		return 0, fmt.Errorf("failed to count admin content data by author: %w", err)
	}
	return count, nil
}

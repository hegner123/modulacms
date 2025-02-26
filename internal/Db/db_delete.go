package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func DeleteAdminDatatype(db *sql.DB, ctx context.Context, id int64) (*string, error) {
	queries := mdb.New(db)
	err := queries.DeleteAdminDatatype(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete Admin Route %d ", id)
	}
	s := fmt.Sprintf("Deleted Admin Route %d successfully", id)
	return &s, nil
}

func DeleteAdminField(db *sql.DB, ctx context.Context, id int64) (*string, error) {
	queries := mdb.New(db)
	err := queries.DeleteAdminField(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete Admin Route %d ", id)
	}
	s := fmt.Sprintf("Deleted Admin Route %d successfully", id)
	return &s, nil
}

func DeleteAdminRoute(db *sql.DB, ctx context.Context, slug string) (*string, error) {
	queries := mdb.New(db)
	err := queries.DeleteAdminRoute(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to delete Admin Route %s ", slug)
	}
	s := fmt.Sprintf("Deleted Admin Route %s successfully", slug)
	return &s, nil
}

func DeleteContentData(db *sql.DB, ctx context.Context, id int64) (*string, error) {
	queries := mdb.New(db)
	err := queries.DeleteContentData(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete content data %d ", id)
	}
	s := fmt.Sprintf("Deleted ContentData %d successfully", id)
	return &s, nil
}

func DeleteContentField(db *sql.DB, ctx context.Context, id int64) (*string, error) {
	queries := mdb.New(db)
	err := queries.DeleteContentField(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete content field %d ", id)
	}
	s := fmt.Sprintf("Deleted Content Field %d successfully", id)
	return &s, nil
}

func DeleteDatatype(db *sql.DB, ctx context.Context, id int64) (*string, error) {
	queries := mdb.New(db)
	err := queries.DeleteDatatype(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete datatype %d ", id)
	}
	s := fmt.Sprintf("Deleted Field %d successfully", id)
	return &s, nil
}

func DeleteField(db *sql.DB, ctx context.Context, id int64) (*string, error) {
	queries := mdb.New(db)
	err := queries.DeleteField(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("failed to delete Field %d ", id)
	}
	s := fmt.Sprintf("Deleted Field %d successfully", id)
	return &s, nil
}

func DeleteMedia(db *sql.DB, ctx context.Context, id int64) (*string, error) {
	queries := mdb.New(db)
	err := queries.DeleteMedia(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("failed to delete Media %d ", id)
	}
	s := fmt.Sprintf("Deleted Media %d successfully", id)
	return &s, nil
}

func DeleteMediaDimension(db *sql.DB, ctx context.Context, id int64) (*string, error) {
	queries := mdb.New(db)
	err := queries.DeleteMediaDimension(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("failed to delete MediaDimension %d ", id)
	}
	s := fmt.Sprintf("Deleted Media Dimension %d successfully", id)
	return &s, nil
}

func DeleteRoute(db *sql.DB, ctx context.Context, slug string) (*string, error) {
	queries := mdb.New(db)
	err := queries.DeleteRoute(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to delete  Route %s ", slug)
	}
	s := fmt.Sprintf("Deleted Route %s successfully", slug)
	return &s, nil
}

func DeleteTable(db *sql.DB, ctx context.Context, id int64) (*string, error) {
	queries := mdb.New(db)
	err := queries.DeleteTable(ctx, id)
	if err != nil {
		return nil, err
	}
	s := fmt.Sprintf("Deleted Table %d successfully", id)
	return &s, nil
}

func DeleteToken(db *sql.DB, ctx context.Context, id int64) (*string, error) {
	queries := mdb.New(db)
	err := queries.DeleteToken(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete Token %d ", id)
	}
	s := fmt.Sprintf("Deleted Table %d successfully", id)
	return &s, nil
}

func DeleteUser(db *sql.DB, ctx context.Context, id int64) (*string, error) {
	queries := mdb.New(db)
	err := queries.DeleteUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete User %d ", id)
	}
	s := fmt.Sprintf("Deleted User %d successfully", id)
	return &s, nil
}

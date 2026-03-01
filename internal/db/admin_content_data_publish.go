package db

import (
	"context"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

// UpdateAdminContentDataPublishMetaParams contains parameters for updating admin publish metadata.
type UpdateAdminContentDataPublishMetaParams struct {
	Status             types.ContentStatus  `json:"status"`
	PublishedAt        types.Timestamp      `json:"published_at"`
	PublishedBy        types.NullableUserID `json:"published_by"`
	DateModified       types.Timestamp      `json:"date_modified"`
	AdminContentDataID types.AdminContentID `json:"admin_content_data_id"`
}

// UpdateAdminContentDataWithRevisionParams contains parameters for updating admin content data with a new revision.
type UpdateAdminContentDataWithRevisionParams struct {
	AdminRouteID       types.NullableAdminRouteID    `json:"admin_route_id"`
	ParentID           types.NullableAdminContentID  `json:"parent_id"`
	FirstChildID       types.NullableAdminContentID  `json:"first_child_id"`
	NextSiblingID      types.NullableAdminContentID  `json:"next_sibling_id"`
	PrevSiblingID      types.NullableAdminContentID  `json:"prev_sibling_id"`
	AdminDatatypeID    types.NullableAdminDatatypeID `json:"admin_datatype_id"`
	AuthorID           types.UserID                  `json:"author_id"`
	Status             types.ContentStatus           `json:"status"`
	DateCreated        types.Timestamp               `json:"date_created"`
	DateModified       types.Timestamp               `json:"date_modified"`
	AdminContentDataID types.AdminContentID          `json:"admin_content_data_id"`
	Revision           int64                         `json:"revision"`
}

// UpdateAdminContentDataScheduleParams contains parameters for scheduling admin content publication.
type UpdateAdminContentDataScheduleParams struct {
	PublishAt          types.Timestamp      `json:"publish_at"`
	DateModified       types.Timestamp      `json:"date_modified"`
	AdminContentDataID types.AdminContentID `json:"admin_content_data_id"`
}

// ClearAdminContentDataScheduleParams contains parameters for clearing a scheduled admin publication.
type ClearAdminContentDataScheduleParams struct {
	DateModified       types.Timestamp      `json:"date_modified"`
	AdminContentDataID types.AdminContentID `json:"admin_content_data_id"`
}

// ===== SQLite =====

func (d Database) UpdateAdminContentDataPublishMeta(ctx context.Context, p UpdateAdminContentDataPublishMetaParams) error {
	queries := mdb.New(d.Connection)
	return queries.UpdateAdminContentDataPublishMeta(ctx, mdb.UpdateAdminContentDataPublishMetaParams{
		Status:             p.Status,
		PublishedAt:        p.PublishedAt,
		PublishedBy:        p.PublishedBy,
		DateModified:       p.DateModified,
		AdminContentDataID: p.AdminContentDataID,
	})
}

func (d Database) UpdateAdminContentDataWithRevision(ctx context.Context, p UpdateAdminContentDataWithRevisionParams) error {
	queries := mdb.New(d.Connection)
	return queries.UpdateAdminContentDataWithRevision(ctx, mdb.UpdateAdminContentDataWithRevisionParams{
		AdminRouteID:       p.AdminRouteID,
		ParentID:           p.ParentID,
		FirstChildID:       p.FirstChildID,
		NextSiblingID:      p.NextSiblingID,
		PrevSiblingID:      p.PrevSiblingID,
		AdminDatatypeID:    p.AdminDatatypeID,
		AuthorID:           p.AuthorID,
		Status:             p.Status,
		DateCreated:        p.DateCreated,
		DateModified:       p.DateModified,
		AdminContentDataID: p.AdminContentDataID,
		Revision:           p.Revision,
	})
}

func (d Database) UpdateAdminContentDataSchedule(ctx context.Context, p UpdateAdminContentDataScheduleParams) error {
	queries := mdb.New(d.Connection)
	return queries.UpdateAdminContentDataSchedule(ctx, mdb.UpdateAdminContentDataScheduleParams{
		PublishAt:          p.PublishAt,
		DateModified:       p.DateModified,
		AdminContentDataID: p.AdminContentDataID,
	})
}

func (d Database) ClearAdminContentDataSchedule(ctx context.Context, p ClearAdminContentDataScheduleParams) error {
	queries := mdb.New(d.Connection)
	return queries.ClearAdminContentDataSchedule(ctx, mdb.ClearAdminContentDataScheduleParams{
		DateModified:       p.DateModified,
		AdminContentDataID: p.AdminContentDataID,
	})
}

func (d Database) ListAdminContentDataDueForPublish(now types.Timestamp) (*[]AdminContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentDataDueForPublish(d.Context, mdb.ListAdminContentDataDueForPublishParams{PublishAt: now})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content data due for publish: %w", err)
	}
	res := make([]AdminContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminContentData(v))
	}
	return &res, nil
}

// ===== MySQL =====

func (d MysqlDatabase) UpdateAdminContentDataPublishMeta(ctx context.Context, p UpdateAdminContentDataPublishMetaParams) error {
	queries := mdbm.New(d.Connection)
	return queries.UpdateAdminContentDataPublishMeta(ctx, mdbm.UpdateAdminContentDataPublishMetaParams{
		Status:             p.Status,
		PublishedAt:        p.PublishedAt,
		PublishedBy:        p.PublishedBy,
		DateModified:       p.DateModified,
		AdminContentDataID: p.AdminContentDataID,
	})
}

func (d MysqlDatabase) UpdateAdminContentDataWithRevision(ctx context.Context, p UpdateAdminContentDataWithRevisionParams) error {
	queries := mdbm.New(d.Connection)
	return queries.UpdateAdminContentDataWithRevision(ctx, mdbm.UpdateAdminContentDataWithRevisionParams{
		AdminRouteID:       p.AdminRouteID,
		ParentID:           p.ParentID,
		FirstChildID:       p.FirstChildID,
		NextSiblingID:      p.NextSiblingID,
		PrevSiblingID:      p.PrevSiblingID,
		AdminDatatypeID:    p.AdminDatatypeID,
		AuthorID:           p.AuthorID,
		Status:             p.Status,
		DateCreated:        p.DateCreated,
		DateModified:       p.DateModified,
		AdminContentDataID: p.AdminContentDataID,
		Revision:           int32(p.Revision),
	})
}

func (d MysqlDatabase) UpdateAdminContentDataSchedule(ctx context.Context, p UpdateAdminContentDataScheduleParams) error {
	queries := mdbm.New(d.Connection)
	return queries.UpdateAdminContentDataSchedule(ctx, mdbm.UpdateAdminContentDataScheduleParams{
		PublishAt:          p.PublishAt,
		DateModified:       p.DateModified,
		AdminContentDataID: p.AdminContentDataID,
	})
}

func (d MysqlDatabase) ClearAdminContentDataSchedule(ctx context.Context, p ClearAdminContentDataScheduleParams) error {
	queries := mdbm.New(d.Connection)
	return queries.ClearAdminContentDataSchedule(ctx, mdbm.ClearAdminContentDataScheduleParams{
		DateModified:       p.DateModified,
		AdminContentDataID: p.AdminContentDataID,
	})
}

func (d MysqlDatabase) ListAdminContentDataDueForPublish(now types.Timestamp) (*[]AdminContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentDataDueForPublish(d.Context, mdbm.ListAdminContentDataDueForPublishParams{PublishAt: now})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content data due for publish: %w", err)
	}
	res := make([]AdminContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminContentData(v))
	}
	return &res, nil
}

// ===== PostgreSQL =====

func (d PsqlDatabase) UpdateAdminContentDataPublishMeta(ctx context.Context, p UpdateAdminContentDataPublishMetaParams) error {
	queries := mdbp.New(d.Connection)
	return queries.UpdateAdminContentDataPublishMeta(ctx, mdbp.UpdateAdminContentDataPublishMetaParams{
		Status:             p.Status,
		PublishedAt:        p.PublishedAt,
		PublishedBy:        p.PublishedBy,
		DateModified:       p.DateModified,
		AdminContentDataID: p.AdminContentDataID,
	})
}

func (d PsqlDatabase) UpdateAdminContentDataWithRevision(ctx context.Context, p UpdateAdminContentDataWithRevisionParams) error {
	queries := mdbp.New(d.Connection)
	return queries.UpdateAdminContentDataWithRevision(ctx, mdbp.UpdateAdminContentDataWithRevisionParams{
		AdminRouteID:       p.AdminRouteID,
		ParentID:           p.ParentID,
		FirstChildID:       p.FirstChildID,
		NextSiblingID:      p.NextSiblingID,
		PrevSiblingID:      p.PrevSiblingID,
		AdminDatatypeID:    p.AdminDatatypeID,
		AuthorID:           p.AuthorID,
		Status:             p.Status,
		DateCreated:        p.DateCreated,
		DateModified:       p.DateModified,
		AdminContentDataID: p.AdminContentDataID,
		Revision:           int32(p.Revision),
	})
}

func (d PsqlDatabase) UpdateAdminContentDataSchedule(ctx context.Context, p UpdateAdminContentDataScheduleParams) error {
	queries := mdbp.New(d.Connection)
	return queries.UpdateAdminContentDataSchedule(ctx, mdbp.UpdateAdminContentDataScheduleParams{
		PublishAt:          p.PublishAt,
		DateModified:       p.DateModified,
		AdminContentDataID: p.AdminContentDataID,
	})
}

func (d PsqlDatabase) ClearAdminContentDataSchedule(ctx context.Context, p ClearAdminContentDataScheduleParams) error {
	queries := mdbp.New(d.Connection)
	return queries.ClearAdminContentDataSchedule(ctx, mdbp.ClearAdminContentDataScheduleParams{
		DateModified:       p.DateModified,
		AdminContentDataID: p.AdminContentDataID,
	})
}

func (d PsqlDatabase) ListAdminContentDataDueForPublish(now types.Timestamp) (*[]AdminContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentDataDueForPublish(d.Context, mdbp.ListAdminContentDataDueForPublishParams{PublishAt: now})
	if err != nil {
		return nil, fmt.Errorf("failed to list admin content data due for publish: %w", err)
	}
	res := make([]AdminContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminContentData(v))
	}
	return &res, nil
}

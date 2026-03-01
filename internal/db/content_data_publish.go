package db

import (
	"context"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

// UpdateContentDataPublishMetaParams contains parameters for updating publish metadata.
type UpdateContentDataPublishMetaParams struct {
	Status        types.ContentStatus  `json:"status"`
	PublishedAt   types.Timestamp      `json:"published_at"`
	PublishedBy   types.NullableUserID `json:"published_by"`
	DateModified  types.Timestamp      `json:"date_modified"`
	ContentDataID types.ContentID      `json:"content_data_id"`
}

// UpdateContentDataWithRevisionParams contains parameters for updating content data with a new revision.
type UpdateContentDataWithRevisionParams struct {
	RouteID       types.NullableRouteID    `json:"route_id"`
	ParentID      types.NullableContentID  `json:"parent_id"`
	FirstChildID  types.NullableContentID  `json:"first_child_id"`
	NextSiblingID types.NullableContentID  `json:"next_sibling_id"`
	PrevSiblingID types.NullableContentID  `json:"prev_sibling_id"`
	DatatypeID    types.NullableDatatypeID `json:"datatype_id"`
	AuthorID      types.UserID             `json:"author_id"`
	Status        types.ContentStatus      `json:"status"`
	DateCreated   types.Timestamp          `json:"date_created"`
	DateModified  types.Timestamp          `json:"date_modified"`
	ContentDataID types.ContentID          `json:"content_data_id"`
	Revision      int64                    `json:"revision"`
}

// UpdateContentDataScheduleParams contains parameters for scheduling content publication.
type UpdateContentDataScheduleParams struct {
	PublishAt     types.Timestamp `json:"publish_at"`
	DateModified  types.Timestamp `json:"date_modified"`
	ContentDataID types.ContentID `json:"content_data_id"`
}

// ClearContentDataScheduleParams contains parameters for clearing a scheduled publication.
type ClearContentDataScheduleParams struct {
	DateModified  types.Timestamp `json:"date_modified"`
	ContentDataID types.ContentID `json:"content_data_id"`
}

// ===== SQLite =====

func (d Database) UpdateContentDataPublishMeta(ctx context.Context, p UpdateContentDataPublishMetaParams) error {
	queries := mdb.New(d.Connection)
	return queries.UpdateContentDataPublishMeta(ctx, mdb.UpdateContentDataPublishMetaParams{
		Status:        p.Status,
		PublishedAt:   p.PublishedAt,
		PublishedBy:   p.PublishedBy,
		DateModified:  p.DateModified,
		ContentDataID: p.ContentDataID,
	})
}

func (d Database) UpdateContentDataWithRevision(ctx context.Context, p UpdateContentDataWithRevisionParams) error {
	queries := mdb.New(d.Connection)
	return queries.UpdateContentDataWithRevision(ctx, mdb.UpdateContentDataWithRevisionParams{
		RouteID:       p.RouteID,
		ParentID:      p.ParentID,
		FirstChildID:  p.FirstChildID,
		NextSiblingID: p.NextSiblingID,
		PrevSiblingID: p.PrevSiblingID,
		DatatypeID:    p.DatatypeID,
		AuthorID:      p.AuthorID,
		Status:        p.Status,
		DateCreated:   p.DateCreated,
		DateModified:  p.DateModified,
		ContentDataID: p.ContentDataID,
		Revision:      p.Revision,
	})
}

func (d Database) UpdateContentDataSchedule(ctx context.Context, p UpdateContentDataScheduleParams) error {
	queries := mdb.New(d.Connection)
	return queries.UpdateContentDataSchedule(ctx, mdb.UpdateContentDataScheduleParams{
		PublishAt:     p.PublishAt,
		DateModified:  p.DateModified,
		ContentDataID: p.ContentDataID,
	})
}

func (d Database) ClearContentDataSchedule(ctx context.Context, p ClearContentDataScheduleParams) error {
	queries := mdb.New(d.Connection)
	return queries.ClearContentDataSchedule(ctx, mdb.ClearContentDataScheduleParams{
		DateModified:  p.DateModified,
		ContentDataID: p.ContentDataID,
	})
}

func (d Database) ListContentDataDueForPublish(now types.Timestamp) (*[]ContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentDataDueForPublish(d.Context, mdb.ListContentDataDueForPublishParams{PublishAt: now})
	if err != nil {
		return nil, fmt.Errorf("failed to list content data due for publish: %w", err)
	}
	res := make([]ContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapContentData(v))
	}
	return &res, nil
}

// ===== MySQL =====

func (d MysqlDatabase) UpdateContentDataPublishMeta(ctx context.Context, p UpdateContentDataPublishMetaParams) error {
	queries := mdbm.New(d.Connection)
	return queries.UpdateContentDataPublishMeta(ctx, mdbm.UpdateContentDataPublishMetaParams{
		Status:        p.Status,
		PublishedAt:   p.PublishedAt,
		PublishedBy:   p.PublishedBy,
		DateModified:  p.DateModified,
		ContentDataID: p.ContentDataID,
	})
}

func (d MysqlDatabase) UpdateContentDataWithRevision(ctx context.Context, p UpdateContentDataWithRevisionParams) error {
	queries := mdbm.New(d.Connection)
	return queries.UpdateContentDataWithRevision(ctx, mdbm.UpdateContentDataWithRevisionParams{
		RouteID:       p.RouteID,
		ParentID:      p.ParentID,
		FirstChildID:  p.FirstChildID,
		NextSiblingID: p.NextSiblingID,
		PrevSiblingID: p.PrevSiblingID,
		DatatypeID:    p.DatatypeID,
		AuthorID:      p.AuthorID,
		Status:        p.Status,
		DateCreated:   p.DateCreated,
		DateModified:  p.DateModified,
		ContentDataID: p.ContentDataID,
		Revision:      int32(p.Revision),
	})
}

func (d MysqlDatabase) UpdateContentDataSchedule(ctx context.Context, p UpdateContentDataScheduleParams) error {
	queries := mdbm.New(d.Connection)
	return queries.UpdateContentDataSchedule(ctx, mdbm.UpdateContentDataScheduleParams{
		PublishAt:     p.PublishAt,
		DateModified:  p.DateModified,
		ContentDataID: p.ContentDataID,
	})
}

func (d MysqlDatabase) ClearContentDataSchedule(ctx context.Context, p ClearContentDataScheduleParams) error {
	queries := mdbm.New(d.Connection)
	return queries.ClearContentDataSchedule(ctx, mdbm.ClearContentDataScheduleParams{
		DateModified:  p.DateModified,
		ContentDataID: p.ContentDataID,
	})
}

func (d MysqlDatabase) ListContentDataDueForPublish(now types.Timestamp) (*[]ContentData, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentDataDueForPublish(d.Context, mdbm.ListContentDataDueForPublishParams{PublishAt: now})
	if err != nil {
		return nil, fmt.Errorf("failed to list content data due for publish: %w", err)
	}
	res := make([]ContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapContentData(v))
	}
	return &res, nil
}

// ===== PostgreSQL =====

func (d PsqlDatabase) UpdateContentDataPublishMeta(ctx context.Context, p UpdateContentDataPublishMetaParams) error {
	queries := mdbp.New(d.Connection)
	return queries.UpdateContentDataPublishMeta(ctx, mdbp.UpdateContentDataPublishMetaParams{
		Status:        p.Status,
		PublishedAt:   p.PublishedAt,
		PublishedBy:   p.PublishedBy,
		DateModified:  p.DateModified,
		ContentDataID: p.ContentDataID,
	})
}

func (d PsqlDatabase) UpdateContentDataWithRevision(ctx context.Context, p UpdateContentDataWithRevisionParams) error {
	queries := mdbp.New(d.Connection)
	return queries.UpdateContentDataWithRevision(ctx, mdbp.UpdateContentDataWithRevisionParams{
		RouteID:       p.RouteID,
		ParentID:      p.ParentID,
		FirstChildID:  p.FirstChildID,
		NextSiblingID: p.NextSiblingID,
		PrevSiblingID: p.PrevSiblingID,
		DatatypeID:    p.DatatypeID,
		AuthorID:      p.AuthorID,
		Status:        p.Status,
		DateCreated:   p.DateCreated,
		DateModified:  p.DateModified,
		ContentDataID: p.ContentDataID,
		Revision:      int32(p.Revision),
	})
}

func (d PsqlDatabase) UpdateContentDataSchedule(ctx context.Context, p UpdateContentDataScheduleParams) error {
	queries := mdbp.New(d.Connection)
	return queries.UpdateContentDataSchedule(ctx, mdbp.UpdateContentDataScheduleParams{
		PublishAt:     p.PublishAt,
		DateModified:  p.DateModified,
		ContentDataID: p.ContentDataID,
	})
}

func (d PsqlDatabase) ClearContentDataSchedule(ctx context.Context, p ClearContentDataScheduleParams) error {
	queries := mdbp.New(d.Connection)
	return queries.ClearContentDataSchedule(ctx, mdbp.ClearContentDataScheduleParams{
		DateModified:  p.DateModified,
		ContentDataID: p.ContentDataID,
	})
}

func (d PsqlDatabase) ListContentDataDueForPublish(now types.Timestamp) (*[]ContentData, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListContentDataDueForPublish(d.Context, mdbp.ListContentDataDueForPublishParams{PublishAt: now})
	if err != nil {
		return nil, fmt.Errorf("failed to list content data due for publish: %w", err)
	}
	res := make([]ContentData, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapContentData(v))
	}
	return &res, nil
}

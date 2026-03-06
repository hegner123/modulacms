-- name: DropContentDataTable :exec
DROP TABLE content_data;

-- name: CreateContentDataTable :exec
CREATE TABLE IF NOT EXISTS content_data (
    content_data_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        CONSTRAINT fk_parent_id
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    first_child_id TEXT
        CONSTRAINT fk_first_child_id
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    next_sibling_id TEXT
        CONSTRAINT fk_next_sibling_id
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    prev_sibling_id TEXT
        CONSTRAINT fk_prev_sibling_id
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    route_id TEXT
        CONSTRAINT fk_routes
            REFERENCES routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    datatype_id TEXT
        CONSTRAINT fk_datatypes
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE SET NULL,
    author_id TEXT NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP,
    published_by TEXT
        CONSTRAINT fk_published_by
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    publish_at TIMESTAMP,
    revision INTEGER NOT NULL DEFAULT 0
);

-- name: CountContentData :one
SELECT COUNT(*)
FROM content_data;

-- name: GetContentData :one
SELECT * FROM content_data
WHERE content_data_id = $1 LIMIT 1;

-- name: ListContentData :many
SELECT * FROM content_data
ORDER BY content_data_id;

-- name: ListContentDataByRoute :many
SELECT * FROM content_data
WHERE route_id = $1
ORDER BY content_data_id;

-- name: CreateContentData :one
INSERT INTO content_data (
    content_data_id,
    parent_id,
    first_child_id,
    next_sibling_id,
    prev_sibling_id,
    route_id,
    datatype_id,
    author_id,
    status,
    date_created,
    date_modified
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11
) RETURNING *;

-- name: UpdateContentData :exec
UPDATE content_data
SET route_id = $1,
    parent_id = $2,
    first_child_id = $3,
    next_sibling_id = $4,
    prev_sibling_id = $5,
    datatype_id = $6,
    author_id = $7,
    status = $8,
    date_created = $9,
    date_modified = $10
WHERE content_data_id = $11;

-- name: DeleteContentData :exec
DELETE FROM content_data
WHERE content_data_id = $1;

-- name: ListContentDataPaginated :many
SELECT * FROM content_data
ORDER BY content_data_id
LIMIT $1 OFFSET $2;

-- name: ListContentDataByRoutePaginated :many
SELECT * FROM content_data
WHERE route_id = $1
ORDER BY content_data_id
LIMIT $2 OFFSET $3;

-- name: GetContentDataDescendants :many
WITH RECURSIVE tree AS (
    SELECT cd1.content_data_id AS cid FROM content_data cd1 WHERE cd1.content_data_id = $1
    UNION ALL
    SELECT cd2.content_data_id FROM content_data cd2
    INNER JOIN tree t ON cd2.parent_id = t.cid
)
SELECT cd.* FROM content_data cd
INNER JOIN tree t ON cd.content_data_id = t.cid;

-- name: ListContentDataTopLevelPaginated :many
SELECT cd.*, u.name AS author_name, COALESCE(r.slug, '') AS route_slug, COALESCE(r.title, '') AS route_title, COALESCE(dt.label, '') AS datatype_label, COALESCE(dt.type, '') AS datatype_type FROM content_data cd
LEFT JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
LEFT JOIN users u ON cd.author_id = u.user_id
LEFT JOIN routes r ON cd.route_id = r.route_id
WHERE dt.type IN ('_root', '_global')
ORDER BY cd.content_data_id
LIMIT $1 OFFSET $2;

-- name: CountContentDataTopLevel :one
SELECT COUNT(*) FROM content_data cd
LEFT JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE dt.type IN ('_root', '_global');

-- name: UpdateContentDataPublishMeta :exec
UPDATE content_data
SET status = $1,
    published_at = $2,
    published_by = $3,
    revision = revision + 1,
    date_modified = $4
WHERE content_data_id = $5;

-- name: UpdateContentDataWithRevision :exec
UPDATE content_data
SET route_id = $1,
    parent_id = $2,
    first_child_id = $3,
    next_sibling_id = $4,
    prev_sibling_id = $5,
    datatype_id = $6,
    author_id = $7,
    status = $8,
    date_created = $9,
    date_modified = $10
WHERE content_data_id = $11 AND revision = $12;

-- name: UpdateContentDataSchedule :exec
UPDATE content_data
SET publish_at = $1,
    date_modified = $2
WHERE content_data_id = $3;

-- name: ClearContentDataSchedule :exec
UPDATE content_data
SET publish_at = NULL,
    date_modified = $1
WHERE content_data_id = $2;

-- name: ListContentDataTopLevelPaginatedByStatus :many
SELECT cd.*, u.name AS author_name, COALESCE(r.slug, '') AS route_slug, COALESCE(r.title, '') AS route_title, COALESCE(dt.label, '') AS datatype_label, COALESCE(dt.type, '') AS datatype_type FROM content_data cd
LEFT JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
LEFT JOIN users u ON cd.author_id = u.user_id
LEFT JOIN routes r ON cd.route_id = r.route_id
WHERE dt.type IN ('_root', '_global') AND cd.status = $1
ORDER BY cd.content_data_id
LIMIT $2 OFFSET $3;

-- name: CountContentDataTopLevelByStatus :one
SELECT COUNT(*) FROM content_data cd
LEFT JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE dt.type IN ('_root', '_global') AND cd.status = $1;

-- name: ListContentDataGlobal :many
SELECT cd.* FROM content_data cd
LEFT JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE dt.type = '_global' AND cd.parent_id IS NULL
ORDER BY cd.content_data_id;

-- name: ListContentDataDueForPublish :many
SELECT * FROM content_data
WHERE publish_at IS NOT NULL AND publish_at <= $1 AND status = 'draft';

-- name: ListContentDataByDatatypeID :many
SELECT * FROM content_data
WHERE datatype_id = $1;

-- name: ReassignContentDataAuthor :exec
UPDATE content_data SET author_id = $1 WHERE author_id = $2;

-- name: CountContentDataByAuthor :one
SELECT COUNT(*) FROM content_data WHERE author_id = $1;

-- name: DropContentDataTable :exec
DROP TABLE content_data;

-- name: CreateContentDataTable :exec
CREATE TABLE IF NOT EXISTS content_data (
    content_data_id TEXT PRIMARY KEY NOT NULL CHECK (length(content_data_id) = 26),
    parent_id TEXT
        REFERENCES content_data
            ON DELETE SET NULL,
    first_child_id TEXT
        REFERENCES content_data
            ON DELETE SET NULL,
    next_sibling_id TEXT
        REFERENCES content_data
            ON DELETE SET NULL,
    prev_sibling_id TEXT
        REFERENCES content_data
            ON DELETE SET NULL,
    route_id TEXT
        REFERENCES routes
            ON DELETE SET NULL,
    datatype_id TEXT NOT NULL
        REFERENCES datatypes
            ON DELETE SET NULL,
    author_id TEXT NOT NULL
        REFERENCES users
            ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    published_at TEXT,
    published_by TEXT REFERENCES users(user_id) ON DELETE SET NULL,
    publish_at TEXT,
    revision INTEGER NOT NULL DEFAULT 0,

    FOREIGN KEY (parent_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (first_child_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (next_sibling_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (prev_sibling_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (route_id) REFERENCES routes(route_id) ON DELETE SET NULL,
    FOREIGN KEY (datatype_id) REFERENCES datatypes(datatype_id) ON DELETE RESTRICT,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE SET NULL
);



-- name: CountContentData :one
SELECT COUNT(*)
FROM content_data;

-- name: GetContentData :one
SELECT * FROM content_data
WHERE content_data_id = ? LIMIT 1;

-- name: ListContentData :many
SELECT * FROM content_data
ORDER BY content_data_id;

-- name: ListContentDataByRoute :many
SELECT * FROM content_data
WHERE route_id = ?
ORDER BY content_data_id;

-- name: CreateContentData :one
INSERT INTO content_data (
    content_data_id,
    route_id,
    parent_id,
    first_child_id,
    next_sibling_id,
    prev_sibling_id,
    datatype_id,
    author_id,
    status,
    date_created,
    date_modified
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
) RETURNING *;

-- name: UpdateContentData :exec
UPDATE content_data
SET route_id = ?,
    parent_id = ?,
    first_child_id = ?,
    next_sibling_id = ?,
    prev_sibling_id = ?,
    datatype_id = ?,
    author_id = ?,
    status = ?,
    date_created = ?,
    date_modified = ?
WHERE content_data_id = ?;

-- name: DeleteContentData :exec
DELETE FROM content_data
WHERE content_data_id = ?;

-- name: ListContentDataPaginated :many
SELECT * FROM content_data
ORDER BY content_data_id
LIMIT ? OFFSET ?;

-- name: ListContentDataByRoutePaginated :many
SELECT * FROM content_data
WHERE route_id = ?
ORDER BY content_data_id
LIMIT ? OFFSET ?;

-- name: GetContentDataDescendants :many
WITH RECURSIVE tree AS (
    SELECT cd1.content_data_id AS cid FROM content_data cd1 WHERE cd1.content_data_id = ?
    UNION ALL
    SELECT cd2.content_data_id FROM content_data cd2
    INNER JOIN tree t ON cd2.parent_id = t.cid
)
SELECT cd.* FROM content_data cd
INNER JOIN tree t ON cd.content_data_id = t.cid;

-- name: ListContentDataTopLevelPaginated :many
SELECT cd.*, u.name AS author_name, COALESCE(r.slug, '') AS route_slug, COALESCE(r.title, '') AS route_title, COALESCE(dt.label, '') AS datatype_label FROM content_data cd
LEFT JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
LEFT JOIN users u ON cd.author_id = u.user_id
LEFT JOIN routes r ON cd.route_id = r.route_id
WHERE cd.route_id IS NOT NULL OR dt.type = '_root'
ORDER BY cd.content_data_id
LIMIT ? OFFSET ?;

-- name: CountContentDataTopLevel :one
SELECT COUNT(*) FROM content_data cd
LEFT JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE cd.route_id IS NOT NULL OR dt.type = '_root';

-- name: UpdateContentDataPublishMeta :exec
UPDATE content_data
SET status = ?,
    published_at = ?,
    published_by = ?,
    revision = revision + 1,
    date_modified = ?
WHERE content_data_id = ?;

-- name: UpdateContentDataWithRevision :exec
UPDATE content_data
SET route_id = ?,
    parent_id = ?,
    first_child_id = ?,
    next_sibling_id = ?,
    prev_sibling_id = ?,
    datatype_id = ?,
    author_id = ?,
    status = ?,
    date_created = ?,
    date_modified = ?
WHERE content_data_id = ? AND revision = ?;

-- name: UpdateContentDataSchedule :exec
UPDATE content_data
SET publish_at = ?,
    date_modified = ?
WHERE content_data_id = ?;

-- name: ClearContentDataSchedule :exec
UPDATE content_data
SET publish_at = NULL,
    date_modified = ?
WHERE content_data_id = ?;

-- name: ListContentDataTopLevelPaginatedByStatus :many
SELECT cd.*, u.name AS author_name, COALESCE(r.slug, '') AS route_slug, COALESCE(r.title, '') AS route_title, COALESCE(dt.label, '') AS datatype_label FROM content_data cd
LEFT JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
LEFT JOIN users u ON cd.author_id = u.user_id
LEFT JOIN routes r ON cd.route_id = r.route_id
WHERE (cd.route_id IS NOT NULL OR dt.type = '_root') AND cd.status = ?
ORDER BY cd.content_data_id
LIMIT ? OFFSET ?;

-- name: CountContentDataTopLevelByStatus :one
SELECT COUNT(*) FROM content_data cd
LEFT JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE (cd.route_id IS NOT NULL OR dt.type = '_root') AND cd.status = ?;

-- name: ListContentDataDueForPublish :many
SELECT * FROM content_data
WHERE publish_at IS NOT NULL AND publish_at <= ? AND status = 'draft';

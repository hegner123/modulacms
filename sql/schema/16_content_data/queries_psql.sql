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
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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

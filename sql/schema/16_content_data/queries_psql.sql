-- name: DropContentDataTable :exec
DROP TABLE content_data;

-- name: CreateContentDataTable :exec
CREATE TABLE content_data (
    content_data_id SERIAL
        PRIMARY KEY,
    parent_id INTEGER
        CONSTRAINT fk_parent_id
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    route_id INTEGER
        CONSTRAINT fk_routes
            REFERENCES routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    datatype_id INTEGER
        CONSTRAINT fk_datatypes
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE SET NULL,
    author_id INTEGER NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT
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
    parent_id,
    route_id,
    datatype_id,
    author_id,
    date_created,
    date_modified,
    history
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
) RETURNING *;

-- name: UpdateContentData :exec
UPDATE content_data
SET route_id = $1,
    parent_id = $2,
    datatype_id =$3,
    author_id = $4,
    date_created = $5,
    date_modified = $6,
    history = $7
WHERE content_data_id = $8;

-- name: DeleteContentData :exec
DELETE FROM content_data
WHERE content_data_id = $1;

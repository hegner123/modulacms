-- name: DropContentFieldTable :exec
DROP TABLE content_fields;

-- name: CreateContentFieldTable :exec
CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id TEXT PRIMARY KEY NOT NULL,
    route_id TEXT
        CONSTRAINT fk_route_id
            REFERENCES routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    content_data_id TEXT NOT NULL
        CONSTRAINT fk_content_data
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_id TEXT NOT NULL
        CONSTRAINT fk_fields_field
            REFERENCES fields
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_value TEXT NOT NULL,
    author_id TEXT NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- name: CountContentField :one
SELECT COUNT(*)
FROM content_fields;

-- name: GetContentField :one
SELECT * FROM content_fields
WHERE content_field_id = $1 LIMIT 1;

-- name: ListContentFields :many
SELECT * FROM content_fields
ORDER BY content_field_id;

-- name: ListContentFieldsByRoute :many
SELECT * FROM content_fields
WHERE route_id = $1
ORDER BY content_field_id;

-- name: CreateContentField :one
INSERT INTO content_fields (
    content_field_id,
    route_id,
    content_data_id,
    field_id,
    field_value,
    author_id, 

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
    $8
) RETURNING *;

-- name: UpdateContentField :exec
UPDATE content_fields
SET  content_field_id = $1,
    route_id = $2,
    content_data_id = $3,
    field_id = $4,
    field_value = $5,
    author_id = $6,
    date_created = $7,
    date_modified = $8
WHERE content_field_id = $9;

-- name: DeleteContentField :exec
DELETE FROM content_fields
WHERE content_field_id = $1;

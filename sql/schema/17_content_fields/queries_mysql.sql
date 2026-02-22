-- name: DropContentFieldTable :exec
DROP TABLE content_fields;

-- name: CreateContentFieldTable :exec
CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id VARCHAR(26) PRIMARY KEY NOT NULL,
    route_id VARCHAR(26) NULL,
    content_data_id VARCHAR(26) NOT NULL,
    field_id VARCHAR(26) NOT NULL,
    field_value TEXT NOT NULL,
    author_id VARCHAR(26) NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_content_field_content_data
        FOREIGN KEY (content_data_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_content_field_fields
        FOREIGN KEY (field_id) REFERENCES fields (field_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_content_field_route_id
        FOREIGN KEY (route_id) REFERENCES routes (route_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_field_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

-- name: CountContentField :one
SELECT COUNT(*)
FROM content_fields;

-- name: GetContentField :one
SELECT * FROM content_fields
WHERE content_field_id = ? LIMIT 1;

-- name: ListContentFields :many
SELECT * FROM content_fields
ORDER BY content_field_id;

-- name: ListContentFieldsByRoute :many
SELECT * FROM content_fields
WHERE route_id = ?
ORDER BY content_field_id;

-- name: ListContentFieldsByContentData :many
SELECT * FROM content_fields
WHERE content_data_id = ?
ORDER BY content_field_id;

-- name: CreateContentField :exec
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
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
);

-- name: UpdateContentField :exec
UPDATE content_fields
SET route_id = ?,
    content_data_id = ?,
    field_id = ?,
    field_value = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
WHERE content_field_id = ?;

-- name: DeleteContentField :exec
DELETE FROM content_fields
WHERE content_field_id = ?;

-- name: ListContentFieldsPaginated :many
SELECT * FROM content_fields
ORDER BY content_field_id
LIMIT ? OFFSET ?;

-- name: ListContentFieldsByRoutePaginated :many
SELECT * FROM content_fields
WHERE route_id = ?
ORDER BY content_field_id
LIMIT ? OFFSET ?;

-- name: ListContentFieldsByContentDataPaginated :many
SELECT * FROM content_fields
WHERE content_data_id = ?
ORDER BY content_field_id
LIMIT ? OFFSET ?;

-- name: DropFieldTable :exec
DROP TABLE fields;

-- name: CreateFieldTable :exec
CREATE TABLE IF NOT EXISTS fields (
    field_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    label VARCHAR(255) DEFAULT 'unlabeled' NOT NULL,
    data TEXT NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('text', 'textarea', 'number', 'date', 'datetime', 'boolean', 'select', 'media', 'relation', 'json', 'richtext', 'slug', 'email', 'url')),
    author_id VARCHAR(26) NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_fields_datatypes
        FOREIGN KEY (parent_id) REFERENCES datatypes (datatype_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_fields_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);

-- name: CountField :one
SELECT COUNT(*)
FROM fields ;

-- name: GetField :one
SELECT * FROM fields
WHERE field_id = ? LIMIT 1;

-- name: ListField :many
SELECT * FROM fields
ORDER BY field_id;

-- name: ListFieldByDatatypeID :many
SELECT * FROM fields
WHERE parent_id = ?
ORDER BY field_id;

-- name: CreateField :exec
INSERT INTO fields  (
    field_id,
    parent_id,
    label,
    data,
    validation,
    ui_config,
    type,
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
    ?,
    ?,
    ?
);

-- name: UpdateField :exec
UPDATE fields
set
    parent_id = ?,
    label = ?,
    data = ?,
    validation = ?,
    ui_config = ?,
    type = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
    WHERE field_id = ?;

-- name: DeleteField :exec
DELETE FROM fields
WHERE field_id = ?;

-- name: ListFieldPaginated :many
SELECT * FROM fields
ORDER BY field_id
LIMIT ? OFFSET ?;

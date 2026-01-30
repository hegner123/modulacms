-- name: DropFieldTable :exec
DROP TABLE fields;

-- name: CreateFieldTable :exec
CREATE TABLE IF NOT EXISTS fields (
    field_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    label VARCHAR(255) DEFAULT 'unlabeled' NOT NULL,
    data TEXT NOT NULL,
    type TEXT NOT NULL,
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
    ?
);

-- name: GetLastField :one
SELECT * FROM fields WHERE field_id = LAST_INSERT_ID();

-- name: UpdateField :exec
UPDATE fields 
set 
    parent_id = ?,
    label = ?,
    data = ?,
    type = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
    WHERE field_id = ?;

-- name: DeleteField :exec
DELETE FROM fields 
WHERE field_id = ?;


-- name: CreateDatatypeTable :exec
CREATE TABLE IF NOT EXISTS datatypes
(
    datatype_id   INTEGER
        primary key,
    parent_id     INTEGER default NULL
        references datatypes
            on update cascade on delete set default,
    label         TEXT                     not null,
    type          TEXT                     not null,
    author        TEXT    default "system" not null
        references users (username)
            on update cascade on delete set default,
    author_id     INTEGER default 1        not null
        references users (user_id)
            on update cascade on delete set default,
    history TEXT,
    date_created  TEXT    default CURRENT_TIMESTAMP,
    date_modified TEXT    default CURRENT_TIMESTAMP
);
-- name: GetDatatype :one
SELECT * FROM datatypes
WHERE datatype_id = ? LIMIT 1;

-- name: CountDatatype :one
SELECT COUNT(*)
FROM datatypes;


-- name: ListDatatype :many
SELECT * FROM datatypes
ORDER BY datatype_id;


-- name: CreateDatatype :one
INSERT INTO datatypes (
    parent_id,
    label,
    type,
    author,
    author_id,
    history,
    date_created,
    date_modified
    ) VALUES (
  ?,?,?,?,?,?,?,?
    ) RETURNING *;


-- name: UpdateDatatype :exec
UPDATE datatypes
set parent_id = ?,
    label = ?,
    type = ?,
    author = ?,
    author_id = ?,
    history = ?,
    date_created = ?,
    date_modified = ?
    WHERE datatype_id = ?
    RETURNING *;

-- name: DeleteDatatype :exec
DELETE FROM datatypes
WHERE datatype_id = ?;




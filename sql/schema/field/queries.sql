-- name: CreateFieldTable :exec
CREATE TABLE IF NOT EXISTS fields
(
    field_id      INTEGER
        primary key,
    parent_id     INTEGER default NULL
        references datatypes
            on update cascade on delete set default,
    label         TEXT    default "unlabeled" not null,
    data          TEXT                        not null,
    type          TEXT                        not null,
    author        TEXT    default "system"    not null
        references users (username)
            on update cascade on delete set default,
    author_id     INTEGER default 1           not null
        references users (user_id)
            on update cascade on delete set default,
    date_created  TEXT    default CURRENT_TIMESTAMP,
    date_modified TEXT    default CURRENT_TIMESTAMP,
    history TEXT
);
-- name: GetField :one
SELECT * FROM fields 
WHERE field_id = ? LIMIT 1;

-- name: CountField :one
SELECT COUNT(*)
FROM fields ;

-- name: ListField :many
SELECT * FROM fields 
ORDER BY field_id;

-- name: CreateField :one
INSERT INTO fields  (
    parent_id,
    label,
    data,
    type,
    author,
    author_id,
    history,
    date_created,
    date_modified
    ) VALUES (
?,?,?,?,?,?,?,?,?
    ) RETURNING *;


-- name: UpdateField :exec
UPDATE fields 
set parent_id = ?,
    label = ?,
    data = ?,
    type = ?,
    author = ?,
    author_id = ?,
    history =?,
    date_created = ?,
    date_modified = ?
    WHERE field_id = ?
    RETURNING *;

-- name: DeleteField :exec
DELETE FROM fields 
WHERE field_id = ?;


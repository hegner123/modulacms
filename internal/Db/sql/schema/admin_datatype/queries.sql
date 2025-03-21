-- name: CreateAdminDatatypeTable :exec
CREATE TABLE IF NOT EXISTS admin_datatypes(
    admin_datatype_id    INTEGER
        primary key,
    parent_id      INTEGER default NULL
        references admin_datatypes
            on update cascade on delete set default,
    label          TEXT    not null,
    type           TEXT    not null,
    author         TEXT    not null
        references users (username)
            on update cascade on delete set default,
    author_id      INTEGER not null
        references users (user_id)
            on update cascade on delete set default,
    date_created   TEXT    default CURRENT_TIMESTAMP,
    date_modified  TEXT    default CURRENT_TIMESTAMP,
    history        TEXT
);
-- name: GetAdminDatatype :one
SELECT * FROM admin_datatypes
WHERE admin_datatype_id = ? LIMIT 1;

-- name: CountAdminDatatype :one
SELECT COUNT(*)
FROM admin_datatypes;

-- name: GetAdminDatatypeId :one
SELECT admin_datatype_id FROM admin_datatypes
WHERE admin_datatype_id = ? LIMIT 1;

-- name: ListAdminDatatype :many
SELECT * FROM admin_datatypes
ORDER BY admin_datatype_id ;

-- name: ListAdminDatatypeTree :many
SELECT 
    child.admin_datatype_id AS child_id,
    child.label       AS child_label,
    parent.admin_datatype_id AS parent_id,
    parent.label       AS parent_label
FROM admin_datatypes AS child
LEFT JOIN admin_datatypes AS parent 
    ON child.parent_id = parent.admin_datatype_id;


-- name: GetGlobalAdminDatatypeId :one
SELECT * FROM admin_datatypes
WHERE type = "GLOBALS" LIMIT 1;

-- name: ListAdminDatatypeChildren :many
SELECT * FROM admin_datatypes
WHERE parent_id = ?;

-- name: CreateAdminDatatype :one
INSERT INTO admin_datatypes (
    parent_id,
    label,
    type,
    author,
    author_id,
    date_created,
    date_modified,
    history
    ) VALUES (
?,?,?,?,?,?,?,?
    ) RETURNING *;


-- name: UpdateAdminDatatype :exec
UPDATE admin_datatypes
set parent_id = ?,
    label = ?,
    type = ?,
    author = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?,
    history = ?
    WHERE admin_datatype_id = ?
    RETURNING *;

-- name: DeleteAdminDatatype :exec
DELETE FROM admin_datatypes
WHERE admin_datatype_id = ?;

-- name: ListAdminDatatypeByRouteId :many
SELECT admin_datatype_id, parent_id, label, type, history
FROM admin_datatypes;

-- name: GetRootAdminDtByAdminRtId :one
SELECT admin_datatype_id, parent_id, label, type, history
FROM admin_datatypes
ORDER BY admin_datatype_id;



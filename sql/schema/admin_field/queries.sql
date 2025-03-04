-- name: CreateAdminFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_fields
(
    admin_field_id INTEGER
        primary key,
    admin_route_id INTEGER default 1
        references admin_routes
            on update cascade on delete set default,
    parent_id      INTEGER default NULL
        references admin_datatypes
            on update cascade on delete set default,
    label          TEXT    default "unlabeled" not null,
    data           TEXT    default ""          not null,
    type           TEXT    default "text"      not null,
    author         TEXT    default "system"    not null
        references users (username)
            on update cascade on delete set default,
    author_id      INTEGER default 1           not null
        references users (user_id)
            on update cascade on delete set default,
    date_created   TEXT    default CURRENT_TIMESTAMP,
    date_modified  TEXT    default CURRENT_TIMESTAMP,
    history        TEXT
);
-- name: GetAdminField :one
SELECT * FROM admin_fields
WHERE admin_field_id = ? LIMIT 1;

-- name: CountAdminField :one
SELECT COUNT(*)
FROM admin_fields;

-- name: GetAdminFieldId :one
SELECT admin_field_id FROM admin_fields
WHERE admin_field_id = ? LIMIT 1;

-- name: ListAdminField :many
SELECT * FROM admin_fields
ORDER BY admin_field_id;

-- name: CreateAdminField :one
INSERT INTO admin_fields (
    admin_route_id,
    parent_id,
    label,
    data,
    type,
    author,
    author_id,
    date_created,
    date_modified,
    history
    ) VALUES (
    ?,?,?,?,?,?,?,?,?,?
    ) RETURNING *;


-- name: UpdateAdminField :exec
UPDATE admin_fields
set admin_route_id = ?,
    parent_id = ?,
    label = ?,
    data = ?,
    type = ?,
    author = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?,
    history =?
    WHERE admin_field_id = ?
    RETURNING *;

-- name: DeleteAdminField :exec
DELETE FROM admin_fields
WHERE admin_field_id = ?;

-- name: ListAdminFieldByRouteId :many
SELECT admin_field_id, admin_route_id, parent_id, label, data, type, history
FROM admin_fields
WHERE admin_route_id = ?;


-- name: ListAdminFieldsByDatatypeID :many
SELECT admin_field_id, admin_route_id, parent_id, label, data, type, history
FROM admin_fields
WHERE parent_id = ?;


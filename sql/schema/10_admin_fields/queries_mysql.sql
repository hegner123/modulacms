-- name: DropAdminFieldTable :exec
DROP TABLE admin_fields;

-- name: CreateAdminFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_fields (
    admin_field_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    label VARCHAR(255) DEFAULT 'unlabeled' NOT NULL,
    data TEXT NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type VARCHAR(255) DEFAULT 'text' NOT NULL,
    author_id VARCHAR(26) NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_admin_fields_admin_datatypes
        FOREIGN KEY (parent_id) REFERENCES admin_datatypes (admin_datatype_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_admin_fields_users_user_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);

-- name: CountAdminField :one
SELECT COUNT(*) FROM admin_fields;

-- name: GetAdminField :one
SELECT * FROM admin_fields
WHERE admin_field_id = ? LIMIT 1;

-- name: ListAdminField :many
SELECT * FROM admin_fields
ORDER BY admin_field_id;

-- name: ListAdminFieldByParentID :many
SELECT * FROM admin_fields
WHERE parent_id = ?
ORDER BY admin_field_id;

-- name: CreateAdminField :exec
INSERT INTO admin_fields (
    admin_field_id,
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

-- name: UpdateAdminField :exec
UPDATE admin_fields
SET  parent_id = ?,
    label = ?,
    data = ?,
    validation = ?,
    ui_config = ?,
    type = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
WHERE admin_field_id = ?;

-- name: DeleteAdminField :exec
DELETE FROM admin_fields
WHERE admin_field_id = ?;

-- name: ListAdminFieldPaginated :many
SELECT * FROM admin_fields
ORDER BY admin_field_id
LIMIT ? OFFSET ?;

-- name: ListAdminFieldByParentIDPaginated :many
SELECT * FROM admin_fields
WHERE parent_id = ?
ORDER BY admin_field_id
LIMIT ? OFFSET ?;

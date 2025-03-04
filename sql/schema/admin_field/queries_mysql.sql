-- name: CreateAdminFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_fields (
    admin_field_id INT AUTO_INCREMENT PRIMARY KEY,
    admin_route_id INT NOT NULL DEFAULT 1,
    parent_id INT DEFAULT NULL,
    label VARCHAR(255) NOT NULL DEFAULT 'unlabeled',
    data TEXT NOT NULL, -- MySQL does not allow a default value for TEXT
    type VARCHAR(255) NOT NULL DEFAULT 'text',
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    history TEXT,
    CONSTRAINT fk_admin_fields_admin_routes FOREIGN KEY (admin_route_id)
        REFERENCES admin_routes(admin_route_id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_admin_fields_admin_datatypes FOREIGN KEY (parent_id)
        REFERENCES admin_datatypes(admin_dt_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_admin_fields_users_username FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_admin_fields_users_user_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- name: GetAdminField :one
SELECT * FROM admin_fields
WHERE admin_field_id = ? LIMIT 1;

-- name: CountAdminField :one
SELECT COUNT(*) FROM admin_fields;

-- name: GetAdminFieldId :one
SELECT admin_field_id FROM admin_fields
WHERE admin_field_id = ? LIMIT 1;

-- name: ListAdminField :many
SELECT * FROM admin_fields
ORDER BY admin_field_id;

-- name: CreateAdminField :exec
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
);
-- name: GetLastAdminField :one
SELECT * FROM admin_fields WHERE admin_field_id = LAST_INSERT_ID();

-- name: UpdateAdminField :exec
UPDATE admin_fields
SET admin_route_id = ?,
    parent_id = ?,
    label = ?,
    data = ?,
    type = ?,
    author = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?,
    history = ?
WHERE admin_field_id = ?;
-- Note: MySQL does not support RETURNING *; execute a SELECT query afterward if needed.

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


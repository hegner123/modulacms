-- name: DropAdminContentFieldTable :exec
DROP TABLE admin_content_fields;

-- name: CreateAdminContentFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id VARCHAR(26) PRIMARY KEY NOT NULL,
    admin_route_id VARCHAR(26) NULL,
    admin_content_data_id VARCHAR(26) NOT NULL,
    admin_field_id VARCHAR(26) NOT NULL,
    admin_field_value TEXT NOT NULL,
    author_id VARCHAR(26) NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_admin_content_field_admin_content_data
        FOREIGN KEY (admin_content_data_id) REFERENCES admin_content_data (admin_content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_field_admin_route_id
        FOREIGN KEY (admin_route_id) REFERENCES admin_routes (admin_route_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_admin_content_field_fields
        FOREIGN KEY (admin_field_id) REFERENCES admin_fields (admin_field_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_field_author_users_user_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

-- name: CountAdminContentField :one
SELECT COUNT(*)
FROM admin_content_fields;

-- name: GetAdminContentField :one
SELECT * FROM admin_content_fields
WHERE admin_content_field_id = ? LIMIT 1;

-- name: ListAdminContentFields :many
SELECT * FROM admin_content_fields
ORDER BY admin_content_field_id;

-- name: ListAdminContentFieldsByRoute :many
SELECT * FROM admin_content_fields
WHERE admin_route_id = ?
ORDER BY admin_content_field_id;

-- name: CreateAdminContentField :exec
INSERT INTO admin_content_fields (
    admin_content_field_id,
    admin_route_id,
    admin_content_data_id,
    admin_field_id,
    admin_field_value,
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

-- name: UpdateAdminContentField :exec
UPDATE admin_content_fields
SET admin_route_id=?,
    admin_content_data_id=?,
    admin_field_id=?,
    admin_field_value=?, 
    author_id=?,
    date_created=?,
    date_modified=?
WHERE admin_content_field_id = ?;

-- name: DeleteAdminContentField :exec
DELETE FROM admin_content_fields
WHERE admin_content_field_id = ?;

-- name: ListAdminContentFieldsPaginated :many
SELECT * FROM admin_content_fields
ORDER BY admin_content_field_id
LIMIT ? OFFSET ?;

-- name: ListAdminContentFieldsByRoutePaginated :many
SELECT * FROM admin_content_fields
WHERE admin_route_id = ?
ORDER BY admin_content_field_id
LIMIT ? OFFSET ?;

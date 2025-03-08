-- name: CreateAdminContentFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id INT AUTO_INCREMENT PRIMARY KEY,
    admin_route_id INT,
    admin_content_data_id INT NOT NULL,
    admin_field_id INT NOT NULL,
    admin_field_value TEXT NOT NULL,
    history TEXT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_admin_content_field_admin_content_data FOREIGN KEY (admin_content_data_id)
        REFERENCES admin_content_data(admin_content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_field_admin_route_id FOREIGN KEY (admin_route_id)
        REFERENCES admin_routes(admin_route_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_admin_content_field_fields FOREIGN KEY (field_id)
        REFERENCES admin_fields(field_id)
        ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
-- name: GetAdminContentField :one
SELECT * FROM admin_content_fields
WHERE admin_content_field_id = ? LIMIT 1;

-- name: CountAdminContentField :one
SELECT COUNT(*)
FROM admin_content_fields;

-- name: ListAdminContentFields :many
SELECT * FROM admin_content_fields
ORDER BY admin_content_fields_id;

-- name: CreateAdminContentField :exec
INSERT INTO admin_content_fields (
    admin_content_field_id,
    admin_route_id,
    admin_content_data_id,
    admin_field_id,
    admin_field_value, 
    history,
    date_created, 
    date_modified
    ) VALUES (
  ?,?,?,?,?,?,?,?
    );
-- name: GetLastAdminContentField :one
SELECT * FROM admin_content_fields WHERE admin_content_fields_id = LAST_INSERT_ID();

-- name: UpdateAdminContentField :exec
UPDATE admin_content_fields
set  admin_content_field_id=?,
    admin_route_id=?,
    admin_content_data_id=?,
    admin_field_id=?,
    admin_field_value=?, 
    history=?,
    date_created=?, 
    date_modified=?
    WHERE admin_content_field_id = ?;

-- name: DeleteAdminContentField :exec
DELETE FROM admin_content_fields
WHERE admin_content_field_id = ?;

-- name: ListAdminContentFieldsByRoute :many
SELECT * FROM admin_content_fields
WHERE admin_route_id = ?;

-- name: CreateAdminContentDataTable :exec
CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id INT AUTO_INCREMENT PRIMARY KEY,
    admin_route_id    INT DEFAULT NULL,
    admin_datatype_id INT DEFAULT NULL,
    history TEXT DEFAULT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_admin_content_data_admin_route_id FOREIGN KEY (admin_route_id)
        REFERENCES admin_routes(admin_route_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_admin_datatypes FOREIGN KEY (admin_datatype_id)
        REFERENCES admin_datatypes(admin_datatype_id)
        ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- name: GetAdminContentData :one
SELECT * FROM admin_content_data
WHERE admin_content_data_id = ? LIMIT 1;

-- name: CountAdminContentData :one
SELECT COUNT(*)
FROM admin_content_data;

-- name: ListAdminContentData :many
SELECT * FROM admin_content_data
ORDER BY admin_content_data_id;

-- name: CreateAdminContentData :exec
INSERT INTO admin_content_data (
    admin_route_id,
    admin_datatype_id,
    history,
    date_created,
    date_modified
    ) VALUES (
?,?,?,?,?
    );
-- name: GetLastAdminContentData :one
SELECT * FROM admin_content_data WHERE content_data_id = LAST_INSERT_ID();


-- name: UpdateAdminContentData :exec
UPDATE admin_content_data
set admin_route_id = ?,
    admin_datatype_id = ?,
    history = ?,
    date_created = ?,
    date_modified = ?
    WHERE admin_content_data_id = ?;

-- name: DeleteAdminContentData :exec
DELETE FROM admin_content_data
WHERE admin_content_data_id = ?;

-- name: ListAdminContentDataByRoute :many
SELECT * FROM admin_content_data
WHERE admin_route_id = ?;

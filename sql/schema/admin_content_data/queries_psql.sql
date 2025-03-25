-- name: CreateAdminContentDataTable :exec
CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id SERIAL PRIMARY KEY,
    admin_route_id INTEGER,
    parent_id INTEGER,
    admin_datatype_id INTEGER,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT DEFAULT NULL,
    CONSTRAINT fk_admin_routes FOREIGN KEY (admin_route_id)
        REFERENCES admin_routes(route_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_parent_id FOREIGN KEY (parent_id)
        REFERENCES admin_content_data(admin_content_data_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_admin_datatypes FOREIGN KEY (admin_datatype_id)
        REFERENCES admin_datatypes(admin_datatype_id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

-- name: GetAdminContentData :one
SELECT * FROM admin_content_data
WHERE admin_content_data_id = $1 LIMIT 1;

-- name: CountAdminContentData :one
SELECT COUNT(*)
FROM admin_content_data;

-- name: ListAdminContentData :many
SELECT * FROM admin_content_data
ORDER BY admin_content_data_id;

-- name: CreateAdminContentData :one
INSERT INTO admin_content_data (
    admin_route_id,
    parent_id,
    admin_datatype_id,
    history,
    date_created,
    date_modified
    ) VALUES (
$1,$2,$3,$4,$5,$6
    ) RETURNING *;


-- name: UpdateAdminContentData :exec
UPDATE admin_content_data
set admin_route_id = $1,
    parent_id = $2,
    admin_datatype_id =$3,
    history = $4,
    date_created = $5,
    date_modified = $6
    WHERE admin_content_data_id = $7
    RETURNING *;

-- name: DeleteAdminContentData :exec
DELETE FROM admin_content_data
WHERE admin_content_data_id = $1;

-- name: ListAdminContentDataByRoute :many
SELECT * FROM admin_content_data
WHERE admin_route_id = $1;

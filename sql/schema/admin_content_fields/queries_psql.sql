-- name: CreateAdminContentFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id SERIAL PRIMARY KEY,
    admin_route_id INTEGER,
    admin_content_data_id INTEGER NOT NULL,
    admin_field_id INTEGER NOT NULL,
    admin_field_value TEXT NOT NULL,
    history TEXT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_admin_route_id FOREIGN KEY (admin_route_id)
        REFERENCES admin_routes(admin_route_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_admin_content_data FOREIGN KEY (admin_content_data_id)
        REFERENCES admin_content_data(admin_content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_fields FOREIGN KEY (admin_field_id)
        REFERENCES admin_fields(admin_field_id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

-- name: GetAdminContentField :one
SELECT * FROM admin_content_fields
WHERE admin_content_field_id = $1 LIMIT 1;

-- name: CountAdminContentField :one
SELECT COUNT(*)
FROM admin_content_fields;

-- name: ListAdminContentFields :many
SELECT * FROM admin_content_fields
ORDER BY admin_content_field_id;

-- name: CreateAdminContentField :one
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
  $1,$2,$3,$4,$5,$6,$7,$8
    ) RETURNING *;


-- name: UpdateAdminContentField :exec
UPDATE admin_content_fields
set  admin_content_field_id=$1,
    admin_route_id=$2,
    admin_content_data_id=$3,
    admin_field_id=$4,
    admin_field_value=$5, 
    history=$6,
    date_created=$7, 
    date_modified=$8
    WHERE admin_content_field_id = $9
    RETURNING *;

-- name: DeleteAdminContentField :exec
DELETE FROM admin_content_fields
WHERE admin_content_field_id = $1;

-- name: ListAdminContentFieldsByRoute :many
SELECT * FROM admin_content_fields
WHERE admin_route_id = $1;

-- name: DropAdminContentDataTable :exec
DROP TABLE admin_content_data;

-- name: CreateAdminContentDataTable :exec
CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id SERIAL
        PRIMARY KEY,
    parent_id INTEGER
        CONSTRAINT fk_parent_id
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    first_child_id INTEGER
        CONSTRAINT fk_first_child_id
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    next_sibling_id INTEGER
        CONSTRAINT fk_first_child_id
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    prev_sibling_id INTEGER
        CONSTRAINT fk_first_child_id
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    admin_route_id INTEGER NOT NULL
        CONSTRAINT fk_admin_routes
            REFERENCES admin_routes
            ON UPDATE CASCADE,
    admin_datatype_id INTEGER NOT NULL
        CONSTRAINT fk_admin_datatypes
            REFERENCES admin_datatypes
            ON UPDATE CASCADE,
    author_id INTEGER DEFAULT '1' NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);


-- name: CountAdminContentData :one
SELECT COUNT(*)
FROM admin_content_data;

-- name: GetAdminContentData :one
SELECT * FROM admin_content_data
WHERE admin_content_data_id = $1 LIMIT 1;

-- name: ListAdminContentData :many
SELECT * FROM admin_content_data
ORDER BY admin_content_data_id;

-- name: ListAdminContentDataByRoute :many
SELECT * FROM admin_content_data
WHERE admin_route_id = $1
ORDER BY admin_content_data_id;

-- name: CreateAdminContentData :one
INSERT INTO admin_content_data (
    parent_id,
    admin_route_id,
    admin_datatype_id,
    author_id,
    date_created,
    date_modified,
    history
) VALUES (    
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
) RETURNING *;

-- name: UpdateAdminContentData :exec
UPDATE admin_content_data
SET parent_id = $1,
    admin_route_id = $2,
    admin_datatype_id =$3,
    author_id = $4,
    date_created = $5,
    date_modified = $6,
    history = $7
WHERE admin_content_data_id = $8;

-- name: DeleteAdminContentData :exec
DELETE FROM admin_content_data
WHERE admin_content_data_id = $1;

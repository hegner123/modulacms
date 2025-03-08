-- name: CreateAdminDatatypeTable :exec
CREATE TABLE IF NOT EXISTS admin_datatypes (
    admin_datatype_id SERIAL PRIMARY KEY,
    admin_route_id INT DEFAULT NULL,
    parent_id INT DEFAULT NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author TEXT NOT NULL,
    author_id INT NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT,
    CONSTRAINT fk_admin_route_id FOREIGN KEY (admin_route_id)
        REFERENCES admin_routes(admin_route_id)
        ON UPDATE CASCADE
        ON DELETE SET DEFAULT,
    CONSTRAINT fk_parent_id FOREIGN KEY (parent_id)
        REFERENCES admin_datatypes(admin_datatype_id)
        ON UPDATE CASCADE
        ON DELETE SET DEFAULT,
    CONSTRAINT fk_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE
        ON DELETE SET DEFAULT,
    CONSTRAINT fk_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE SET DEFAULT
);

-- name: GetAdminDatatype :one
SELECT * FROM admin_datatypes
WHERE admin_datatype_id = $1
LIMIT 1;

-- name: CountAdminDatatype :one
SELECT COUNT(*)
FROM admin_datatypes;

-- name: GetAdminDatatypeId :one
SELECT admin_datatype_id FROM admin_datatypes
WHERE admin_datatype_id = $1
LIMIT 1;

-- name: ListAdminDatatype :many
SELECT * FROM admin_datatypes
ORDER BY admin_datatype_id;

-- name: ListAdminDatatypeTree :many
SELECT 
    child.admin_datatype_id AS child_id,
    child.label AS child_label,
    parent.admin_datatype_id AS parent_id,
    parent.label AS parent_label
FROM admin_datatypes AS child
LEFT JOIN admin_datatypes AS parent 
    ON child.parent_id = parent.admin_datatype_id;

-- name: GetGlobalAdminDatatypeId :one
SELECT * FROM admin_datatypes
WHERE type = 'GLOBALS'
LIMIT 1;

-- name: ListAdminDatatypeChildren :many
SELECT * FROM admin_datatypes
WHERE parent_id = $1;

-- name: CreateAdminDatatype :one
INSERT INTO admin_datatypes (
    admin_route_id,
    parent_id,
    label,
    type,
    author,
    author_id,
    date_created,
    date_modified,
    history
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: UpdateAdminDatatype :exec
UPDATE admin_datatypes
SET admin_route_id = $1,
    parent_id = $2,
    label = $3,
    type = $4,
    author = $5,
    author_id = $6,
    date_created = $7,
    date_modified = $8,
    history = $9
WHERE admin_datatype_id = $10
RETURNING *;

-- name: DeleteAdminDatatype :exec
DELETE FROM admin_datatypes
WHERE admin_datatype_id = $1;

-- name: ListAdminDatatypeByRouteId :many
SELECT admin_datatype_id, admin_route_id, parent_id, label, type, history
FROM admin_datatypes
WHERE admin_route_id = $1;

-- name: GetRootAdminDtByAdminRtId :one
SELECT admin_datatype_id, admin_route_id, parent_id, label, type, history
FROM admin_datatypes
WHERE admin_route_id = $1
ORDER BY admin_datatype_id;

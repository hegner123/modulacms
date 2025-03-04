-- name: CreateAdminDatatypeTable :exec
CREATE TABLE IF NOT EXISTS admin_datatypes (
    admin_dt_id INT NOT NULL AUTO_INCREMENT,
    admin_route_id INT DEFAULT NULL,
    parent_id INT DEFAULT NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT,
    PRIMARY KEY (admin_dt_id),
    CONSTRAINT fk_admin_datatypes_admin_route_id FOREIGN KEY (admin_route_id)
        REFERENCES admin_routes(admin_route_id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION,
    CONSTRAINT fk_admin_datatypes_parent_id FOREIGN KEY (parent_id)
        REFERENCES admin_datatypes(admin_dt_id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION,
    CONSTRAINT fk_admin_datatypes_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE
        ON DELETE NO ACTION,
    CONSTRAINT fk_admin_datatypes_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- name: GetAdminDatatype :one
SELECT * FROM admin_datatypes
WHERE admin_dt_id = ? 
LIMIT 1;

-- name: CountAdminDatatype :one
SELECT COUNT(*)
FROM admin_datatypes;

-- name: GetAdminDatatypeId :one
SELECT admin_dt_id FROM admin_datatypes
WHERE admin_dt_id = ? 
LIMIT 1;

-- name: ListAdminDatatype :many
SELECT * FROM admin_datatypes
ORDER BY admin_dt_id;

-- name: ListAdminDatatypeTree :many
SELECT 
    child.admin_dt_id AS child_id,
    child.label AS child_label,
    parent.admin_dt_id AS parent_id,
    parent.label AS parent_label
FROM admin_datatypes AS child
LEFT JOIN admin_datatypes AS parent 
    ON child.parent_id = parent.admin_dt_id;

-- name: GetGlobalAdminDatatypeId :one
SELECT * FROM admin_datatypes
WHERE type = 'GLOBALS' 
LIMIT 1;

-- name: ListAdminDatatypeChildren :many
SELECT * FROM admin_datatypes
WHERE parent_id = ?;

-- name: CreateAdminDatatype :exec
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
    ?,?,?,?,?,?,?,?,?
);
-- To retrieve the inserted row, consider using LAST_INSERT_ID() in a subsequent SELECT.
-- name: GetLastAdminDatatype :one
SELECT * FROM admin_datatypes WHERE admin_dt_id = LAST_INSERT_ID();

-- name: UpdateAdminDatatype :exec
UPDATE admin_datatypes
SET admin_route_id = ?,
    parent_id = ?,
    label = ?,
    type = ?,
    author = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?,
    history = ?
WHERE admin_dt_id = ?;
-- To retrieve the updated row, execute a subsequent SELECT.

-- name: DeleteAdminDatatype :exec
DELETE FROM admin_datatypes
WHERE admin_dt_id = ?;

-- name: ListAdminDatatypeByRouteId :many
SELECT admin_dt_id, admin_route_id, parent_id, label, type, history
FROM admin_datatypes
WHERE admin_route_id = ?;

-- name: GetRootAdminDtByAdminRtId :one
SELECT admin_dt_id, admin_route_id, parent_id, label, type, history
FROM admin_datatypes
WHERE admin_route_id = ?
ORDER BY admin_dt_id;

-- name: CheckAuthorIdExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE user_id=?);

-- name: CheckAuthorExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE username=?);

-- name: CheckAdminRouteExists :one
SELECT EXISTS(SELECT 1 FROM admin_routes WHERE admin_route_id=?);

-- name: CheckAdminParentExists :one
SELECT EXISTS(SELECT 1 FROM admin_datatypes WHERE admin_dt_id =?);

-- name: CheckRouteExists :one
SELECT EXISTS(SELECT 1 FROM routes WHERE route_id=?);

-- name: CheckParentExists :one
SELECT EXISTS(SELECT 1 FROM datatypes WHERE datatype_id =?);


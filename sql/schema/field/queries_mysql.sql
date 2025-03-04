-- name: CreateFieldTable :exec
CREATE TABLE IF NOT EXISTS fields (
    field_id INT AUTO_INCREMENT PRIMARY KEY,
    route_id INT DEFAULT NULL,
    parent_id INT DEFAULT NULL,
    label VARCHAR(255) NOT NULL DEFAULT 'unlabeled',
    data TEXT NOT NULL,
    type TEXT NOT NULL,
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    history TEXT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_fields_routes FOREIGN KEY (route_id)
        REFERENCES routes(route_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_fields_datatypes FOREIGN KEY (parent_id)
        REFERENCES datatypes(datatype_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_fields_users_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_fields_users_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- name: GetField :one
SELECT * FROM fields 
WHERE field_id = ? LIMIT 1;

-- name: CountField :one
SELECT COUNT(*)
FROM fields ;

-- name: ListField :many
SELECT * FROM fields 
ORDER BY field_id;

-- name: CreateField :exec
INSERT INTO fields  (
    route_id,
    parent_id,
    label,
    data,
    type,
    author,
    author_id,
    history,
    date_created,
    date_modified
    ) VALUES (
?,?,?,?,?,?,?,?,?,?
    );
-- name: GetLastField :one
SELECT * FROM fields WHERE field_id = LAST_INSERT_ID();


-- name: UpdateField :exec
UPDATE fields 
set route_id = ?,
    parent_id = ?,
    label = ?,
    data = ?,
    type = ?,
    author = ?,
    author_id = ?,
    history =?,
    date_created = ?,
    date_modified = ?
    WHERE field_id = ?;

-- name: DeleteField :exec
DELETE FROM fields 
WHERE field_id = ?;

-- name: ListFieldByRouteId :many
SELECT field_id, route_id, parent_id, label, data, type
FROM fields 
WHERE route_id = ?;

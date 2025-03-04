-- name: CreateDatatypeTable :exec
CREATE TABLE IF NOT EXISTS datatypes (
    datatype_id SERIAL PRIMARY KEY,
    route_id INTEGER DEFAULT NULL,
    parent_id INTEGER DEFAULT NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author TEXT NOT NULL DEFAULT 'system',
    author_id INTEGER NOT NULL DEFAULT 1,
    history TEXT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_routes FOREIGN KEY (route_id)
        REFERENCES routes(route_id)
        ON UPDATE CASCADE ON DELETE SET DEFAULT,
    CONSTRAINT fk_datatypes_parent FOREIGN KEY (parent_id)
        REFERENCES datatypes(datatype_id)
        ON UPDATE CASCADE ON DELETE SET DEFAULT,
    CONSTRAINT fk_users_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE ON DELETE SET DEFAULT,
    CONSTRAINT fk_users_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE SET DEFAULT
);

-- name: GetDatatype :one
SELECT * FROM datatypes
WHERE datatype_id = $1 LIMIT 1;

-- name: CountDatatype :one
SELECT COUNT(*)
FROM datatypes;

-- name: ListDatatype :many
SELECT * FROM datatypes
ORDER BY datatype_id;

-- name: CreateDatatype :one
INSERT INTO datatypes (
    route_id,
    parent_id,
    label,
    type,
    author,
    author_id,
    history,
    date_created,
    date_modified
    ) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9
    ) RETURNING *;

-- name: UpdateDatatype :exec
UPDATE datatypes
set route_id = $1,
    parent_id = $2,
    label = $3,
    type = $4,
    author = $5,
    author_id = $6,
    history = $7,
    date_created = $8,
    date_modified = $9
    WHERE datatype_id = $10
    RETURNING *;

-- name: DeleteDatatype :exec
DELETE FROM datatypes
WHERE datatype_id = $1;

-- name: ListDatatypeByRouteId :many
SELECT datatype_id, route_id, parent_id, label, type
FROM datatypes
WHERE route_id = $1;

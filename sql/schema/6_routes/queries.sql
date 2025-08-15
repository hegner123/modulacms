-- name: DropRouteTable :exec
DROP TABLE routes;

-- name: CreateRouteTable :exec
CREATE TABLE IF NOT EXISTS routes (
    route_id INTEGER
        PRIMARY KEY,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT  DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);

-- name: CountRoute :one
SELECT COUNT(*)
FROM routes;

-- name: GetRoute :one
SELECT * 
FROM routes
WHERE route_id = ? 
LIMIT 1;


-- name: GetRouteIDBySlug :one
SELECT route_id 
FROM routes
WHERE slug = ? 
LIMIT 1;

-- name: ListRoute :many
SELECT * 
FROM routes
ORDER BY slug;

-- name: CreateRoute :one
INSERT INTO routes (
    slug,
    title,
    status,
    author_id,
    history,
    date_created,
    date_modified
) VALUES ( 
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
) 
RETURNING *;

-- name: UpdateRoute :exec
UPDATE routes
SET slug = ?,
    title = ?,
    status = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?,
    history = ? 
WHERE slug = ?
RETURNING *;

-- name: DeleteRoute :exec
DELETE FROM routes
WHERE route_id = ?;

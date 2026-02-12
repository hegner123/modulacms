-- name: DropRouteTable :exec
DROP TABLE routes;

-- name: CreateRouteTable :exec
CREATE TABLE IF NOT EXISTS routes (
    route_id TEXT PRIMARY KEY NOT NULL CHECK (length(route_id) = 26),
    author_id TEXT NOT NULL
        REFERENCES users
            ON DELETE SET NULL,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT  DEFAULT CURRENT_TIMESTAMP
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
    route_id,
    slug,
    title,
    status,
    author_id,

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
    date_modified = ?
WHERE slug = ?
RETURNING *;

-- name: DeleteRoute :exec
DELETE FROM routes
WHERE route_id = ?;

-- name: ListRoutePaginated :many
SELECT * FROM routes
ORDER BY slug
LIMIT ? OFFSET ?;

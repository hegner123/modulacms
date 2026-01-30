-- name: DropRouteTable :exec
DROP TABLE routes;

-- name: CreateRouteTable :exec
CREATE TABLE IF NOT EXISTS routes (
    route_id TEXT PRIMARY KEY NOT NULL,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id TEXT NOT NULL
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- name: CountRoute :one
SELECT COUNT(*)
FROM routes;

-- name: GetRoute :one
SELECT * 
FROM routes
WHERE route_id = $1
LIMIT 1;

-- name: GetRouteIDBySlug :one
SELECT route_id 
FROM routes
WHERE slug = $1
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
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING *;

-- name: UpdateRoute :exec
UPDATE routes
SET slug = $1,
    title = $2,
    status = $3,
    author_id = $4,
    date_created = $5,
    date_modified = $6
WHERE slug = $7
RETURNING *;

-- name: DeleteRoute :exec
DELETE FROM routes
WHERE route_id = $1;


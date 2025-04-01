-- name: DropRouteTable :exec
DROP TABLE routes;

-- name: CreateRouteTable :exec
CREATE TABLE IF NOT EXISTS routes (
    route_id SERIAL
        PRIMARY KEY,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
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
    slug,
    title,
    status,
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
)
RETURNING *;

-- name: UpdateRoute :exec
UPDATE routes
SET slug = $1,
    title = $2,
    status = $3,
    author_id = $4,
    history = $5,
    date_created = $6,
    date_modified = $7
WHERE slug = $8
RETURNING *;

-- name: DeleteRoute :exec
DELETE FROM routes
WHERE route_id = $1;


-- name: CreateRouteTable :exec
CREATE TABLE IF NOT EXISTS routes (
    route_id INTEGER
        PRIMARY KEY,
    author TEXT DEFAULT "system" NOT NULL
        REFERENCES users (username)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    history TEXT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
-- name: GetRoute :one
SELECT * FROM routes
WHERE slug = $1
LIMIT 1;

-- name: CountRoute :one
SELECT COUNT(*)
FROM routes;

-- name: GetRouteID :one
SELECT route_id FROM routes
WHERE slug = $1
LIMIT 1;

-- name: ListRoute :many
SELECT * FROM routes
ORDER BY slug;

-- name: CreateRoute :one
INSERT INTO routes (
    author,
    author_id,
    slug,
    title,
    status,
    history,
    date_created,
    date_modified
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdateRoute :exec
UPDATE routes
SET slug = $1,
    title = $2,
    status = $3,
    history = $4,
    author = $5,
    author_id = $6,
    date_created = $7,
    date_modified = $8
WHERE slug = $9
RETURNING *;

-- name: DeleteRoute :exec
DELETE FROM routes
WHERE slug = $1;


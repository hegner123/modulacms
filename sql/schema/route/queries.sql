-- name: CreateRouteTable :exec
CREATE TABLE IF NOT EXISTS routes (
    route_id INTEGER PRIMARY KEY,
    author TEXT DEFAULT 'system' NOT NULL
        REFERENCES users(username)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users(user_id)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    history TEXT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- name: GetRoute :one
SELECT * FROM routes
WHERE slug = ? LIMIT 1;

-- name: CountRoute :one
SELECT COUNT(*)
FROM routes;

-- name: GetRouteID :one
SELECT route_id FROM routes
WHERE slug = ? LIMIT 1;

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
?,?,?,?,?,?,?,?
) RETURNING *;

-- name: UpdateRoute :exec
UPDATE routes
set slug = ?,
    title = ?,
    status = ?,
    history= ?,
    author = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
    WHERE slug = ?
    RETURNING *;

-- name: DeleteRoute :exec
DELETE FROM routes
WHERE slug = ?;

-- name: CreateRouteTable :exec
CREATE TABLE IF NOT EXISTS routes (
    route_id INT NOT NULL AUTO_INCREMENT,
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    slug VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    status INT NOT NULL,
    history TEXT,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (route_id),
    UNIQUE KEY unique_slug (slug),
    CONSTRAINT fk_routes_routes_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE
        ON DELETE NO ACTION,
    CONSTRAINT fk_routes_routes_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- name: GetRoute :one
SELECT * FROM routes
WHERE slug = ? 
LIMIT 1;

-- name: CountRoute :one
SELECT COUNT(*) 
FROM routes;

-- name: GetRouteId :one
SELECT route_id 
FROM routes
WHERE slug = ? 
LIMIT 1;

-- name: GetLastRoute :one
SELECT * FROM routes WHERE route_id = LAST_INSERT_ID();

-- name: ListRoute :many
SELECT * FROM routes
ORDER BY slug;

-- name: CreateRoute :exec
INSERT INTO routes (
    author,
    author_id,
    slug,
    title,
    status,
    history,
    date_created,
    date_modified
) VALUES (?,?,?,?,?,?,?,?);

-- name: UpdateRoute :exec
UPDATE routes
SET slug = ?,
    title = ?,
    status = ?,
    history = ?,
    author = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
WHERE slug = ?;

-- name: DeleteRoute :exec
DELETE FROM routes
WHERE slug = ?;


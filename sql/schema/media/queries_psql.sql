-- name: CreateMediaTable :exec
CREATE TABLE IF NOT EXISTS media (
    media_id SERIAL PRIMARY KEY,
    name TEXT,
    display_name TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    mimetype TEXT,
    dimensions TEXT,
    url TEXT UNIQUE,
    optimized_mobile TEXT,
    optimized_tablet TEXT,
    optimized_desktop TEXT,
    optimized_ultra_wide TEXT,
    author TEXT NOT NULL DEFAULT 'system',
    author_id INTEGER NOT NULL DEFAULT 1,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_users_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE ON DELETE SET DEFAULT,
    CONSTRAINT fk_users_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE SET DEFAULT
);

-- name: GetMedia :one
SELECT * FROM media
WHERE media_id = $1 LIMIT 1;

-- name: CountMedia :one
SELECT COUNT(*)
FROM media;

-- name: ListMedia :many
SELECT * FROM media
ORDER BY name;

-- name: CreateMedia :one
INSERT INTO media (
    name,
    display_name,
    alt,
    caption,
    description,
    class,
    url,
    mimetype,
    dimensions,
    optimized_mobile,
    optimized_tablet,
    optimized_desktop,
    optimized_ultra_wide,
    author,
    author_id,
    date_created,
    date_modified
) VALUES (
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17
)
RETURNING *;

-- name: UpdateMedia :exec
UPDATE media
  set   name = $1,
        display_name = $2,
        alt = $3,
        caption = $4,
        description = $5,
        class = $6,
        author = $7,
        author_id = $8,
        date_created = $9,
        date_modified = $10,
        url = $11,
        mimetype = $12,
        dimensions = $13,
        optimized_mobile = $14,
        optimized_tablet = $15,
        optimized_desktop = $16,
        optimized_ultra_wide = $17
        WHERE media_id = $18;

-- name: DeleteMedia :exec
DELETE FROM media
WHERE media_id = $1;

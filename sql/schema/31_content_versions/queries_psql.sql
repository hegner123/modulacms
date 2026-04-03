-- name: CreateContentVersionTable :exec
CREATE TABLE IF NOT EXISTS content_versions (
    content_version_id TEXT PRIMARY KEY NOT NULL CHECK (length(content_version_id) = 26),
    content_data_id TEXT NOT NULL
        REFERENCES content_data(content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    locale TEXT NOT NULL DEFAULT '',
    snapshot TEXT NOT NULL,
    trigger TEXT NOT NULL DEFAULT 'manual',
    label TEXT NOT NULL DEFAULT '',
    published BOOLEAN NOT NULL DEFAULT FALSE,
    published_by TEXT
        REFERENCES users(user_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- name: DropContentVersionTable :exec
DROP TABLE IF EXISTS content_versions;

-- name: CountContentVersions :one
SELECT COUNT(*) FROM content_versions;

-- name: CreateContentVersion :one
INSERT INTO content_versions (
    content_version_id,
    content_data_id,
    version_number,
    locale,
    snapshot,
    trigger,
    label,
    published,
    published_by,
    date_created
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetContentVersion :one
SELECT * FROM content_versions
WHERE content_version_id = $1 LIMIT 1;

-- name: GetPublishedSnapshot :one
SELECT * FROM content_versions
WHERE content_data_id = $1 AND locale = $2 AND published = TRUE
ORDER BY version_number DESC LIMIT 1;

-- name: ListContentVersionsByContent :many
SELECT * FROM content_versions
WHERE content_data_id = $1
ORDER BY version_number DESC;

-- name: ListContentVersionsByContentLocale :many
SELECT * FROM content_versions
WHERE content_data_id = $1 AND locale = $2
ORDER BY version_number DESC;

-- name: ClearPublishedFlag :exec
UPDATE content_versions
SET published = FALSE
WHERE content_data_id = $1 AND locale = $2 AND published = TRUE;

-- name: DeleteContentVersion :exec
DELETE FROM content_versions
WHERE content_version_id = $1;

-- name: CountContentVersionsByContent :one
SELECT COUNT(*) FROM content_versions
WHERE content_data_id = $1;

-- name: GetMaxVersionNumber :one
SELECT COALESCE(MAX(version_number), 0) FROM content_versions
WHERE content_data_id = $1 AND locale = $2;

-- name: PruneOldVersions :exec
DELETE FROM content_versions
WHERE content_version_id IN (
    SELECT cv.content_version_id FROM content_versions cv
    WHERE cv.content_data_id = $1 AND cv.locale = $2
        AND cv.published = FALSE AND cv.label = ''
    ORDER BY cv.version_number ASC
    LIMIT $3
);

-- name: GetMaxVersionNumberForUpdate :one
SELECT COALESCE(MAX(version_number), 0) FROM content_versions
WHERE content_data_id = $1 AND locale = $2
FOR UPDATE;

-- name: ListDuplicatePublished :many
SELECT content_data_id, locale, COUNT(*) as pub_count
FROM content_versions WHERE published = TRUE
GROUP BY content_data_id, locale HAVING COUNT(*) > 1;

-- name: ClearPublishedFlagExcept :exec
UPDATE content_versions SET published = FALSE
WHERE content_data_id = $1 AND locale = $2 AND published = TRUE AND content_version_id != $3;

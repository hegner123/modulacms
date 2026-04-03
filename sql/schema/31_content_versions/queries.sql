-- name: CreateContentVersionTable :exec
CREATE TABLE IF NOT EXISTS content_versions (
    content_version_id TEXT PRIMARY KEY NOT NULL CHECK (length(content_version_id) = 26),
    content_data_id TEXT NOT NULL
        REFERENCES content_data(content_data_id)
            ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    locale TEXT NOT NULL DEFAULT '',
    snapshot TEXT NOT NULL,
    trigger TEXT NOT NULL DEFAULT 'manual',
    label TEXT NOT NULL DEFAULT '',
    published INTEGER NOT NULL DEFAULT 0,
    published_by TEXT
        REFERENCES users(user_id)
            ON DELETE SET NULL,
    date_created TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
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
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetContentVersion :one
SELECT * FROM content_versions
WHERE content_version_id = ? LIMIT 1;

-- name: GetPublishedSnapshot :one
SELECT * FROM content_versions
WHERE content_data_id = ? AND locale = ? AND published = 1
ORDER BY version_number DESC LIMIT 1;

-- name: ListContentVersionsByContent :many
SELECT * FROM content_versions
WHERE content_data_id = ?
ORDER BY version_number DESC;

-- name: ListContentVersionsByContentLocale :many
SELECT * FROM content_versions
WHERE content_data_id = ? AND locale = ?
ORDER BY version_number DESC;

-- name: ClearPublishedFlag :exec
UPDATE content_versions
SET published = 0
WHERE content_data_id = ? AND locale = ? AND published = 1;

-- name: DeleteContentVersion :exec
DELETE FROM content_versions
WHERE content_version_id = ?;

-- name: CountContentVersionsByContent :one
SELECT COUNT(*) FROM content_versions
WHERE content_data_id = ?;

-- name: GetMaxVersionNumber :one
SELECT COALESCE(MAX(version_number), 0) FROM content_versions
WHERE content_data_id = ? AND locale = ?;

-- name: PruneOldVersions :exec
DELETE FROM content_versions
WHERE content_version_id IN (
    SELECT cv.content_version_id FROM content_versions cv
    WHERE cv.content_data_id = ? AND cv.locale = ?
        AND cv.published = 0 AND cv.label = ''
    ORDER BY cv.version_number ASC
    LIMIT ?
);

-- name: GetMaxVersionNumberForUpdate :one
SELECT COALESCE(MAX(version_number), 0) FROM content_versions
WHERE content_data_id = ? AND locale = ?;

-- name: ListDuplicatePublished :many
SELECT content_data_id, locale, COUNT(*) as pub_count
FROM content_versions WHERE published = 1
GROUP BY content_data_id, locale HAVING COUNT(*) > 1;

-- name: ClearPublishedFlagExcept :exec
UPDATE content_versions SET published = 0
WHERE content_data_id = ? AND locale = ? AND published = 1 AND content_version_id != ?;

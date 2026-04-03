-- name: CreateContentVersionTable :exec
CREATE TABLE IF NOT EXISTS content_versions (
    content_version_id VARCHAR(26) NOT NULL,
    content_data_id VARCHAR(26) NOT NULL,
    version_number INT NOT NULL,
    locale VARCHAR(10) NOT NULL DEFAULT '',
    snapshot MEDIUMTEXT NOT NULL,
    `trigger` VARCHAR(50) NOT NULL DEFAULT 'manual',
    label VARCHAR(255) NOT NULL DEFAULT '',
    published TINYINT NOT NULL DEFAULT 0,
    published_by VARCHAR(26),
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (content_version_id),
    CONSTRAINT fk_cv_content FOREIGN KEY (content_data_id)
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_cv_published_by FOREIGN KEY (published_by)
        REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

-- name: DropContentVersionTable :exec
DROP TABLE IF EXISTS content_versions;

-- name: CountContentVersions :one
SELECT COUNT(*) FROM content_versions;

-- name: CreateContentVersion :exec
INSERT INTO content_versions (
    content_version_id,
    content_data_id,
    version_number,
    locale,
    snapshot,
    `trigger`,
    label,
    published,
    published_by,
    date_created
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

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
WHERE content_data_id = ? AND locale = ?
    AND published = 0 AND label = ''
ORDER BY version_number ASC
LIMIT ?;

-- name: GetMaxVersionNumberForUpdate :one
SELECT COALESCE(MAX(version_number), 0) FROM content_versions
WHERE content_data_id = ? AND locale = ?
FOR UPDATE;

-- name: ListDuplicatePublished :many
SELECT content_data_id, locale, COUNT(*) as pub_count
FROM content_versions WHERE published = 1
GROUP BY content_data_id, locale HAVING COUNT(*) > 1;

-- name: ClearPublishedFlagExcept :exec
UPDATE content_versions SET published = 0
WHERE content_data_id = ? AND locale = ? AND published = 1 AND content_version_id != ?;

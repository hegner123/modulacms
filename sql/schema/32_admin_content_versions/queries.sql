-- name: CreateAdminContentVersionTable :exec
CREATE TABLE IF NOT EXISTS admin_content_versions (
    admin_content_version_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_content_version_id) = 26),
    admin_content_data_id TEXT NOT NULL
        REFERENCES admin_content_data(admin_content_data_id)
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

-- name: DropAdminContentVersionTable :exec
DROP TABLE IF EXISTS admin_content_versions;

-- name: CountAdminContentVersions :one
SELECT COUNT(*) FROM admin_content_versions;

-- name: CreateAdminContentVersion :one
INSERT INTO admin_content_versions (
    admin_content_version_id,
    admin_content_data_id,
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

-- name: GetAdminContentVersion :one
SELECT * FROM admin_content_versions
WHERE admin_content_version_id = ? LIMIT 1;

-- name: GetAdminPublishedSnapshot :one
SELECT * FROM admin_content_versions
WHERE admin_content_data_id = ? AND locale = ? AND published = 1
ORDER BY version_number DESC LIMIT 1;

-- name: ListAdminContentVersionsByContent :many
SELECT * FROM admin_content_versions
WHERE admin_content_data_id = ?
ORDER BY version_number DESC;

-- name: ListAdminContentVersionsByContentLocale :many
SELECT * FROM admin_content_versions
WHERE admin_content_data_id = ? AND locale = ?
ORDER BY version_number DESC;

-- name: ClearAdminPublishedFlag :exec
UPDATE admin_content_versions
SET published = 0
WHERE admin_content_data_id = ? AND locale = ? AND published = 1;

-- name: DeleteAdminContentVersion :exec
DELETE FROM admin_content_versions
WHERE admin_content_version_id = ?;

-- name: CountAdminContentVersionsByContent :one
SELECT COUNT(*) FROM admin_content_versions
WHERE admin_content_data_id = ?;

-- name: GetAdminMaxVersionNumber :one
SELECT COALESCE(MAX(version_number), 0) FROM admin_content_versions
WHERE admin_content_data_id = ? AND locale = ?;

-- name: PruneAdminOldVersions :exec
DELETE FROM admin_content_versions
WHERE admin_content_version_id IN (
    SELECT acv.admin_content_version_id FROM admin_content_versions acv
    WHERE acv.admin_content_data_id = ? AND acv.locale = ?
        AND acv.published = 0 AND acv.label = ''
    ORDER BY acv.version_number ASC
    LIMIT ?
);

-- name: GetAdminMaxVersionNumberForUpdate :one
SELECT COALESCE(MAX(version_number), 0) FROM admin_content_versions
WHERE admin_content_data_id = ? AND locale = ?;

-- name: ListAdminDuplicatePublished :many
SELECT admin_content_data_id, locale, COUNT(*) as pub_count
FROM admin_content_versions WHERE published = 1
GROUP BY admin_content_data_id, locale HAVING COUNT(*) > 1;

-- name: ClearAdminPublishedFlagExcept :exec
UPDATE admin_content_versions SET published = 0
WHERE admin_content_data_id = ? AND locale = ? AND published = 1 AND admin_content_version_id != ?;

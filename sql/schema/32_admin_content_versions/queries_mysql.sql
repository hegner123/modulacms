-- name: CreateAdminContentVersionTable :exec
CREATE TABLE IF NOT EXISTS admin_content_versions (
    admin_content_version_id VARCHAR(26) NOT NULL,
    admin_content_data_id VARCHAR(26) NOT NULL,
    version_number INT NOT NULL,
    locale VARCHAR(10) NOT NULL DEFAULT '',
    snapshot MEDIUMTEXT NOT NULL,
    `trigger` VARCHAR(50) NOT NULL DEFAULT 'manual',
    label VARCHAR(255) NOT NULL DEFAULT '',
    published TINYINT NOT NULL DEFAULT 0,
    published_by VARCHAR(26),
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (admin_content_version_id),
    CONSTRAINT fk_acv_content FOREIGN KEY (admin_content_data_id)
        REFERENCES admin_content_data(admin_content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_acv_published_by FOREIGN KEY (published_by)
        REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

-- name: DropAdminContentVersionTable :exec
DROP TABLE IF EXISTS admin_content_versions;

-- name: CountAdminContentVersions :one
SELECT COUNT(*) FROM admin_content_versions;

-- name: CreateAdminContentVersion :exec
INSERT INTO admin_content_versions (
    admin_content_version_id,
    admin_content_data_id,
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
WHERE admin_content_data_id = ? AND locale = ?
    AND published = 0 AND label = ''
ORDER BY version_number ASC
LIMIT ?;

-- name: GetAdminMaxVersionNumberForUpdate :one
SELECT COALESCE(MAX(version_number), 0) FROM admin_content_versions
WHERE admin_content_data_id = ? AND locale = ?
FOR UPDATE;

-- name: ListAdminDuplicatePublished :many
SELECT admin_content_data_id, locale, COUNT(*) as pub_count
FROM admin_content_versions WHERE published = 1
GROUP BY admin_content_data_id, locale HAVING COUNT(*) > 1;

-- name: ClearAdminPublishedFlagExcept :exec
UPDATE admin_content_versions SET published = 0
WHERE admin_content_data_id = ? AND locale = ? AND published = 1 AND admin_content_version_id != ?;

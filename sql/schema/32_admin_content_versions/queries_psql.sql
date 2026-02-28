-- name: CreateAdminContentVersionTable :exec
CREATE TABLE IF NOT EXISTS admin_content_versions (
    admin_content_version_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_content_version_id) = 26),
    admin_content_data_id TEXT NOT NULL
        REFERENCES admin_content_data(admin_content_data_id)
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
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetAdminContentVersion :one
SELECT * FROM admin_content_versions
WHERE admin_content_version_id = $1 LIMIT 1;

-- name: GetAdminPublishedSnapshot :one
SELECT * FROM admin_content_versions
WHERE admin_content_data_id = $1 AND locale = $2 AND published = TRUE LIMIT 1;

-- name: ListAdminContentVersionsByContent :many
SELECT * FROM admin_content_versions
WHERE admin_content_data_id = $1
ORDER BY version_number DESC;

-- name: ListAdminContentVersionsByContentLocale :many
SELECT * FROM admin_content_versions
WHERE admin_content_data_id = $1 AND locale = $2
ORDER BY version_number DESC;

-- name: ClearAdminPublishedFlag :exec
UPDATE admin_content_versions
SET published = FALSE
WHERE admin_content_data_id = $1 AND locale = $2 AND published = TRUE;

-- name: DeleteAdminContentVersion :exec
DELETE FROM admin_content_versions
WHERE admin_content_version_id = $1;

-- name: CountAdminContentVersionsByContent :one
SELECT COUNT(*) FROM admin_content_versions
WHERE admin_content_data_id = $1;

-- name: GetAdminMaxVersionNumber :one
SELECT COALESCE(MAX(version_number), 0) FROM admin_content_versions
WHERE admin_content_data_id = $1 AND locale = $2;

-- name: PruneAdminOldVersions :exec
DELETE FROM admin_content_versions
WHERE admin_content_version_id IN (
    SELECT acv.admin_content_version_id FROM admin_content_versions acv
    WHERE acv.admin_content_data_id = $1 AND acv.locale = $2
        AND acv.published = FALSE AND acv.label = ''
    ORDER BY acv.version_number ASC
    LIMIT $3
);

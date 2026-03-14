-- name: DropFieldPluginConfigTable :exec
DROP TABLE IF EXISTS field_plugin_config;

-- name: DropAdminFieldPluginConfigTable :exec
DROP TABLE IF EXISTS admin_field_plugin_config;

-- name: CreateFieldPluginConfigTable :exec
CREATE TABLE IF NOT EXISTS field_plugin_config (
    field_id         TEXT PRIMARY KEY NOT NULL
        REFERENCES fields(field_id) ON DELETE CASCADE,
    plugin_name      TEXT NOT NULL,
    plugin_interface TEXT NOT NULL,
    plugin_version   TEXT NOT NULL DEFAULT '',
    date_created     TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    date_modified    TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- name: CreateAdminFieldPluginConfigTable :exec
CREATE TABLE IF NOT EXISTS admin_field_plugin_config (
    field_id         TEXT PRIMARY KEY NOT NULL
        REFERENCES admin_fields(field_id) ON DELETE CASCADE,
    plugin_name      TEXT NOT NULL,
    plugin_interface TEXT NOT NULL,
    plugin_version   TEXT NOT NULL DEFAULT '',
    date_created     TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    date_modified    TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- name: GetFieldPluginConfig :one
SELECT * FROM field_plugin_config
WHERE field_id = ? LIMIT 1;

-- name: GetAdminFieldPluginConfig :one
SELECT * FROM admin_field_plugin_config
WHERE field_id = ? LIMIT 1;

-- name: CreateFieldPluginConfig :exec
INSERT INTO field_plugin_config (
    field_id, plugin_name, plugin_interface, plugin_version,
    date_created, date_modified
) VALUES (?, ?, ?, ?, ?, ?);

-- name: CreateAdminFieldPluginConfig :exec
INSERT INTO admin_field_plugin_config (
    field_id, plugin_name, plugin_interface, plugin_version,
    date_created, date_modified
) VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateFieldPluginConfig :exec
UPDATE field_plugin_config
SET plugin_name = ?, plugin_interface = ?, plugin_version = ?, date_modified = ?
WHERE field_id = ?;

-- name: UpdateAdminFieldPluginConfig :exec
UPDATE admin_field_plugin_config
SET plugin_name = ?, plugin_interface = ?, plugin_version = ?, date_modified = ?
WHERE field_id = ?;

-- name: DeleteFieldPluginConfig :exec
DELETE FROM field_plugin_config WHERE field_id = ?;

-- name: DeleteAdminFieldPluginConfig :exec
DELETE FROM admin_field_plugin_config WHERE field_id = ?;

-- name: ListFieldPluginConfigByPlugin :many
SELECT * FROM field_plugin_config
WHERE plugin_name = ?
ORDER BY date_created;

-- name: ListAdminFieldPluginConfigByPlugin :many
SELECT * FROM admin_field_plugin_config
WHERE plugin_name = ?
ORDER BY date_created;

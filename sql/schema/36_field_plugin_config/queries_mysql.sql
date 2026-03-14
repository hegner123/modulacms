-- name: DropFieldPluginConfigTable :exec
DROP TABLE IF EXISTS field_plugin_config;

-- name: DropAdminFieldPluginConfigTable :exec
DROP TABLE IF EXISTS admin_field_plugin_config;

-- name: CreateFieldPluginConfigTable :exec
CREATE TABLE IF NOT EXISTS field_plugin_config (
    field_id         VARCHAR(26) PRIMARY KEY NOT NULL,
    plugin_name      VARCHAR(255) NOT NULL,
    plugin_interface VARCHAR(255) NOT NULL,
    plugin_version   VARCHAR(255) NOT NULL DEFAULT '',
    date_created     TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified    TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_fpc_field FOREIGN KEY (field_id) REFERENCES fields(field_id) ON DELETE CASCADE
);

-- name: CreateAdminFieldPluginConfigTable :exec
CREATE TABLE IF NOT EXISTS admin_field_plugin_config (
    field_id         VARCHAR(26) PRIMARY KEY NOT NULL,
    plugin_name      VARCHAR(255) NOT NULL,
    plugin_interface VARCHAR(255) NOT NULL,
    plugin_version   VARCHAR(255) NOT NULL DEFAULT '',
    date_created     TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified    TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_afpc_field FOREIGN KEY (field_id) REFERENCES admin_fields(field_id) ON DELETE CASCADE
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

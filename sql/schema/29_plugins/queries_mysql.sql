-- name: CreatePluginsTable :exec
CREATE TABLE IF NOT EXISTS plugins (
    plugin_id VARCHAR(26) NOT NULL,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(64) NOT NULL,
    description TEXT NOT NULL,
    author VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'installed',
    capabilities JSON NOT NULL,
    approved_access JSON NOT NULL,
    manifest_hash VARCHAR(64) NOT NULL DEFAULT '',
    date_installed DATETIME NOT NULL,
    date_modified DATETIME NOT NULL,
    PRIMARY KEY (plugin_id),
    CONSTRAINT uq_plugins_name UNIQUE (name)
);

-- name: CreatePluginsIndexStatus :exec
CREATE INDEX idx_plugins_status ON plugins(status);

-- name: CreatePluginsIndexName :exec
CREATE INDEX idx_plugins_name ON plugins(name);

-- name: DropPluginsTable :exec
DROP TABLE IF EXISTS plugins;

-- name: GetPlugin :one
SELECT * FROM plugins WHERE plugin_id = ? LIMIT 1;

-- name: GetPluginByName :one
SELECT * FROM plugins WHERE name = ? LIMIT 1;

-- name: ListPlugins :many
SELECT * FROM plugins ORDER BY name;

-- name: ListPluginsByStatus :many
SELECT * FROM plugins WHERE status = ? ORDER BY name;

-- name: CreatePlugin :exec
INSERT INTO plugins (plugin_id, name, version, description, author, status, capabilities, approved_access, manifest_hash, date_installed, date_modified)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdatePlugin :exec
UPDATE plugins SET version = ?, description = ?, author = ?, status = ?, capabilities = ?, approved_access = ?, manifest_hash = ?, date_modified = ? WHERE plugin_id = ?;

-- name: UpdatePluginStatus :exec
UPDATE plugins SET status = ?, date_modified = ? WHERE plugin_id = ?;

-- name: DeletePlugin :exec
DELETE FROM plugins WHERE plugin_id = ?;

-- name: CountPlugins :one
SELECT COUNT(*) FROM plugins;

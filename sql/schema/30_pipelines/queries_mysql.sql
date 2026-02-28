-- name: CreatePipelinesTable :exec
CREATE TABLE IF NOT EXISTS pipelines (
    pipeline_id VARCHAR(26) NOT NULL,
    plugin_id VARCHAR(26) NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    operation VARCHAR(64) NOT NULL,
    plugin_name VARCHAR(255) NOT NULL,
    handler VARCHAR(255) NOT NULL,
    priority INT NOT NULL DEFAULT 50,
    enabled TINYINT NOT NULL DEFAULT 1,
    config JSON NOT NULL,
    date_created DATETIME NOT NULL,
    date_modified DATETIME NOT NULL,
    PRIMARY KEY (pipeline_id),
    CONSTRAINT fk_pipeline_plugin FOREIGN KEY (plugin_id) REFERENCES plugins(plugin_id) ON DELETE CASCADE
);

-- name: CreatePipelinesIndexUnique :exec
CREATE UNIQUE INDEX idx_pipeline_unique ON pipelines(table_name, operation, plugin_id);

-- name: CreatePipelinesIndexPlugin :exec
CREATE INDEX idx_pipelines_plugin ON pipelines(plugin_id);

-- name: CreatePipelinesIndexTable :exec
CREATE INDEX idx_pipelines_table ON pipelines(table_name);

-- name: DropPipelinesTable :exec
DROP TABLE IF EXISTS pipelines;

-- name: GetPipeline :one
SELECT * FROM pipelines WHERE pipeline_id = ? LIMIT 1;

-- name: ListPipelines :many
SELECT * FROM pipelines ORDER BY table_name, operation, priority;

-- name: ListPipelinesByTable :many
SELECT * FROM pipelines WHERE table_name = ? ORDER BY operation, priority;

-- name: ListPipelinesByPluginID :many
SELECT * FROM pipelines WHERE plugin_id = ? ORDER BY table_name, operation, priority;

-- name: ListPipelinesByTableOperation :many
SELECT * FROM pipelines WHERE table_name = ? AND operation = ? ORDER BY priority, plugin_name;

-- name: ListEnabledPipelines :many
SELECT * FROM pipelines WHERE enabled = 1 ORDER BY table_name, operation, priority;

-- name: CreatePipeline :exec
INSERT INTO pipelines (pipeline_id, plugin_id, table_name, operation, plugin_name, handler, priority, enabled, config, date_created, date_modified)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdatePipeline :exec
UPDATE pipelines SET table_name = ?, operation = ?, handler = ?, priority = ?, enabled = ?, config = ?, date_modified = ? WHERE pipeline_id = ?;

-- name: UpdatePipelineEnabled :exec
UPDATE pipelines SET enabled = ?, date_modified = ? WHERE pipeline_id = ?;

-- name: DeletePipeline :exec
DELETE FROM pipelines WHERE pipeline_id = ?;

-- name: DeletePipelinesByPluginID :exec
DELETE FROM pipelines WHERE plugin_id = ?;

-- name: CountPipelines :one
SELECT COUNT(*) FROM pipelines;

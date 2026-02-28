CREATE TABLE IF NOT EXISTS pipelines (
    pipeline_id TEXT PRIMARY KEY NOT NULL CHECK (length(pipeline_id) = 26),
    plugin_id TEXT NOT NULL,
    table_name TEXT NOT NULL,
    operation TEXT NOT NULL,
    plugin_name TEXT NOT NULL,
    handler TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 50,
    enabled INTEGER NOT NULL DEFAULT 1,
    config TEXT NOT NULL DEFAULT '{}',
    date_created TEXT NOT NULL,
    date_modified TEXT NOT NULL,
    FOREIGN KEY (plugin_id) REFERENCES plugins(plugin_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_pipeline_unique ON pipelines(table_name, operation, plugin_id);
CREATE INDEX IF NOT EXISTS idx_pipelines_plugin ON pipelines(plugin_id);
CREATE INDEX IF NOT EXISTS idx_pipelines_table ON pipelines(table_name);

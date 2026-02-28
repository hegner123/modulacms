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

CREATE UNIQUE INDEX idx_pipeline_unique ON pipelines(table_name, operation, plugin_id);
CREATE INDEX idx_pipelines_plugin ON pipelines(plugin_id);
CREATE INDEX idx_pipelines_table ON pipelines(table_name);

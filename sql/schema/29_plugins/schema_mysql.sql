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

CREATE INDEX idx_plugins_status ON plugins(status);
CREATE INDEX idx_plugins_name ON plugins(name);

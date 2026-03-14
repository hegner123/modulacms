CREATE TABLE IF NOT EXISTS field_plugin_config (
    field_id         VARCHAR(26) PRIMARY KEY NOT NULL,
    plugin_name      VARCHAR(255) NOT NULL,
    plugin_interface VARCHAR(255) NOT NULL,
    plugin_version   VARCHAR(255) NOT NULL DEFAULT '',
    date_created     TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified    TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_fpc_field FOREIGN KEY (field_id) REFERENCES fields(field_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS admin_field_plugin_config (
    field_id         VARCHAR(26) PRIMARY KEY NOT NULL,
    plugin_name      VARCHAR(255) NOT NULL,
    plugin_interface VARCHAR(255) NOT NULL,
    plugin_version   VARCHAR(255) NOT NULL DEFAULT '',
    date_created     TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified    TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_afpc_field FOREIGN KEY (field_id) REFERENCES admin_fields(field_id) ON DELETE CASCADE
);

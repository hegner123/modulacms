CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id SERIAL PRIMARY KEY,
    content_data_id INTEGER NOT NULL,
    admin_field_id INTEGER NOT NULL,
    field_value TEXT NOT NULL,
    history TEXT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_content_data FOREIGN KEY (content_data_id)
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_fields FOREIGN KEY (admin_field_id)
        REFERENCES admin_fields(admin_field_id)
        ON UPDATE CASCADE ON DELETE CASCADE
);


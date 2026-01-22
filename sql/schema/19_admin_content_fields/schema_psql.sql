CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id SERIAL
        PRIMARY KEY,
    admin_route_id INTEGER
        CONSTRAINT fk_admin_route_id
            REFERENCES admin_routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    admin_content_data_id INTEGER NOT NULL
        CONSTRAINT fk_admin_content_data
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_id INTEGER NOT NULL
        CONSTRAINT fk_admin_fields
            REFERENCES admin_fields
            ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_value TEXT NOT NULL,
    author_id INTEGER NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_content_fields_route ON admin_content_fields(admin_route_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_fields_content ON admin_content_fields(admin_content_data_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_fields_field ON admin_content_fields(admin_field_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_fields_author ON admin_content_fields(author_id);

CREATE OR REPLACE FUNCTION update_admin_content_fields_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_admin_content_fields_modified_trigger
    BEFORE UPDATE ON admin_content_fields
    FOR EACH ROW
    EXECUTE FUNCTION update_admin_content_fields_modified();

CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id INTEGER 
        PRIMARY KEY ,
    admin_route_id INTEGER,
    admin_content_data_id INTEGER NOT NULL,
    admin_field_id INTEGER NOT NULL,
    admin_field_value TEXT NOT NULL,
    author_id INTEGER NOT NULL DEFAULT 0,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (admin_route_id) REFERENCES admin_routes(admin_route_id)
        ON DELETE SET NULL,
    FOREIGN KEY (admin_content_data_id) REFERENCES admin_content_data(admin_content_data_id)
        ON DELETE CASCADE,
    FOREIGN KEY (admin_field_id) REFERENCES admin_fields(admin_field_id)
        ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES users(user_id)
        ON DELETE SET DEFAULT
);

CREATE INDEX IF NOT EXISTS idx_admin_content_fields_route ON admin_content_fields(admin_route_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_fields_content ON admin_content_fields(admin_content_data_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_fields_field ON admin_content_fields(admin_field_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_fields_author ON admin_content_fields(author_id);

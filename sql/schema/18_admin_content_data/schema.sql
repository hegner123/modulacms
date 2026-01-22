CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id INTEGER PRIMARY KEY,
    parent_id INTEGER,
    first_child_id INTEGER,
    next_sibling_id INTEGER,
    prev_sibling_id INTEGER,
    admin_route_id INTEGER NOT NULL,
    admin_datatype_id INTEGER NOT NULL,
    author_id INTEGER NOT NULL DEFAULT 1,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (parent_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (first_child_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (next_sibling_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (prev_sibling_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (admin_route_id) REFERENCES admin_routes(admin_route_id) ON DELETE RESTRICT,
    FOREIGN KEY (admin_datatype_id) REFERENCES admin_datatypes(admin_datatype_id) ON DELETE RESTRICT,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE SET DEFAULT
);

CREATE INDEX IF NOT EXISTS idx_admin_content_data_parent ON admin_content_data(parent_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_data_route ON admin_content_data(admin_route_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_data_datatype ON admin_content_data(admin_datatype_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_data_author ON admin_content_data(author_id);

CREATE TRIGGER IF NOT EXISTS update_admin_content_data_modified
    AFTER UPDATE ON admin_content_data
    FOR EACH ROW
    BEGIN
        UPDATE admin_content_data SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE admin_content_data_id = NEW.admin_content_data_id;
    END;

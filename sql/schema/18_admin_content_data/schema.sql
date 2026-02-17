CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_content_data_id) = 26),
    parent_id TEXT,
    first_child_id TEXT,
    next_sibling_id TEXT,
    prev_sibling_id TEXT,
    admin_route_id TEXT NOT NULL,
    admin_datatype_id TEXT NOT NULL,
    author_id TEXT,
    status TEXT NOT NULL DEFAULT 'draft',
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (parent_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (first_child_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (next_sibling_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (prev_sibling_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (admin_route_id) REFERENCES admin_routes(admin_route_id) ON DELETE RESTRICT,
    FOREIGN KEY (admin_datatype_id) REFERENCES admin_datatypes(admin_datatype_id) ON DELETE RESTRICT,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE SET NULL
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

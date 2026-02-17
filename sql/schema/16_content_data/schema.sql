CREATE TABLE IF NOT EXISTS content_data (
    content_data_id TEXT PRIMARY KEY NOT NULL CHECK (length(content_data_id) = 26),
    parent_id TEXT,
    first_child_id TEXT,
    next_sibling_id TEXT,
    prev_sibling_id TEXT,
    route_id TEXT NOT NULL,
    datatype_id TEXT NOT NULL,
    author_id TEXT,
    status TEXT NOT NULL DEFAULT 'draft',
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (parent_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (first_child_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (next_sibling_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (prev_sibling_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (route_id) REFERENCES routes(route_id) ON DELETE RESTRICT,
    FOREIGN KEY (datatype_id) REFERENCES datatypes(datatype_id) ON DELETE RESTRICT,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_content_data_parent ON content_data(parent_id);
CREATE INDEX IF NOT EXISTS idx_content_data_route ON content_data(route_id);
CREATE INDEX IF NOT EXISTS idx_content_data_datatype ON content_data(datatype_id);
CREATE INDEX IF NOT EXISTS idx_content_data_author ON content_data(author_id);

CREATE TRIGGER IF NOT EXISTS update_content_data_modified
    AFTER UPDATE ON content_data
    FOR EACH ROW
    BEGIN
        UPDATE content_data SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE content_data_id = NEW.content_data_id;
    END;


CREATE TABLE IF NOT EXISTS role_permissions (
    id TEXT PRIMARY KEY NOT NULL CHECK (length(id) = 26),
    role_id TEXT NOT NULL REFERENCES roles(role_id) ON UPDATE CASCADE ON DELETE CASCADE,
    permission_id TEXT NOT NULL REFERENCES permissions(permission_id) ON UPDATE CASCADE ON DELETE CASCADE,
    UNIQUE(role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission ON role_permissions(permission_id);

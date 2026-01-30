CREATE TABLE IF NOT EXISTS admin_datatypes_fields (
    id TEXT PRIMARY KEY NOT NULL CHECK (length(id) = 26),
    admin_datatype_id TEXT NOT NULL
        CONSTRAINT fk_df_admin_datatype
            REFERENCES admin_datatypes
            ON DELETE CASCADE,
    admin_field_id TEXT NOT NULL
        CONSTRAINT fk_df_admin_field
            REFERENCES admin_fields
            ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_admin_datatypes_fields_datatype ON admin_datatypes_fields(admin_datatype_id);
CREATE INDEX IF NOT EXISTS idx_admin_datatypes_fields_field ON admin_datatypes_fields(admin_field_id);

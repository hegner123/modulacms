CREATE TABLE IF NOT EXISTS admin_datatypes_fields (
    id SERIAL
        PRIMARY KEY,
    admin_datatype_id INTEGER NOT NULL
        CONSTRAINT fk_df_admin_datatype
            REFERENCES admin_datatypes
            ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_id INTEGER NOT NULL
        CONSTRAINT fk_df_admin_field
            REFERENCES admin_fields
            ON UPDATE CASCADE ON DELETE CASCADE
);

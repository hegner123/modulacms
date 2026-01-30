CREATE TABLE IF NOT EXISTS datatypes_fields (
    id TEXT PRIMARY KEY NOT NULL,
    datatype_id TEXT NOT NULL
        CONSTRAINT fk_df_datatype
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_id TEXT NOT NULL
        CONSTRAINT fk_df_field
            REFERENCES fields
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_datatypes_fields_datatype ON datatypes_fields(datatype_id);
CREATE INDEX IF NOT EXISTS idx_datatypes_fields_field ON datatypes_fields(field_id);

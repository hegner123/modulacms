CREATE TABLE IF NOT EXISTS datatypes_fields (
    id INTEGER
        PRIMARY KEY,
    datatype_id INTEGER NOT NULL
        CONSTRAINT fk_df_datatype
            REFERENCES datatypes
            ON DELETE CASCADE,
    field_id INTEGER NOT NULL
        CONSTRAINT fk_df_field
            REFERENCES fields
            ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_datatypes_fields_datatype ON datatypes_fields(datatype_id);
CREATE INDEX IF NOT EXISTS idx_datatypes_fields_field ON datatypes_fields(field_id);

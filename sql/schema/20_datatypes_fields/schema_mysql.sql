CREATE TABLE IF NOT EXISTS datatypes_fields (
    id INT NOT NULL
        PRIMARY KEY,
    datatype_id INT NOT NULL,
    field_id INT NOT NULL,
    CONSTRAINT fk_df_datatype
        FOREIGN KEY (datatype_id) REFERENCES datatypes (datatype_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_df_field
        FOREIGN KEY (field_id) REFERENCES fields (field_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

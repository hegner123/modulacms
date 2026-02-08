CREATE TABLE IF NOT EXISTS admin_content_relations (
    admin_content_relation_id VARCHAR(26) NOT NULL,
    -- holds admin_content_data_id, named for code symmetry with content_relations
    source_content_id VARCHAR(26) NOT NULL,
    -- holds admin_content_data_id, named for code symmetry with content_relations
    target_content_id VARCHAR(26) NOT NULL,
    admin_field_id VARCHAR(26) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (admin_content_relation_id),
    CONSTRAINT chk_admin_content_relations_no_self_ref CHECK (source_content_id != target_content_id),
    CONSTRAINT fk_admin_content_relations_source FOREIGN KEY (source_content_id)
        REFERENCES admin_content_data(admin_content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_relations_target FOREIGN KEY (target_content_id)
        REFERENCES admin_content_data(admin_content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_relations_field FOREIGN KEY (admin_field_id)
        REFERENCES admin_fields(admin_field_id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT uq_admin_content_relations_unique UNIQUE (source_content_id, admin_field_id, target_content_id)
);

CREATE INDEX idx_admin_content_relations_target ON admin_content_relations(target_content_id, date_created);
CREATE INDEX idx_admin_content_relations_field ON admin_content_relations(admin_field_id);

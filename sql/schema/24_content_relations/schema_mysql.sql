CREATE TABLE IF NOT EXISTS content_relations (
    content_relation_id VARCHAR(26) NOT NULL,
    source_content_id VARCHAR(26) NOT NULL,
    target_content_id VARCHAR(26) NOT NULL,
    field_id VARCHAR(26) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (content_relation_id),
    CONSTRAINT fk_content_relations_source FOREIGN KEY (source_content_id)
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_content_relations_target FOREIGN KEY (target_content_id)
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_content_relations_field FOREIGN KEY (field_id)
        REFERENCES fields(field_id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT uq_content_relations_unique UNIQUE (source_content_id, field_id, target_content_id)
);

CREATE INDEX idx_content_relations_target ON content_relations(target_content_id, date_created);
CREATE INDEX idx_content_relations_field ON content_relations(field_id);

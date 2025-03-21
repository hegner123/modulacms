CREATE TABLE IF NOT EXISTS admin_datatypes (
    admin_datatype_id SERIAL PRIMARY KEY,
    
    parent_id INT DEFAULT NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author TEXT NOT NULL,
    author_id INT NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT,
    CONSTRAINT fk_parent_id FOREIGN KEY (parent_id)
        REFERENCES admin_datatypes(admin_dt_id)
        ON UPDATE CASCADE
        ON DELETE SET DEFAULT,
    CONSTRAINT fk_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE
        ON DELETE SET DEFAULT,
    CONSTRAINT fk_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE SET DEFAULT
);


CREATE TABLE IF NOT EXISTS admin_datatypes (
    admin_dt_id INT NOT NULL AUTO_INCREMENT,
    admin_route_id INT DEFAULT NULL,
    parent_id INT DEFAULT NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author TEXT NOT NULL,
    author_id INT NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT,
    PRIMARY KEY (admin_dt_id),
    CONSTRAINT fk_admin_route_id FOREIGN KEY (admin_route_id)
        REFERENCES admin_routes(admin_route_id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION,
    CONSTRAINT fk_parent_id FOREIGN KEY (parent_id)
        REFERENCES admin_datatypes(admin_dt_id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION,
    CONSTRAINT fk_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE
        ON DELETE NO ACTION,
    CONSTRAINT fk_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


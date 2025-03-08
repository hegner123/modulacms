CREATE TABLE IF NOT EXISTS content_data (
    content_data_id INT AUTO_INCREMENT PRIMARY KEY,
    route_id    INT DEFAULT NULL,
    datatype_id INT DEFAULT NULL,
    history TEXT DEFAULT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_content_data_route_id FOREIGN KEY (route_id)
        REFERENCES routes(route_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_datatypes FOREIGN KEY (datatype_id)
        REFERENCES datatypes(datatype_id)
        ON UPDATE CASCADE ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


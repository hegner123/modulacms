CREATE TABLE IF NOT EXISTS media_dimensions (
    md_id INT AUTO_INCREMENT PRIMARY KEY,
    label VARCHAR(255) UNIQUE,
    width INT,
    height INT,
    aspect_ratio TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


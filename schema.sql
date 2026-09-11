CREATE TABLE lookup_logs (
    id                 BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    ip                 VARCHAR(45) NOT NULL,
    requested_at       DATETIME(3) NOT NULL,
    status             ENUM('cache_hit', 'cache_miss', 'private_ip') NOT NULL,
    ip_api_duration_ms INT UNSIGNED NULL,
    country            VARCHAR(100) NULL,
    country_code       VARCHAR(10) NULL,
    city               VARCHAR(100) NULL,
    region_name        VARCHAR(100) NULL,
    INDEX idx_requested_at (requested_at),
    INDEX idx_ip (ip)
);

CREATE TABLE `audit_log` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `action` VARCHAR(255) NOT NULL DEFAULT '',
  `user_id` BIGINT NOT NULL DEFAULT 0,
  `patient_id` BIGINT NOT NULL DEFAULT 0,
  `resource_type` VARCHAR(255) NOT NULL DEFAULT '',
  `resource_id` BIGINT NOT NULL DEFAULT 0,
  `ip_address` VARCHAR(45) NOT NULL DEFAULT '',
  `user_agent` VARCHAR(512) NOT NULL DEFAULT '',
  `details` TEXT,
  `success` TINYINT(1) NOT NULL DEFAULT 1
);

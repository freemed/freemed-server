CREATE TABLE `record_lock` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `table_name` VARCHAR(255) NOT NULL DEFAULT '',
  `record_id` BIGINT NOT NULL DEFAULT 0,
  `user_id` BIGINT NOT NULL DEFAULT 0,
  `locked_at` DATETIME NOT NULL,
  `expires_at` DATETIME NOT NULL
);

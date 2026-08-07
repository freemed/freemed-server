CREATE TABLE `user_preferences` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `option_key` VARCHAR(255) NOT NULL,
  `default_value` VARCHAR(255) NOT NULL DEFAULT '',
  `title` VARCHAR(255) NOT NULL DEFAULT '',
  `section` VARCHAR(255) NOT NULL DEFAULT '',
  `option_type` VARCHAR(255) NOT NULL DEFAULT '',
  `options` TEXT,
  UNIQUE KEY `idx_option_key` (`option_key`)
);

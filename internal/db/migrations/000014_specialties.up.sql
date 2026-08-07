CREATE TABLE `specialties` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `name` VARCHAR(255) NOT NULL DEFAULT '',
  `display_value` VARCHAR(255) NOT NULL DEFAULT '',
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

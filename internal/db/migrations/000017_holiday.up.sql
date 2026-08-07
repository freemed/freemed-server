CREATE TABLE `holiday` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `holiday_date` DATETIME NOT NULL,
  `holiday_name` VARCHAR(255) NOT NULL DEFAULT '',
  `description` VARCHAR(255) NOT NULL DEFAULT '',
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

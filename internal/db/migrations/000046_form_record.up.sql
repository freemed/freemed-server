CREATE TABLE `form_record` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `fr_id` BIGINT NOT NULL,
  `fr_uuid` VARCHAR(255) NOT NULL DEFAULT '',
  `fr_name` VARCHAR(255) NOT NULL DEFAULT '',
  `fr_value` TEXT
);

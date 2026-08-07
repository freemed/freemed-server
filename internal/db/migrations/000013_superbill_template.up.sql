CREATE TABLE `superbill_template` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `name` VARCHAR(255) NOT NULL DEFAULT '',
  `template_data` LONGBLOB,
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `form_templates` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `name` VARCHAR(255) NOT NULL DEFAULT '',
  `description` TEXT,
  `form_type` VARCHAR(100) NOT NULL DEFAULT 'encounter',
  `template_data` LONGTEXT,
  `is_default` TINYINT(1) NOT NULL DEFAULT 0,
  `user` BIGINT NOT NULL DEFAULT 0
);

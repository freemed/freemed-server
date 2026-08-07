CREATE TABLE `tools` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `tool_name` VARCHAR(255) NOT NULL DEFAULT '',
  `tool_description` TEXT,
  `tool_class` VARCHAR(255) NOT NULL DEFAULT '',
  `tool_parameters` TEXT,
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

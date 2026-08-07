CREATE TABLE `previous_operations` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `operation_date` DATE,
  `operation` VARCHAR(255) NOT NULL DEFAULT '',
  `user` BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE `form_results` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `fr_patient` BIGINT NOT NULL,
  `fr_timestamp` DATETIME NOT NULL,
  `fr_template` VARCHAR(50) NOT NULL DEFAULT '',
  `fr_formname` VARCHAR(50) NOT NULL DEFAULT '',
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT 'active'
);

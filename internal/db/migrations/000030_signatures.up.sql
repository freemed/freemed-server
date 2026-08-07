CREATE TABLE `signatures` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `module` VARCHAR(255) NOT NULL DEFAULT '',
  `module_field` VARCHAR(255) NOT NULL DEFAULT '',
  `oid` BIGINT NOT NULL DEFAULT 0,
  `signature_data` LONGBLOB,
  `format` VARCHAR(50) NOT NULL DEFAULT 'UNKNOWN',
  `collector_location` VARCHAR(100) NOT NULL DEFAULT '',
  `collector_model` VARCHAR(100) NOT NULL DEFAULT '',
  `collector_jobid` VARCHAR(100) NOT NULL DEFAULT '',
  `collector_finished` TINYINT(1) NOT NULL DEFAULT 0,
  `user` BIGINT NOT NULL DEFAULT 0
);

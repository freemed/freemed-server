CREATE TABLE `family_history` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `relationship` VARCHAR(255) NOT NULL DEFAULT '',
  `condition_name` VARCHAR(255) NOT NULL DEFAULT '',
  `icd10_code` VARCHAR(20) NOT NULL DEFAULT '',
  `onset_age` INT NOT NULL DEFAULT 0,
  `deceased` TINYINT(1) NOT NULL DEFAULT 0,
  `notes` TEXT,
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT 'active',
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME
);

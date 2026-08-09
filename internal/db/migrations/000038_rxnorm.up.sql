CREATE TABLE `rxnorm` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `rxcui` VARCHAR(255) NOT NULL DEFAULT '',
  `drug_name` VARCHAR(255) NOT NULL DEFAULT '',
  `drug_class` VARCHAR(255) NOT NULL DEFAULT ''
);

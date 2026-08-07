CREATE TABLE `scanned_docs` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `filename` VARCHAR(255) NOT NULL DEFAULT '',
  `description` VARCHAR(255) NOT NULL DEFAULT '',
  `page_count` BIGINT NOT NULL DEFAULT 0,
  `document_date` DATETIME,
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT 'active'
);

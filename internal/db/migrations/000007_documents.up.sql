CREATE TABLE `unfiled_docs` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `filename` VARCHAR(255) NOT NULL DEFAULT '',
  `page_count` BIGINT NOT NULL DEFAULT 0,
  `file_type` VARCHAR(255) NOT NULL DEFAULT '',
  `received_date` DATETIME,
  `assigned_to` BIGINT NOT NULL DEFAULT 0,
  `status` VARCHAR(255) NOT NULL DEFAULT '',
  `notes` TEXT,
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `unread_docs` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `document_type` VARCHAR(255) NOT NULL DEFAULT '',
  `patient` BIGINT NOT NULL DEFAULT 0,
  `filename` VARCHAR(255) NOT NULL DEFAULT '',
  `page_count` BIGINT NOT NULL DEFAULT 0,
  `sent_date` DATETIME,
  `sending_provider` BIGINT NOT NULL DEFAULT 0,
  `assigned_to` BIGINT NOT NULL DEFAULT 0,
  `reviewed` TINYINT(1) NOT NULL DEFAULT 0,
  `review_date` DATETIME,
  `notes` TEXT,
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

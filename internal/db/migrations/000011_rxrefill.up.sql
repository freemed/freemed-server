CREATE TABLE `rxrefillrequest` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `user` BIGINT NOT NULL DEFAULT 0,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `provider` BIGINT NOT NULL DEFAULT 0,
  `rxorig` TEXT,
  `note` VARCHAR(250) NOT NULL DEFAULT '',
  `approved` DATETIME,
  `locked` BIGINT NOT NULL DEFAULT 0
);

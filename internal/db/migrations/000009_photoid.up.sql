CREATE TABLE `photoid` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `photo` LONGBLOB,
  `photo_mime` VARCHAR(255) NOT NULL DEFAULT '',
  `page_count` BIGINT NOT NULL DEFAULT 1,
  `description` VARCHAR(255) NOT NULL DEFAULT '',
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT 'active'
);

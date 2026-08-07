CREATE TABLE `letters` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `letter_type` VARCHAR(255) NOT NULL DEFAULT '',
  `recipient` VARCHAR(255) NOT NULL DEFAULT '',
  `subject` VARCHAR(255) NOT NULL DEFAULT '',
  `body` TEXT,
  `date_sent` DATETIME,
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT '',
  FOREIGN KEY (`patient`) REFERENCES `patient`(`id`)
);

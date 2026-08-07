CREATE TABLE `patient_correspondence` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `correspondence_type` VARCHAR(255) NOT NULL DEFAULT '',
  `direction` VARCHAR(255) NOT NULL DEFAULT '',
  `contact_name` VARCHAR(255) NOT NULL DEFAULT '',
  `contact_method` VARCHAR(255) NOT NULL DEFAULT '',
  `summary` TEXT,
  `date` DATETIME,
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT '',
  FOREIGN KEY (`patient`) REFERENCES `patient`(`id`)
);

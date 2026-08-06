CREATE TABLE `prescriptions` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `drug_name` VARCHAR(255) NOT NULL DEFAULT '',
  `dosage` VARCHAR(255) NOT NULL DEFAULT '',
  `frequency` VARCHAR(255) NOT NULL DEFAULT '',
  `quantity` VARCHAR(255) NOT NULL DEFAULT '',
  `refills` BIGINT NOT NULL DEFAULT 0,
  `date_written` DATETIME NOT NULL,
  `prescribing_provider` BIGINT NOT NULL DEFAULT 0,
  `pharmacy` VARCHAR(255) NOT NULL DEFAULT '',
  `status` VARCHAR(255) NOT NULL DEFAULT 'active',
  `notes` TEXT,
  `user` BIGINT NOT NULL DEFAULT 0,
  FOREIGN KEY (`patient`) REFERENCES `patient`(`id`)
);

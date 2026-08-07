CREATE TABLE `clinical_orders` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `order_type` VARCHAR(255) NOT NULL DEFAULT '',
  `description` VARCHAR(255) NOT NULL DEFAULT '',
  `status` VARCHAR(255) NOT NULL DEFAULT '',
  `date_ordered` DATETIME,
  `ordering_provider` BIGINT NOT NULL DEFAULT 0,
  `notes` TEXT,
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT '',
  FOREIGN KEY (`patient`) REFERENCES `patient`(`id`)
);

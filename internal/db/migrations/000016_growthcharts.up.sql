CREATE TABLE `growthcharts` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `record_date` DATETIME NOT NULL,
  `age_months` DECIMAL(10,2),
  `height_cm` DECIMAL(10,2),
  `weight_kg` DECIMAL(10,2),
  `head_circumference_cm` DECIMAL(10,2),
  `bmi` DECIMAL(10,2),
  `notes` VARCHAR(255) NOT NULL DEFAULT '',
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT ''
);

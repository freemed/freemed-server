CREATE TABLE `chronic_problems` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `date` DATE NOT NULL,
  `problem` VARCHAR(250) NOT NULL DEFAULT '',
  `user` BIGINT NOT NULL DEFAULT 0
);

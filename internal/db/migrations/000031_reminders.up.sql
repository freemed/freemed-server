CREATE TABLE `reminders` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `user` BIGINT NOT NULL DEFAULT 0,
  `patient` BIGINT,
  `title` VARCHAR(255) NOT NULL DEFAULT '',
  `description` TEXT,
  `due_date` DATETIME,
  `priority` INT NOT NULL DEFAULT 0,
  `status` VARCHAR(50) NOT NULL DEFAULT 'pending',
  `completed_at` DATETIME
);

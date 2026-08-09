CREATE TABLE `room` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `name` VARCHAR(255) NOT NULL DEFAULT '',
  `facility_id` BIGINT NOT NULL DEFAULT 0,
  `room_type` VARCHAR(255) NOT NULL DEFAULT '',
  `active` VARCHAR(255) NOT NULL DEFAULT 'active'
);

CREATE TABLE `certifications` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `patient` BIGINT NOT NULL DEFAULT 0,
  `cert_type` BIGINT NOT NULL DEFAULT 0,
  `cert_form_num` BIGINT,
  `cert_desc` VARCHAR(255) NOT NULL DEFAULT '',
  `cert_form_data` TEXT,
  `user` BIGINT NOT NULL DEFAULT 0,
  `active` VARCHAR(255) NOT NULL DEFAULT 'active'
);

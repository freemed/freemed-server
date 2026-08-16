CREATE TABLE `schedulingrules` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `user` BIGINT NOT NULL DEFAULT 0,
  `provider` VARCHAR(250),
  `reason` VARCHAR(150),
  `dowbegin` INT,
  `dowend` INT,
  `datebegin` DATE,
  `dateend` DATE,
  `timebegin` VARCHAR(8),
  `timeend` VARCHAR(8),
  `newpatient` TINYINT(1)
);

CREATE TABLE `fhir_client` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `client_id` VARCHAR(255) NOT NULL DEFAULT '',
  `client_name` VARCHAR(255) NOT NULL DEFAULT '',
  `redirect_uris` TEXT NOT NULL,
  `public_key` TEXT,
  `grant_types` VARCHAR(255) NOT NULL DEFAULT 'authorization_code',
  `scopes` VARCHAR(255) NOT NULL DEFAULT 'launch patient/*.read',
  `is_confidential` TINYINT(1) NOT NULL DEFAULT 0,
  `active` TINYINT(1) NOT NULL DEFAULT 1
);

CREATE TABLE `fhir_auth_code` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `code` VARCHAR(255) NOT NULL DEFAULT '',
  `client_id` VARCHAR(255) NOT NULL DEFAULT '',
  `user_id` BIGINT NOT NULL DEFAULT 0,
  `patient_id` BIGINT NOT NULL DEFAULT 0,
  `scopes` VARCHAR(255) NOT NULL DEFAULT '',
  `redirect_uri` VARCHAR(255) NOT NULL DEFAULT '',
  `expires_at` DATETIME NOT NULL,
  `used` TINYINT(1) NOT NULL DEFAULT 0
);

CREATE TABLE `fhir_access_token` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `token_hash` VARCHAR(255) NOT NULL DEFAULT '',
  `client_id` VARCHAR(255) NOT NULL DEFAULT '',
  `user_id` BIGINT NOT NULL DEFAULT 0,
  `patient_id` BIGINT NOT NULL DEFAULT 0,
  `scopes` VARCHAR(255) NOT NULL DEFAULT '',
  `expires_at` DATETIME NOT NULL
);

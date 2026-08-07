CREATE TABLE `usergroup` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `groupname` VARCHAR(255) NOT NULL DEFAULT '',
  `groupdescrip` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `user_groups` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `user_id` BIGINT NOT NULL DEFAULT 0,
  `group_id` BIGINT NOT NULL DEFAULT 0,
  UNIQUE KEY `uq_user_group` (`user_id`, `group_id`)
);

CREATE TABLE `acl_permissions` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `permission_name` VARCHAR(255) NOT NULL DEFAULT '',
  `permission_desc` VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE `group_permissions` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `group_id` BIGINT NOT NULL DEFAULT 0,
  `permission_id` BIGINT NOT NULL DEFAULT 0,
  UNIQUE KEY `uq_group_permission` (`group_id`, `permission_id`)
);

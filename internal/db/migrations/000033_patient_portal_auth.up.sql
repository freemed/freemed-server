ALTER TABLE patient
  ADD COLUMN portal_password VARCHAR(255) NOT NULL DEFAULT '' AFTER pemail,
  ADD COLUMN portal_pin VARCHAR(255) NOT NULL DEFAULT '' AFTER portal_password,
  ADD COLUMN portal_enabled TINYINT(1) NOT NULL DEFAULT 0 AFTER portal_pin,
  ADD COLUMN portal_last_login DATETIME AFTER portal_enabled,
  ADD COLUMN portal_failed_attempts INT NOT NULL DEFAULT 0 AFTER portal_last_login;

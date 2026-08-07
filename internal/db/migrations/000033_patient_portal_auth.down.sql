ALTER TABLE patient
  DROP COLUMN portal_password,
  DROP COLUMN portal_pin,
  DROP COLUMN portal_enabled,
  DROP COLUMN portal_last_login,
  DROP COLUMN portal_failed_attempts;

-- name: ListUserPreferences :many
SELECT * FROM user_preferences ORDER BY section, option_key;

-- name: GetUserPreferenceByKey :one
SELECT * FROM user_preferences WHERE option_key = sqlc.arg(option_key);

-- name: UpsertUserPreference :exec
INSERT INTO user_preferences (option_key, default_value, title, section, option_type, options)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  default_value = VALUES(default_value),
  title = VALUES(title),
  section = VALUES(section),
  option_type = VALUES(option_type),
  options = VALUES(options);

-- name: GetFhirClientByID :one
SELECT * FROM fhir_client WHERE client_id = ? AND active = 1;

-- name: InsertFhirAuthCode :execresult
INSERT INTO fhir_auth_code (code, client_id, user_id, patient_id, scopes, redirect_uri, expires_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, NOW() + INTERVAL 5 MINUTE, NOW(), NOW());

-- name: GetFhirAuthCode :one
SELECT * FROM fhir_auth_code WHERE code = ? AND used = 0 AND expires_at > NOW();

-- name: MarkFhirAuthCodeUsed :exec
UPDATE fhir_auth_code SET used = 1, updated_at = NOW() WHERE id = ?;

-- name: InsertFhirAccessToken :execresult
INSERT INTO fhir_access_token (token_hash, client_id, user_id, patient_id, scopes, expires_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, NOW() + INTERVAL 1 HOUR, NOW(), NOW());

-- name: GetFhirAccessToken :one
SELECT * FROM fhir_access_token WHERE token_hash = ? AND expires_at > NOW();

-- name: ListFhirClients :many
SELECT * FROM fhir_client WHERE deleted_at IS NULL ORDER BY client_name;

-- name: CreateFhirClient :execresult
INSERT INTO fhir_client (client_id, client_name, redirect_uris, grant_types, scopes, is_confidential, active, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, 1, NOW(), NOW());

-- name: DeactivateFhirClient :exec
UPDATE fhir_client SET active = 0, updated_at = NOW() WHERE id = ?;

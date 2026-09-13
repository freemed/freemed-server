-- name: GetFhirClientByID :one
SELECT * FROM fhir_client WHERE client_id = ? AND active = 1;

-- name: InsertFhirAuthCode :execresult
INSERT INTO fhir_auth_code (code, client_id, user_id, patient_id, scopes, redirect_uri, expires_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, NOW() + INTERVAL 5 MINUTE, NOW(), NOW());

-- name: GetFhirAuthCode :one
SELECT * FROM fhir_auth_code WHERE code = ? AND used = 0 AND expires_at > NOW();

-- name: ConsumeFhirAuthCode :execrows
-- Conditional single-use consumption: the UPDATE only matches while the code is
-- unused, so two concurrent redemptions of the same code cannot both mint a
-- token. The caller must treat 0 rows affected as "already redeemed".
UPDATE fhir_auth_code SET used = 1, updated_at = NOW() WHERE id = ? AND used = 0;

-- name: InsertFhirAccessToken :execresult
INSERT INTO fhir_access_token (token_hash, client_id, user_id, patient_id, scopes, expires_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, NOW() + INTERVAL 1 HOUR, NOW(), NOW());

-- name: GetFhirAccessToken :one
SELECT * FROM fhir_access_token WHERE token_hash = ? AND expires_at > NOW();

-- name: ListFhirClients :many
-- client_secret_hash is deliberately excluded: these rows are serialized
-- straight to the admin UI, and a stored credential hash must not leave the
-- database.
SELECT id, created_at, updated_at, deleted_at, client_id, client_name, redirect_uris,
       public_key, grant_types, scopes, is_confidential, active
FROM fhir_client WHERE deleted_at IS NULL ORDER BY client_name;

-- name: CreateFhirClient :execresult
INSERT INTO fhir_client (client_id, client_name, redirect_uris, grant_types, scopes, is_confidential, client_secret_hash, active, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, 1, NOW(), NOW());

-- name: DeactivateFhirClient :exec
UPDATE fhir_client SET active = 0, updated_at = NOW() WHERE id = ?;

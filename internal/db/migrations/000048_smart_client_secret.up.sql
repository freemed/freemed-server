-- client_credentials authentication (H4).
--
-- fhir_client previously had no credential column at all, so
-- handleClientCredentialsGrant discarded the client secret it was handed and a
-- confidential client's client_id alone minted a one-hour access token. Only a
-- bcrypt hash of the secret is stored; the plaintext is returned once, at
-- creation time (POST /api/smart/clients).
ALTER TABLE `fhir_client`
  ADD COLUMN `client_secret_hash` VARCHAR(255) NOT NULL DEFAULT '' AFTER `public_key`;

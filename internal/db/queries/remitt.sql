-- Patients with unbilled procedures
-- name: PatientsToBill :many
SELECT DISTINCT p.id, p.ptlname, p.ptfname
FROM procrec pr
JOIN patient p ON pr.procpatient = p.id
WHERE pr.procstatus = 'unbilled'
  AND pr.active = 'active';

-- Unbilled procedures for a specific patient
-- name: ProceduresToBill :many
SELECT pr.*, c.cptnameext
FROM procrec pr
LEFT JOIN cpt c ON pr.proccpt = c.id
WHERE pr.procstatus = 'unbilled'
  AND pr.procpatient = sqlc.arg(patient_id)
  AND pr.active = 'active';

-- Claim info: all billkey entries ordered by date
-- name: GetClaimInfo :many
SELECT * FROM billkey ORDER BY billkeydate DESC;

-- Mark a billkey entry as processed (update timestamp)
-- name: MarkAsBilled :exec
UPDATE billkey SET updated_at = NOW() WHERE id = sqlc.arg(id);

-- Rebill list: all billkey entries (candidates for rebilling)
-- name: GetRebillList :many
SELECT * FROM billkey ORDER BY billkeydate DESC;

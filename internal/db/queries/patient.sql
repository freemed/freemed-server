-- Patient EMR attachments: list all EMR entries for a patient
-- name: PatientEmrAttachments :many
SELECT
  p.patient,
  p.module,
  p.oid,
  p.annotation,
  p.summary,
  p.stamp,
  DATE_FORMAT(p.stamp, '%m/%d/%Y') AS date_mdy,
  m.module_name AS type,
  m.module_class AS module_namespace,
  p.locked,
  p.id
FROM patient_emr p
LEFT OUTER JOIN modules m ON m.module_table = p.module
WHERE p.patient = sqlc.arg(patient_id)
  AND m.module_hidden = 0;

-- Patient EMR attachments filtered by module
-- name: PatientEmrAttachmentsByModule :many
SELECT
  p.patient,
  p.module,
  p.oid,
  p.annotation,
  p.summary,
  p.stamp,
  DATE_FORMAT(p.stamp, '%m/%d/%Y') AS date_mdy,
  m.module_name AS type,
  m.module_class AS module_namespace,
  p.locked,
  p.id
FROM patient_emr p
LEFT OUTER JOIN modules m ON m.module_table = p.module
WHERE p.patient = sqlc.arg(patient_id)
  AND p.module = sqlc.arg(module)
  AND m.module_hidden = 0;

-- Patient progress notes listing with author name
-- name: PatientProgressNotes :many
SELECT
  pn.id,
  pn.pnotesdt AS date,
  pn.pnotesdescrip AS description,
  pn.pnotes_S AS soap_subjective,
  pn.pnotes_O AS soap_objective,
  pn.pnotes_A AS soap_assessment,
  pn.pnotes_P AS soap_plan,
  CONCAT(u.userfname, ' ', u.userlname) AS author_name,
  pn.active
FROM pnotes pn
LEFT OUTER JOIN user u ON u.id = pn.user
WHERE pn.pnotespat = sqlc.arg(patient_id)
ORDER BY pn.pnotesdt DESC;

-- Patient information: full patient details with joins
-- name: PatientInformation :one
SELECT
  CONCAT(p.ptlname, ', ', p.ptfname, IF(p.ptmname IS NOT NULL, CONCAT(' ', p.ptmname), '')) AS patient_name,
  p.ptid AS patient_id,
  p.ptdob AS date_of_birth,
  p.ptprimarylanguage AS language,
  DATE_FORMAT(p.ptdob, '%m/%d/%Y') AS date_of_birth_mdy,
  CASE
    WHEN ((TO_DAYS(NOW()) - TO_DAYS(p.ptdob)) / 365) >= 2
    THEN CONCAT(FLOOR((TO_DAYS(NOW()) - TO_DAYS(p.ptdob)) / 365), ' years')
    ELSE CONCAT(FLOOR((TO_DAYS(NOW()) - TO_DAYS(p.ptdob)) / 30), ' months')
  END AS age,
  pa.line1 AS address_line_1,
  pa.line2 AS address_line_2,
  pa.city AS city,
  pa.stpr AS state,
  pa.postal AS postal,
  CONCAT(pa.city, ', ', pa.stpr, ' ', pa.postal) AS csz,
  CASE
    WHEN p.id IN (
      SELECT al.patient FROM allergies al
      WHERE al.patient = sqlc.arg(patient_id) AND al.active = 'active'
    )
    THEN 'true' ELSE 'false'
  END AS hasallergy,
  CONCAT(phy.phylname, ', ', phy.phyfname, ' ', phy.phymname) AS pcp,
  CONCAT(fac.psrname, ' (', fac.psrcity, ', ', fac.psrstate, ')') AS facility,
  CONCAT(ph.phname, ' (', ph.phcity, ', ', ph.phstate, ')') AS pharmacy
FROM patient p
LEFT OUTER JOIN patient_address pa ON pa.patient = p.id AND pa.active = TRUE
LEFT OUTER JOIN physician phy ON phy.id = p.ptpcp
LEFT OUTER JOIN facility fac ON fac.id = p.ptprimaryfacility
LEFT OUTER JOIN pharmacy ph ON ph.id = p.ptpharmacy
WHERE p.id = sqlc.arg(patient_id)
GROUP BY p.id;

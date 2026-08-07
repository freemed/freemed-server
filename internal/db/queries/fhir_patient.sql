-- FHIR Patient: get patient by ID with address join for FHIR Patient resource
-- name: FhirPatientById :one
SELECT
  p.id,
  p.ptlname,
  p.ptfname,
  p.ptmname,
  p.ptsuffix,
  p.ptsalut,
  p.ptsex,
  p.ptdob,
  p.ptid,
  p.pemail,
  p.ptblood,
  p.ssn,
  p.ptprimarylanguage,
  p.ptdead,
  p.ptdeaddt,
  p.ptpcp,
  p.created_at,
  p.updated_at,
  pa.line1   AS address_line1,
  pa.line2   AS address_line2,
  pa.city    AS address_city,
  pa.stpr    AS address_state,
  pa.postal  AS address_postal
FROM patient p
LEFT OUTER JOIN patient_address pa ON pa.patient = p.id AND pa.active = 1
WHERE p.id = sqlc.arg(patient_id)
  AND p.ptarchive = 0;

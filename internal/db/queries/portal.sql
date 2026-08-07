-- Portal: patient demographics for the "me" endpoint
-- name: GetPortalPatientDemographics :one
SELECT
  id,
  ptfname AS first_name,
  ptlname AS last_name,
  ptmname AS middle_name,
  ptsuffix AS suffix,
  ptid AS patient_id_display,
  ptdob AS date_of_birth,
  ptsex AS gender,
  ptprimarylanguage AS language,
  pemail AS email,
  portal_enabled
FROM patient
WHERE id = sqlc.arg(patient_id)
  AND ptarchive = 0;

-- Portal: list appointments for a patient from scheduler
-- name: ListPortalAppointments :many
SELECT
  s.id,
  s.caldateof AS date_of,
  DATE_FORMAT(s.caldateof, '%m/%d/%Y') AS date_of_mdy,
  s.calhour AS hour,
  s.calminute AS minute,
  CONCAT(LPAD(s.calhour, 2, '0'), ':', LPAD(s.calminute, 2, '0')) AS appointment_time,
  s.calduration AS duration,
  s.calphysician AS provider_id,
  CONCAT(ph.phylname, ', ', ph.phyfname) AS provider_name,
  s.caltype AS appointment_type,
  s.calstatus AS status,
  s.calprenote AS note,
  s.calappttemplate AS appointment_template_id,
  s.calfacility AS facility_id,
  s.calroom AS room_id,
  s.created_at,
  s.updated_at
FROM scheduler s
LEFT JOIN physician ph ON s.calphysician = ph.id
WHERE s.calpatient = sqlc.arg(patient_id)
  AND s.caltype = 'patient'
ORDER BY s.caldateof DESC, s.calhour DESC, s.calminute DESC;

-- Portal: create an appointment request (status = 'requested')
-- name: CreatePortalAppointmentRequest :execresult
INSERT INTO scheduler (
  caldateof, calhour, calminute, calduration, caltype,
  calphysician, calpatient, calstatus, calprenote,
  calcreated, calappttemplate, user,
  created_at, updated_at
) VALUES (
  sqlc.arg(date_of), sqlc.arg(hour), sqlc.arg(minute), 30, 'patient',
  sqlc.arg(provider_id), sqlc.arg(patient_id), 'requested', sqlc.arg(note),
  NOW(), sqlc.narg(appointment_template_id), sqlc.arg(user_id),
  NOW(), NOW()
);

-- Portal: list problems (union of current and chronic)
-- name: ListPortalProblems :many
SELECT
  'current' AS problem_type,
  cp.id,
  cp.patient,
  cp.date,
  cp.problem AS description,
  cp.active,
  cp.created_at,
  cp.updated_at
FROM current_problems cp
WHERE cp.patient = sqlc.arg(patient_id)
  AND cp.active = 'active'
  AND cp.deleted_at IS NULL
UNION ALL
SELECT
  'chronic' AS problem_type,
  chp.id,
  chp.patient,
  chp.date,
  chp.problem AS description,
  NULL AS active,
  chp.created_at,
  chp.updated_at
FROM chronic_problems chp
WHERE chp.patient = sqlc.arg(patient_id)
  AND chp.deleted_at IS NULL
ORDER BY date DESC;

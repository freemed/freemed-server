-- Combined timeline of encounters, procedures, and notes for a patient
-- name: PatientHistory :many
SELECT 'encounter' AS entry_type, pe.stamp AS event_date, pe.summary AS description, pe.id
FROM patient_emr pe WHERE pe.patient = sqlc.arg(patient_id)
UNION ALL
SELECT 'procedure' AS entry_type, pr.procdt AS event_date, c.cptnameext AS description, pr.id
FROM procrec pr LEFT JOIN cpt c ON pr.proccpt = c.id WHERE pr.procpatient = sqlc.arg(patient_id)
UNION ALL
SELECT 'note' AS entry_type, pn.pnotesdt AS event_date, pn.pnotesdescrip AS description, pn.id
FROM pnotes pn WHERE pn.pnotespat = sqlc.arg(patient_id)
ORDER BY event_date DESC LIMIT 50;

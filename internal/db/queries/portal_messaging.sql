-- Portal: list messages for a patient (no user_id filter needed for portal)
-- name: PortalListMessages :many
SELECT
  m.id,
  m.msgtime AS message_time,
  m.msgsubject AS subject,
  m.msgtext AS body,
  m.msgby AS sent_by_user_id,
  m.sender,
  m.msgread AS is_read,
  m.msgurgency AS urgency,
  m.msgtag AS tag,
  m.created_at,
  m.updated_at
FROM messages m
WHERE m.msgpatient = sqlc.arg(patient_id)
  AND (m.msgtag IS NULL OR LENGTH(m.msgtag) < 1)
ORDER BY m.msgtime DESC;

-- Portal: send a message from patient to practice
-- name: PortalCreateMessage :execresult
INSERT INTO messages (
  msgby, sender, msgtime, msgfor, msgrecip, msgpatient, msgperson,
  msgurgency, msgsubject, msgtext, msgread, msgunique, msgtag, active,
  created_at, updated_at
) VALUES (
  0, sqlc.arg(sender), NOW(), sqlc.arg(msg_for), 'Practice',
  sqlc.arg(patient_id), sqlc.arg(patient_name),
  sqlc.arg(urgency), sqlc.arg(subject), sqlc.arg(body),
  0, '', '', 'active',
  NOW(), NOW()
);

-- Portal: list outstanding charges for a patient
-- name: PortalOutstandingCharges :many
SELECT
  p.id,
  p.procdt AS date,
  p.proccharges AS total_charge,
  p.procbalcurrent AS balance,
  cpt.abbrev AS cpt_code,
  p.proccomment AS description,
  p.procstatus AS status
FROM procrec p
LEFT OUTER JOIN cpt cpt ON cpt.id = p.proccpt
WHERE p.procpatient = sqlc.arg(patient_id)
  AND p.procbalcurrent > 0
ORDER BY p.procdt DESC;

-- Portal: list unapplied credits for a patient
-- name: PortalUnappliedCredits :many
SELECT
  pr.id,
  pr.payrecdtadd AS date,
  pr.payrecamt AS amount,
  pr.payrectype AS payment_type,
  pr.payrecdescrip AS description,
  pr.payrecproc AS procedure_id
FROM payrec pr
WHERE pr.payrecpatient = sqlc.arg(patient_id)
  AND pr.active = 'active'
  AND pr.payrecproc = 0
ORDER BY pr.payrecdtadd DESC;

-- Portal: list superbills/statements for a patient
-- name: PortalListSuperbills :many
SELECT
  s.id,
  s.created_at,
  s.date_from,
  s.date_to,
  s.provider,
  s.status,
  s.total_charges,
  s.date_created
FROM superbill s
WHERE s.patient = sqlc.arg(patient_id)
ORDER BY s.created_at DESC
LIMIT 12;

-- Portal: cancel appointment (set status to cancelled + reason)
-- name: PortalCancelAppointment :exec
UPDATE scheduler
SET calstatus = 'cancelled',
    calprenote = CONCAT(COALESCE(calprenote, ''), ' [CANCELLED: ', sqlc.arg(cancel_reason), ']'),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND calpatient = sqlc.arg(patient_id)
  AND calstatus != 'cancelled';

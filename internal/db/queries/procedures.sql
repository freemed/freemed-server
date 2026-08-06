-- List procedures for a patient with CPT code name
-- name: PatientProcedures :many
SELECT
  pr.id,
  pr.procdt AS date_of_service,
  pr.proccpt AS cpt_id,
  c.cptnameext AS cpt_name,
  pr.proccharges AS charge,
  pr.procbalcurrent AS balance,
  pr.procamtpain AS amount_paid,
  pr.procstatus AS status,
  pr.procdiag1 AS diagnosis_id,
  pr.procpos AS place_of_service,
  pr.procphysician AS provider_id,
  pr.procunits AS units,
  pr.procdiagset AS diagnosis_set,
  pr.procbalorig AS balance_original,
  pr.procdtend AS date_of_service_end,
  pr.proccptmod AS cpt_modifier_1,
  pr.proccptmod2 AS cpt_modifier_2,
  pr.proccptmod3 AS cpt_modifier_3
FROM procrec pr
LEFT JOIN cpt c ON c.id = pr.proccpt
WHERE pr.procpatient = sqlc.arg(patient_id)
  AND pr.active = 'active'
ORDER BY pr.procdt DESC;

-- Get single procedure detail with joins
-- name: GetProcedure :one
SELECT
  pr.id,
  pr.created_at,
  pr.updated_at,
  pr.procpatient AS patient_id,
  pr.proceoc AS episode_of_care,
  pr.proccpt AS cpt_id,
  c.cptnameext AS cpt_name,
  c.abbrev AS cpt_code,
  pr.proccptmod AS cpt_modifier_1,
  pr.proccptmod2 AS cpt_modifier_2,
  pr.proccptmod3 AS cpt_modifier_3,
  pr.procdiag1 AS diagnosis_1_id,
  pr.procdiag2 AS diagnosis_2_id,
  pr.procdiag3 AS diagnosis_3_id,
  pr.procdiag4 AS diagnosis_4_id,
  pr.procdiagset AS diagnosis_set,
  pr.proccharges AS charge,
  pr.procunits AS units,
  pr.procvoucher AS voucher,
  pr.procphysician AS provider_id,
  CONCAT(ph.phylname, ', ', ph.phyfname) AS provider_name,
  pr.procdt AS date_of_service,
  pr.procdtend AS date_of_service_end,
  pr.procpos AS place_of_service,
  pr.proccomment AS comment,
  pr.procbalorig AS balance_original,
  pr.procbalcurrent AS balance_current,
  pr.procamtpain AS amount_paid,
  pr.procbilled AS billed,
  pr.procbillable AS billable,
  pr.procauth AS authorization_id,
  pr.procrefdoc AS referring_provider_id,
  pr.procrefdt AS referral_date,
  pr.procamtallowed AS amount_allowed,
  pr.procdtbilled AS billed_date,
  pr.proccurcovid AS current_coverage_id,
  pr.proccurcovtp AS current_coverage_type_id,
  pr.proccov1 AS coverage_id_1,
  pr.proccov2 AS coverage_id_2,
  pr.proccov3 AS coverage_id_3,
  pr.proccov4 AS coverage_id_4,
  pr.procmedicaidref AS medicaid_reference,
  pr.procmedicaidresub AS medicaid_resubmission,
  pr.proclabcharges AS lab_charges,
  pr.procstatus AS status,
  pr.procslidingscale AS sliding_scale,
  pr.proctosoverride AS type_of_service_override,
  pr.orderid AS order_id,
  pr.user AS user_id,
  pr.active
FROM procrec pr
LEFT JOIN cpt c ON c.id = pr.proccpt
LEFT JOIN physician ph ON ph.id = pr.procphysician
WHERE pr.id = sqlc.arg(id);

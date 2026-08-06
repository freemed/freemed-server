-- name: SchedulerDailyApptRange :many
SELECT s.caldateof AS date_of
, DATE_FORMAT(s.caldateof, '%m/%d/%Y') AS date_of_mdy
, s.calhour AS hour
, s.calminute AS minute
, CONCAT(LPAD(s.calhour, 2, '0'), ':', LPAD(s.calminute, 2, '0')) AS appointment_time
, s.calduration AS duration
, CONCAT(ph.phylname, ', ', ph.phyfname) AS provider
, ph.id AS provider_id
, s.caltype AS resource_type
, CASE s.caltype WHEN 'block' THEN '-' WHEN 'temp' THEN CONCAT('[!] ', ci.cilname, ', ', ci.cifname, ' (', ci.cicomplaint, ')') WHEN 'group' THEN CONCAT(cg.groupname, ' (', cg.grouplength, ' members)') ELSE CONCAT(pa.ptlname, ', ', pa.ptfname, IF(LENGTH(pa.ptmname)>0, CONCAT(' ', pa.ptmname), ''), IF(LENGTH(pa.ptsuffix)>0, CONCAT(' ', pa.ptsuffix), ''), IF(LENGTH(pa.ptid)>0, CONCAT(' (', pa.ptid, ')'), '')) END AS patient
, s.calpatient AS patient_id
, s.calprenote AS note
, SUBSTRING_INDEX(GROUP_CONCAT(st.sname), ',', -1) AS status
, SUBSTRING_INDEX(GROUP_CONCAT(st.scolor), ',', -1) AS status_color
, s.id AS scheduler_id
, s.calappttemplate AS appointment_template_id
, aptm.atcolor AS template_color
FROM scheduler s
LEFT OUTER JOIN appttemplate aptm ON s.calappttemplate = aptm.id
LEFT OUTER JOIN scheduler_status ss ON s.id = ss.csappt
LEFT OUTER JOIN schedulerstatustype st ON st.id = ss.csstatus
LEFT OUTER JOIN physician ph ON s.calphysician = ph.id
LEFT OUTER JOIN patient pa ON s.calpatient = pa.id
LEFT OUTER JOIN callin ci ON s.calpatient = ci.id
LEFT OUTER JOIN calgroup cg ON s.calpatient = cg.id
WHERE ( s.caldateof >= sqlc.arg(from_date) AND s.caldateof <= sqlc.arg(to_date) )
AND s.calstatus NOT IN ( 'noshow', 'cancelled' )
GROUP BY s.id, ss.csappt
ORDER BY s.caldateof, s.calhour, s.calminute, s.calphysician DESC;

-- name: SchedulerDailyApptRangeByProvider :many
SELECT s.caldateof AS date_of
, DATE_FORMAT(s.caldateof, '%m/%d/%Y') AS date_of_mdy
, s.calhour AS hour
, s.calminute AS minute
, CONCAT(LPAD(s.calhour, 2, '0'), ':', LPAD(s.calminute, 2, '0')) AS appointment_time
, s.calduration AS duration
, CONCAT(ph.phylname, ', ', ph.phyfname) AS provider
, ph.id AS provider_id
, s.caltype AS resource_type
, CASE s.caltype WHEN 'block' THEN '-' WHEN 'temp' THEN CONCAT('[!] ', ci.cilname, ', ', ci.cifname, ' (', ci.cicomplaint, ')') WHEN 'group' THEN CONCAT(cg.groupname, ' (', cg.grouplength, ' members)') ELSE CONCAT(pa.ptlname, ', ', pa.ptfname, IF(LENGTH(pa.ptmname)>0, CONCAT(' ', pa.ptmname), ''), IF(LENGTH(pa.ptsuffix)>0, CONCAT(' ', pa.ptsuffix), ''), IF(LENGTH(pa.ptid)>0, CONCAT(' (', pa.ptid, ')'), '')) END AS patient
, s.calpatient AS patient_id
, s.calprenote AS note
, SUBSTRING_INDEX(GROUP_CONCAT(st.sname), ',', -1) AS status
, SUBSTRING_INDEX(GROUP_CONCAT(st.scolor), ',', -1) AS status_color
, s.id AS scheduler_id
, s.calappttemplate AS appointment_template_id
, aptm.atcolor AS template_color
FROM scheduler s
LEFT OUTER JOIN appttemplate aptm ON s.calappttemplate = aptm.id
LEFT OUTER JOIN scheduler_status ss ON s.id = ss.csappt
LEFT OUTER JOIN schedulerstatustype st ON st.id = ss.csstatus
LEFT OUTER JOIN physician ph ON s.calphysician = ph.id
LEFT OUTER JOIN patient pa ON s.calpatient = pa.id
LEFT OUTER JOIN callin ci ON s.calpatient = ci.id
LEFT OUTER JOIN calgroup cg ON s.calpatient = cg.id
WHERE ( s.caldateof >= sqlc.arg(from_date) AND s.caldateof <= sqlc.arg(to_date) )
AND s.calstatus NOT IN ( 'noshow', 'cancelled' )
AND s.calphysician = sqlc.arg(provider_id)
GROUP BY s.id, ss.csappt
ORDER BY s.caldateof, s.calhour, s.calminute, s.calphysician DESC;

-- name: SchedulerDailyApptScheduler :many
CALL schedulerGenerateDailySchedule(sqlc.arg(req_date), sqlc.arg(start_hour), sqlc.arg(end_hour), sqlc.arg(interval_minutes), sqlc.arg(provider_id));

-- name: SchedulerFindDateAppt :many
SELECT * FROM scheduler
WHERE caldateof = sqlc.arg(req_date)
AND calstatus != 'cancelled';

-- name: SchedulerFindDateApptByProvider :many
SELECT * FROM scheduler
WHERE caldateof = sqlc.arg(req_date)
AND calstatus != 'cancelled'
AND calphysician = sqlc.arg(provider_id);

-- name: SchedulerGetEvent :one
SELECT s.caldateof AS date_of
, DATE_FORMAT(s.caldateof, '%m/%d/%Y') AS date_of_mdy
, s.calhour AS hour
, s.calminute AS minute
, CONCAT(LPAD(s.calhour, 2, '0'), ':', LPAD(s.calminute, 2, '0')) AS appointment_time
, s.calduration AS duration
, CONCAT(ph.phylname, ', ', ph.phyfname) AS provider
, ph.id AS provider_id
, s.caltype AS resource_type
, CASE s.caltype WHEN 'block' THEN '-' WHEN 'temp' THEN CONCAT('[!] ', ci.cilname, ', ', ci.cifname, ' (', ci.cicomplaint, ')') WHEN 'group' THEN CONCAT(cg.groupname, ' (', cg.grouplength, ' members)') ELSE CONCAT(pa.ptlname, ', ', pa.ptfname, IF(LENGTH(pa.ptmname)>0, CONCAT(' ', pa.ptmname), ''), IF(LENGTH(pa.ptsuffix)>0, CONCAT(' ', pa.ptsuffix), ''), IF(LENGTH(pa.ptid)>0, CONCAT(' (', pa.ptid, ')'), '')) END AS patient
, s.calpatient AS patient_id
, s.calprenote AS note
, SUBSTRING_INDEX(GROUP_CONCAT(st.sname), ',', -1) AS status
, SUBSTRING_INDEX(GROUP_CONCAT(st.scolor), ',', -1) AS status_color
, s.id AS scheduler_id
, s.calappttemplate AS appointment_template_id
, aptm.atcolor AS template_color
FROM scheduler s
LEFT OUTER JOIN appttemplate aptm ON s.calappttemplate = aptm.id
LEFT OUTER JOIN scheduler_status ss ON s.id = ss.csappt
LEFT OUTER JOIN schedulerstatustype st ON st.id = ss.csstatus
LEFT OUTER JOIN physician ph ON s.calphysician = ph.id
LEFT OUTER JOIN patient pa ON s.calpatient = pa.id
LEFT OUTER JOIN callin ci ON s.calpatient = ci.id
LEFT OUTER JOIN calgroup cg ON s.calpatient = cg.id
WHERE s.id = sqlc.arg(id);

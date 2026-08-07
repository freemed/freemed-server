-- name: GetSchedulerById :one
SELECT * FROM scheduler WHERE id = sqlc.arg(id);

-- name: UpdateScheduler :exec
UPDATE scheduler SET
  caldateof = sqlc.arg(date_of),
  calhour = sqlc.arg(hour),
  calminute = sqlc.arg(minute),
  calduration = sqlc.arg(duration),
  calmodified = sqlc.arg(modified)
WHERE id = sqlc.arg(id);

-- name: CreateAppointment :execresult
INSERT INTO scheduler (
  caldateof, calhour, calminute, calduration, caltype,
  calphysician, calpatient, calstatus, calprenote,
  calcreated, calmodified, calgroupid, calgroupmembers,
  calattendees, user,
  created_at, updated_at
) VALUES (
  sqlc.arg(caldateof), sqlc.arg(calhour), sqlc.arg(calminute), sqlc.arg(calduration),
  sqlc.arg(caltype), sqlc.arg(calphysician), sqlc.arg(calpatient),
  'scheduled', sqlc.arg(calprenote),
  NOW(), NULL, sqlc.narg(calgroupid), sqlc.narg(calgroupmembers),
  sqlc.narg(calattendees), sqlc.arg(user),
  NOW(), NOW()
);

-- name: CopyAppointment :execresult
INSERT INTO scheduler (
  caldateof, calhour, calminute, calduration, caltype,
  calphysician, calpatient, calstatus, calprenote, calpostnote,
  calcreated, calmodified, calfacility, calroom, calmark,
  calgroupid, calgroupmembers, calrecurnote, calrecurid,
  calappttemplate, calattendees, user,
  created_at, updated_at
)
SELECT
  sqlc.arg(new_caldateof), sqlc.arg(new_calhour), sqlc.arg(new_calminute),
  calduration, caltype,
  calphysician, calpatient, 'scheduled', calprenote, calpostnote,
  NOW(), NULL, calfacility, calroom, calmark,
  calgroupid, calgroupmembers, calrecurnote, calrecurid,
  calappttemplate, calattendees, user,
  NOW(), NOW()
FROM scheduler s
WHERE s.id = sqlc.arg(source_id);

-- name: CreateGroupAppointment :execresult
INSERT INTO scheduler (
  caldateof, calhour, calminute, calduration, caltype,
  calphysician, calpatient, calstatus, calprenote,
  calcreated, calmodified, calgroupid, calgroupmembers,
  calattendees, user,
  created_at, updated_at
) VALUES (
  sqlc.arg(caldateof), sqlc.arg(calhour), sqlc.arg(calminute), sqlc.arg(calduration),
  'group', sqlc.arg(calphysician), 0, 'scheduled', sqlc.arg(calprenote),
  NOW(), NULL, sqlc.arg(calgroupid), sqlc.arg(calgroupmembers),
  sqlc.narg(calattendees), sqlc.arg(user),
  NOW(), NOW()
);

-- name: FindGroupAppointments :many
SELECT s.*, cg.groupname
FROM scheduler s
LEFT JOIN calgroup cg ON s.calgroupid = cg.id
WHERE s.caltype = 'group' AND s.calgroupid = sqlc.arg(calgroupid);

-- name: CreateRecurringAppointment :exec
INSERT INTO scheduler (
  caldateof, calhour, calminute, calduration, caltype,
  calphysician, calpatient, calstatus, calprenote,
  calcreated, user,
  created_at, updated_at
)
SELECT
  sqlc.arg(caldateof), calhour, calminute, calduration, caltype,
  calphysician, calpatient, 'scheduled', calprenote,
  NOW(), sqlc.arg(user),
  NOW(), NOW()
FROM scheduler s
WHERE s.id = sqlc.arg(source_id);

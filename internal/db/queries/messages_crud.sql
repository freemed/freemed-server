-- name: GetMessageById :one
SELECT * FROM messages WHERE id = sqlc.arg(id);

-- name: CreateMessage :execresult
INSERT INTO messages (
  msgby, msgtime, msgfor, msgrecip, msgpatient, msgperson,
  msgurgency, msgsubject, msgtext, msgread, msgunique, msgtag, active
) VALUES (
  sqlc.arg(msgby), sqlc.arg(msgtime), sqlc.arg(msgfor), sqlc.arg(msgrecip),
  sqlc.arg(msgpatient), sqlc.arg(msgperson), sqlc.arg(msgurgency),
  sqlc.arg(msgsubject), sqlc.arg(msgtext), sqlc.arg(msgread),
  sqlc.arg(msgunique), sqlc.arg(msgtag), sqlc.arg(active)
);

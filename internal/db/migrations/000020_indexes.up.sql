-- Migration 000020: Add database indexes for common query patterns
-- Analyzed from internal/db/queries/*.sql WHERE/JOIN clauses

-- patient table: frequently filtered by ptarchive (almost all queries),
-- sorted/searched by ptlname and ptfname (picklists, search, duplicate check)
CREATE INDEX idx_patient_ptarchive_ptlname ON patient (ptarchive, ptlname);
CREATE INDEX idx_patient_ptfname ON patient (ptfname);
CREATE INDEX idx_patient_ptid ON patient (ptid);

-- scheduler table: date range queries (caldateof), provider filter (calphysician),
-- patient lookups (calpatient), status filtering (calstatus)
CREATE INDEX idx_scheduler_caldateof ON scheduler (caldateof);
CREATE INDEX idx_scheduler_calphysician ON scheduler (calphysician);
CREATE INDEX idx_scheduler_calpatient ON scheduler (calpatient);
CREATE INDEX idx_scheduler_calstatus ON scheduler (calstatus);

-- messages table: user inbox (msgfor), patient-scoped messages (msgpatient),
-- tag-based filtering (msgtag)
CREATE INDEX idx_messages_msgfor ON messages (msgfor);
CREATE INDEX idx_messages_msgpatient ON messages (msgpatient);
CREATE INDEX idx_messages_msgtag ON messages (msgtag);

-- procrec table: patient procedures listing (procpatient)
CREATE INDEX idx_procrec_procpatient ON procrec (procpatient);

-- user table: login lookups and duplicate username prevention
CREATE UNIQUE INDEX idx_user_username ON user (username);

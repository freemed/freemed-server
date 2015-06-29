package model

import (
	"database/sql"
	"time"
)

const (
	TABLE_PATIENT_EMR = "patient_emr"
)

type PatientEmrModel struct {
	Patient    int64          `db:"patient" json:"patient_id"`
	Module     string         `db:"module" json:"module"`
	RecordId   int64          `db:"oid" json:"oid"`
	Stamp      time.Time      `db:"stamp" json:"stamp"`
	Summary    string         `db:"summary" json:"summary"`
	Locked     bool           `db:"locked" json:"locked"`
	Annotation sql.NullString `db:"annotation" json:"annotation"`
	User       int64          `db:"user" json:"user_id"`
	Provider   int64          `db:"provider" json:"provider_id"`
	Language   string         `db:"language" json:"language"`
	Status     string         `db:"status" json:"status"`
	Id         int64          `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_PATIENT_EMR, Obj: PatientEmrModel{}, Key: "Id"})
}

package model

import (
	"github.com/freemed/freemed-server/db"
	"time"
)

const (
	TABLE_PATIENT_ID = "patient_ids"
)

type PatientIdModel struct {
	Patient   int64     `db:"patient" json:"patient_id"`
	ForeignId string    `db:"foreign_id" json:"foreign_identifier"`
	Facility  int64     `db:"facility" json:"facility_id"`
	Practice  int64     `db:"practice" json:"practice_id"`
	User      int64     `db:"user" json:"user_id"`
	Stamp     time.Time `db:"stamp" json:"stamp"`
	Active    string    `db:"active" json:"active"`
	Id        int64     `db:"id" json:"id"`
}

func init() {
	db.DbTables = append(db.DbTables, db.DbTable{TableName: TABLE_PATIENT_ID, Obj: PatientIdModel{}, Key: "Id"})
}

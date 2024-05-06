package model

import (
	"time"

	"github.com/freemed/freemed-server/common"
	"gorm.io/gorm"
)

const (
	TABLE_PATIENT_ID  = "patient_ids"
	MODULE_PATIENT_ID = "patient_ids"
)

type PatientIdModel struct {
	gorm.Model
	Patient   int64     `db:"patient" json:"patient_id"`
	ForeignID string    `db:"foreign_id" json:"foreign_identifier"`
	Facility  int64     `db:"facility" json:"facility_id"`
	Practice  int64     `db:"practice" json:"practice_id"`
	User      int64     `db:"user" json:"user_id"`
	Stamp     time.Time `db:"stamp" json:"stamp"`
	Active    string    `db:"active" json:"active"`
	Id        int64     `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{
		TableName: TABLE_PATIENT_ID,
		Obj:       PatientIdModel{},
		Key:       "Id",
	})
	common.EmrModuleMap[MODULE_PATIENT_ID] = common.EmrModuleType{
		Name:         MODULE_PATIENT_ID,
		PatientField: "Patient",
		Type:         PatientIdModel{},
	}
}

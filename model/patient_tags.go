package model

import (
	"time"

	"github.com/freemed/freemed-server/common"
	"gorm.io/gorm"
)

const (
	TABLE_PATIENT_TAGS  = "patienttag"
	MODULE_PATIENT_TAGS = "patienttag"
)

type PatientTagModel struct {
	gorm.Model
	Tag     string    `db:"tag" json:"tag"`
	Patient int64     `db:"patient" json:"patient_id"`
	User    int64     `db:"user" json:"user_id"`
	Stamp   time.Time `db:"datecreate" json:"stamp"`
	Expiry  NullTime  `db:"dateexpire" json:"expiry"`
	Id      int64     `db:"id" json:"id"`
}

func (PatientTagModel) TableName() string {
	return TABLE_PATIENT_TAGS
}

func init() {
	DbTables = append(DbTables, DbTable{
		TableName: TABLE_PATIENT_TAGS,
		Obj:       PatientTagModel{},
		Key:       "Id",
	})
	common.EmrModuleMap[MODULE_PATIENT_TAGS] = common.EmrModuleType{
		Name:         MODULE_PATIENT_TAGS,
		PatientField: "Patient",
		Type:         PatientTagModel{},
	}
}

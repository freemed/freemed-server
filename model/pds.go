package model

import (
	"github.com/freemed/freemed-server/common"
	"gorm.io/gorm"
)

const (
	TABLE_PDS  = "pds"
	MODULE_PDS = "pds"
)

type PatientDataStoreModel struct {
	gorm.Model
	Patient  int64  `db:"patient" json:"patient_id"`
	Module   string `db:"module" json:"module"`
	Contents []byte `db:"contents" json:"data"`
	Id       int64  `db:"id" json:"id"`
}

func (PatientDataStoreModel) TableName() string {
	return TABLE_PDS
}

func init() {
	DbTables = append(DbTables,
		DbTable{
			TableName: TABLE_PDS,
			Obj:       PatientDataStoreModel{},
			Key:       "Id",
		},
	)
	common.EmrModuleMap[MODULE_PDS] = common.EmrModuleType{
		Name:         MODULE_PDS,
		PatientField: "Patient",
		Type:         PatientDataStoreModel{},
	}
}

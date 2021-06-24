package model

import "github.com/freemed/freemed-server/common"

const (
	TABLE_PDS  = "pds"
	MODULE_PDS = "pds"
)

type PatientDataStoreModel struct {
	Patient  int64  `db:"patient" json:"patient_id"`
	Module   string `db:"module" json:"module"`
	Contents []byte `db:"contents" json:"data"`
	Id       int64  `db:"id" json:"id"`
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

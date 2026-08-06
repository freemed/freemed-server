package model

import (
	"database/sql"
	"time"
	"github.com/freemed/freemed-server/common"
)

const (
	TABLE_PDS  = "pds"
	MODULE_PDS = "pds"
)

type PatientDataStoreModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Patient  int64  `db:"patient" json:"patient_id"`
	Module   string `db:"module" json:"module"`
	Contents []byte `db:"contents" json:"data"`
}

func (PatientDataStoreModel) TableName() string {
	return TABLE_PDS
}

func init() {
	common.EmrModuleMap[MODULE_PDS] = common.EmrModuleType{
		Name:         MODULE_PDS,
		PatientField: "Patient",
		Type:         PatientDataStoreModel{},
	}
}

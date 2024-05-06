package model

import (
	"gorm.io/gorm"
)

const (
	TABLE_DATA_STORE = "pds"
)

type DataStoreModel struct {
	gorm.Model
	Patient  int64  `db:"patient" json:"patient_id"`
	Module   string `db:"module" json:"module"`
	Contents []byte `db:"contents" json:"contents"`
	Id       int64  `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{
		TableName: TABLE_DATA_STORE,
		Obj:       DataStoreModel{},
		Key:       "Id",
	})
}

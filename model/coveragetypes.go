package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	TABLE_COVERAGETYPES = "covtypes"
)

type CoverageTypeModel struct {
	gorm.Model
	Name        string    `db:"covtpname" json:"name"`
	Description string    `db:"covtpdescrip" json:"description"`
	Added       time.Time `db:"covtpdtadd" json:"added"`
	Modified    time.Time `db:"covtpdtmod" json:"modified"`
	Id          int64     `db:"id" json:"id"`
}

func (CoverageTypeModel) TableName() string {
	return TABLE_COVERAGETYPES
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_COVERAGETYPES, Obj: CoverageTypeModel{}, Key: "Id"})
}

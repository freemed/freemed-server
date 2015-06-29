package model

import (
	"github.com/freemed/freemed-server/db"
	"time"
)

const (
	TABLE_COVERAGETYPES = "covtypes"
)

type CoverageTypeModel struct {
	Name        string    `db:"covtpname" json:"name"`
	Description string    `db:"covtpdescrip" json:"description"`
	Added       time.Time `db:"covtpdtadd" json:"added"`
	Modified    time.Time `db:"covtpdtmod" json:"modified"`
	Id          int64     `db:"id" json:"id"`
}

func init() {
	db.DbTables = append(db.DbTables, db.DbTable{TableName: TABLE_COVERAGETYPES, Obj: CoverageTypeModel{}, Key: "Id"})
}

package model

import (
	"github.com/freemed/freemed-server/db"
)

const (
	TABLE_SCHEDULERSTATUSTYPE = "schedulerstatustype"
)

type SchedulerStatusTypeModel struct {
	Name        string `db:"sname" json:"name"`
	Description string `db:"sdescrip" json:"description"`
	Color       string `db:"scolor" json:"color"`
	Age         int    `db:"sage" json:"age"`
	Id          int64  `db:"id" json:"id"`
}

func init() {
	db.DbTables = append(db.DbTables, db.DbTable{TableName: TABLE_SCHEDULERSTATUSTYPE, Obj: SchedulerStatusTypeModel{}, Key: "Id"})
}

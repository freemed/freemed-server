package model

import (
	"time"
)

const (
	TABLE_SYSTEMNOTIFICATION = "systemnotification"
)

type SystemNotificationModel struct {
	Stamp   time.Time `db:"stamp" json:"stamp"`
	User    int64     `db:"nuser" json:"user"`
	Text    string    `db:"ntext" json:"text"`
	Action  string    `db:"naction" json:"action"`
	Module  string    `db:"nmodule" json:"module"`
	Patient int64     `db:"npatient" json:"patient"`
	Id      int64     `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_SYSTEMNOTIFICATION, Obj: SystemNotificationModel{}, Key: "Id"})
}

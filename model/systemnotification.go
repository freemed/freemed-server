package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	TABLE_SYSTEMNOTIFICATION = "systemnotification"
)

type SystemNotificationModel struct {
	gorm.Model
	Stamp   time.Time `db:"stamp" json:"stamp"`
	User    int64     `db:"nuser" json:"user"`
	Text    string    `db:"ntext" json:"text"`
	Action  string    `db:"naction" json:"action"`
	Module  string    `db:"nmodule" json:"module"`
	Patient int64     `db:"npatient" json:"patient"`
	Id      int64     `db:"id" json:"id"`
}

func (SystemNotificationModel) TableName() string {
	return TABLE_SYSTEMNOTIFICATION
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_SYSTEMNOTIFICATION, Obj: SystemNotificationModel{}, Key: "Id"})
}

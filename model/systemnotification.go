package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_SYSTEMNOTIFICATION = "systemnotification"
)

type SystemNotificationModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Stamp   time.Time `db:"stamp" json:"stamp"`
	User    int64     `db:"nuser" json:"user"`
	Text    string    `db:"ntext" json:"text"`
	Action  string    `db:"naction" json:"action"`
	Module  string    `db:"nmodule" json:"module"`
	Patient int64     `db:"npatient" json:"patient"`
}

func (SystemNotificationModel) TableName() string {
	return TABLE_SYSTEMNOTIFICATION
}

func init() {
}

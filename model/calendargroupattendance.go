package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_CALENDARGROUPATTENDANCE = "calgroupattend"
)

type CalendarGroupAttendanceModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Group         int64     `db:"calgroupid" json:"group_id"`
	SchedulerItem int64     `db:"calid" json:"scheduler_id"`
	Patient       int64     `db:"patient" json:"patient_id"`
	Status        string    `db:"calstatus" json:"status"`
	Stamp         time.Time `db:"stamp" json:"stamp"`
}

func (CalendarGroupAttendanceModel) TableName() string {
	return TABLE_CALENDARGROUPATTENDANCE
}

func init() {
}

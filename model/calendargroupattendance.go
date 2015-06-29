package model

import (
	"time"
)

const (
	TABLE_CALENDARGROUPATTENDANCE = "calgroupattend"
)

type CalendarGroupAttendanceModel struct {
	Group         int64     `db:"calgroupid" json:"group_id"`
	SchedulerItem int64     `db:"calid" json:"scheduler_id"`
	Patient       int64     `db:"patient" json:"patient_id"`
	Status        string    `db:"calstatus" json:"status"`
	Stamp         time.Time `db:"stamp" json:"stamp"`
	Id            int64     `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_CALENDARGROUPATTENDANCE, Obj: CalendarGroupAttendanceModel{}, Key: "Id"})
}

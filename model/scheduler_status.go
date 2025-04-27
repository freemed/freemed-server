package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	TABLE_SCHEDULER_STATUS = "scheduler_status"
)

type SchedulerStatusModel struct {
	gorm.Model
	Stamp       time.Time `db:"csstamp" json:"stamp"`
	Patient     int64     `db:"cspatient" json:"patient_id"`
	Appointment int64     `db:"csappt" json:"appointment_id"`
	Status      string    `db:"csstatus" json:"status"`
	Note        string    `db:"csenote" json:"note"`
	User        int64     `db:"user" json:"user"`
	Id          int64     `db:"id" json:"id"`
}

func (SchedulerStatusModel) TableName() string {
	return TABLE_SCHEDULER_STATUS
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_SCHEDULER_STATUS, Obj: SchedulerStatusModel{}, Key: "Id"})
}

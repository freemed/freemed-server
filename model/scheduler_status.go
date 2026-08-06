package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_SCHEDULER_STATUS = "scheduler_status"
)

type SchedulerStatusModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Stamp       time.Time `db:"csstamp" json:"stamp"`
	Patient     int64     `db:"cspatient" json:"patient_id"`
	Appointment int64     `db:"csappt" json:"appointment_id"`
	Status      string    `db:"csstatus" json:"status"`
	Note        string    `db:"csenote" json:"note"`
	User        int64     `db:"user" json:"user"`
}

func (SchedulerStatusModel) TableName() string {
	return TABLE_SCHEDULER_STATUS
}

func init() {
}

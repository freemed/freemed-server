package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_WORKFLOW_STATUS = "workflow_status"
)

type WorkflowStatusModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Stamp     time.Time `db:"stamp" json:"stamp"`
	Patient   int64     `db:"patient" json:"patient_id"`
	User      int64     `db:"user" json:"user"`
	Type      int64     `db:"status_type" json:"status_type"`
	Completed bool      `db:"status_completed" json:"status_completed"`
}

func (WorkflowStatusModel) TableName() string {
	return TABLE_WORKFLOW_STATUS
}

func init() {
}

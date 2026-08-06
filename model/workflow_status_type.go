package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_WORKFLOW_STATUS_TYPE = "workflow_status_type"
)

type WorkflowStatusTypeModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Name   string `db:"status_name" json:"status_name"`
	Order  string `db:"status_order" json:"status_order"`
	Module string `db:"status_module" json:"status_module"`
	Active bool   `db:"active" json:"active"`
}

func (WorkflowStatusTypeModel) TableName() string {
	return TABLE_WORKFLOW_STATUS
}

func init() {
}

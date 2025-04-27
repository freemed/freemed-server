package model

import "gorm.io/gorm"

const (
	TABLE_WORKFLOW_STATUS_TYPE = "workflow_status_type"
)

type WorkflowStatusTypeModel struct {
	gorm.Model
	Name   string `db:"status_name" json:"status_name"`
	Order  string `db:"status_order" json:"status_order"`
	Module string `db:"status_module" json:"status_module"`
	Active bool   `db:"active" json:"active"`
	Id     int64  `db:"id" json:"id"`
}

func (WorkflowStatusTypeModel) TableName() string {
	return TABLE_WORKFLOW_STATUS
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_WORKFLOW_STATUS_TYPE, Obj: WorkflowStatusTypeModel{}, Key: "Id"})
}

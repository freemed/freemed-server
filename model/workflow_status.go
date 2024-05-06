package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	TABLE_WORKFLOW_STATUS = "workflow_status"
)

type WorkflowStatusModel struct {
	gorm.Model
	Stamp     time.Time `db:"stamp" json:"stamp"`
	Patient   int64     `db:"patient" json:"patient_id"`
	User      int64     `db:"user" json:"user"`
	Type      int64     `db:"status_type" json:"status_type"`
	Completed bool      `db:"status_completed" json:"status_completed"`
	Id        int64     `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_WORKFLOW_STATUS, Obj: WorkflowStatusModel{}, Key: "Id"})
}

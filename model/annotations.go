package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_ANNOTATIONS = "annotations"
)

type AnnotationModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Stamp    time.Time `db:"atimestamp" json:"stamp"`
	Patient  int64     `db:"apatient" json:"patient"`
	Module   string    `db:"amodule" json:"module"`
	Table    string    `db:"atable" json:"table"`
	TargetId int64     `db:"aid" json:"target_id"`
	User     int64     `db:"auser" json:"user"`
	Text     string    `db:"annotation" json:"text"`
}

func (AnnotationModel) TableName() string {
	return TABLE_ANNOTATIONS
}

func init() {
}

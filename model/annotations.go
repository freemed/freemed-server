package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	TABLE_ANNOTATIONS = "annotations"
)

type AnnotationModel struct {
	gorm.Model
	Stamp    time.Time `db:"atimestamp" json:"stamp"`
	Patient  int64     `db:"apatient" json:"patient"`
	Module   string    `db:"amodule" json:"module"`
	Table    string    `db:"atable" json:"table"`
	TargetId int64     `db:"aid" json:"target_id"`
	User     int64     `db:"auser" json:"user"`
	Text     string    `db:"annotation" json:"text"`
	Id       int64     `db:"id" json:"id"`
}

func (AnnotationModel) TableName() string {
	return TABLE_ANNOTATIONS
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_ANNOTATIONS, Obj: AnnotationModel{}, Key: "Id"})
}

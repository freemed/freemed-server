package main

import (
	"time"
)

const (
	TABLE_ANNOTATIONS = "annotations"
)

type AnnotationModel struct {
	Stamp    time.Time `db:"atimestamp" json:"stamp"`
	Patient  int64     `db:"apatient" json:"patient"`
	Module   string    `db:"amodule" json:"module"`
	Table    string    `db:"atable" json:"table"`
	TargetId int64     `db:"aid" json:"target_id"`
	User     int64     `db:"auser" json:"user"`
	Text     string    `db:"annotation" json:"text"`
	Id       int64     `db:"id" json:"id"`
}

func init() {
	dbTables = append(dbTables, DbTable{TableName: TABLE_ANNOTATIONS, Obj: AnnotationModel{}, Key: "Id"})
}

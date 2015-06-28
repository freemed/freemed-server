package main

import (
	"database/sql"
)

const (
	TABLE_CONFIG = "config"
)

type ConfigModel struct {
	Key     string         `db:"c_option" json:"key"`
	Value   sql.NullString `db:"c_value" json:"value"`
	Title   sql.NullString `db:"c_title" json:"title"`
	Section sql.NullString `db:"c_section" json:"section"`
	Type    string         `db:"c_type" json:"type"`
	Options sql.NullString `db:"c_options" json:"options"`
	Id      int64          `db:"id" json:"id"`
}

func init() {
	dbTables = append(dbTables, DbTable{TableName: TABLE_CONFIG, Obj: ConfigModel{}, Key: "Id"})
}

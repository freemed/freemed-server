package model

import (
//"database/sql"
)

const (
	TABLE_CONFIG = "config"
)

type ConfigModel struct {
	Key     string     `db:"c_option" json:"key"`
	Value   NullString `db:"c_value" json:"value"`
	Title   NullString `db:"c_title" json:"title"`
	Section NullString `db:"c_section" json:"section"`
	Type    string     `db:"c_type" json:"type"`
	Options NullString `db:"c_options" json:"options"`
	Id      int64      `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_CONFIG, Obj: ConfigModel{}, Key: "Id"})
}

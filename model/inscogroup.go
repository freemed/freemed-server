package model

import ()

const (
	TABLE_INSCOGROUP = "inscogroup"
)

type InscoGroupModel struct {
	Name string `db:"inscogroup" json:"name"`
	Id   int64  `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_INSCOGROUP, Obj: InscoGroupModel{}, Key: "Id"})
}

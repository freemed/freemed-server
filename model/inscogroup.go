package model

import "gorm.io/gorm"

const (
	TABLE_INSCOGROUP = "inscogroup"
)

type InscoGroupModel struct {
	gorm.Model
	Name string `db:"inscogroup" json:"name"`
	Id   int64  `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_INSCOGROUP, Obj: InscoGroupModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "inscogroup", Query: "SELECT name AS v, id AS k FROM " + TABLE_INSCOGROUP + " WHERE name LIKE CONCAT('%', :query, '%') ORDER BY name"})
}

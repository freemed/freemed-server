package model

import "gorm.io/gorm"

const (
	TABLE_DRUGFORM = "drugforms"
)

type DrugFormModel struct {
	gorm.Model
	Code        string `db:"code" json:"code"`
	Description string `db:"description" json:"description"`
	Id          int64  `db:"id" json:"id"`
}

func (DrugFormModel) TableName() string {
	return TABLE_DRUGFORM
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_DRUGFORM, Obj: DrugFormModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "drugform", Query: "SELECT CONCAT(code, ' - ', description) AS v, id AS k FROM " + TABLE_DRUGFORM + " WHERE CONCAT(code, ' - ', description) LIKE CONCAT('%', :query, '%') ORDER BY code, description"})
}

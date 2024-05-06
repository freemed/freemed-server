package model

import "gorm.io/gorm"

const (
	TABLE_INSURANCEMODIFIER = "insmod"
)

type InsuranceModifierModel struct {
	gorm.Model
	Modifier    string `db:"insmod" json:"modifier"`
	Description string `db:"insmoddesc" json:"description"`
	Id          int64  `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_INSURANCEMODIFIER, Obj: InsuranceModifierModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "insurancemodifier", Query: "SELECT CONCAT(insmod, ' - ', insmoddesc) AS v, id AS k FROM " + TABLE_INSURANCEMODIFIER + " WHERE CONCAT(insmod, ' - ', insmoddesc) LIKE CONCAT('%', :query, '%') ORDER BY insmod,insmoddesc"})
}

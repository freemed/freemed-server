package model

import "gorm.io/gorm"

const (
	TABLE_CPTMODIFIER = "cptmod"
)

type CptModifierModel struct {
	gorm.Model
	Modifier    string `db:"cptmod" json:"modifier"`
	Description string `db:"cptmoddescrip" json:"description"`
	Id          int64  `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_CPTMODIFIER, Obj: CptModifierModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "cptmodifier", Query: "SELECT CONCAT(cptmod, ' ', cptmoddescrip') AS v, id AS k FROM " + TABLE_CPTMODIFIER + " WHERE CONCAT(cptmod, ' ', cptmoddescrip) LIKE CONCAT('%', :query, '%') ORDER BY cptmod, cptmoddescrip"})
}

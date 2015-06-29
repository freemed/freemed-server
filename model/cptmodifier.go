package model

import ()

const (
	TABLE_CPTMODIFIER = "cptmod"
)

type CptModifierModel struct {
	Modifier    string `db:"cptmod" json:"modifier"`
	Description string `db:"cptmoddescrip" json:"description"`
	Id          int64  `db:"id" json:"id"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_CPTMODIFIER, Obj: CptModifierModel{}, Key: "Id"})
}

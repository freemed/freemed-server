package model

import (
	"github.com/freemed/freemed-server/db"
)

const (
	TABLE_CPTMODIFIER = "cptmod"
)

type CptModifierModel struct {
	Modifier    string `db:"cptmod" json:"modifier"`
	Description string `db:"cptmoddescrip" json:"description"`
	Id          int64  `db:"id" json:"id"`
}

func init() {
	db.DbTables = append(db.DbTables, db.DbTable{TableName: TABLE_CPTMODIFIER, Obj: CptModifierModel{}, Key: "Id"})
}

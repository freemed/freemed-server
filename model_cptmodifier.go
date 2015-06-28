package main

const (
	TABLE_CPTMODIFIER = "cptmod"
)

type CptModifierModel struct {
	Modifier    string `db:"cptmod" json:"modifier"`
	Description string `db:"cptmoddescrip" json:"description"`
	Id          int64  `db:"id" json:"id"`
}

func init() {
	dbTables = append(dbTables, DbTable{TableName: TABLE_CPTMODIFIER, Obj: CptModifierModel{}, Key: "Id"})
}

package main

const (
	TABLE_BCCDC = "bccdc"
)

type BccdcModel struct {
	Code        string `db:"agent_code" json:"code"`
	Description string `db:"description" json:"description"`
	Id          int64  `db:"id" json:"id"`
}

func init() {
	dbTables = append(dbTables, DbTable{TableName: TABLE_BCCDC, Obj: BccdcModel{}, Key: "Id"})
}

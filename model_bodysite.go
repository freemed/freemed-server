package main

const (
	TABLE_BODYSITE = "bodysite"
)

type BodySiteModel struct {
	Abbreviation string `db:"abbrev" json:"abbrev"`
	Language     string `db:"display_value" json:"description"`
	Id           int64  `db:"id" json:"id"`
}

func init() {
	dbTables = append(dbTables, DbTable{TableName: TABLE_BODYSITE, Obj: BodySiteModel{}, Key: "Id"})
}

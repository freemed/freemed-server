package model

import ()

const (
	TABLE_I18NLANGUAGES = "i18nlanguages"
)

type I18nLanguageModel struct {
	Abbreviation string `db:"abbrev" json:"abbrev"`
	Language     string `db:"language" json:"language"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_I18NLANGUAGES, Obj: I18nLanguageModel{}})
}

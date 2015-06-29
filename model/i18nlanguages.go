package model

import (
	"github.com/freemed/freemed-server/db"
)

const (
	TABLE_I18NLANGUAGES = "i18nlanguages"
)

type I18nLanguageModel struct {
	Abbreviation string `db:"abbrev" json:"abbrev"`
	Language     string `db:"language" json:"language"`
}

func init() {
	db.DbTables = append(db.DbTables, db.DbTable{TableName: TABLE_I18NLANGUAGES, Obj: I18nLanguageModel{}})
}

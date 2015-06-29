package model

import (
	"github.com/freemed/freemed-server/db"
)

const (
	TABLE_BODYSITE = "bodysite"
)

type BodySiteModel struct {
	Abbreviation string `db:"abbrev" json:"abbrev"`
	Language     string `db:"display_value" json:"description"`
	Id           int64  `db:"id" json:"id"`
}

func init() {
	db.DbTables = append(db.DbTables, db.DbTable{TableName: TABLE_BODYSITE, Obj: BodySiteModel{}, Key: "Id"})
}

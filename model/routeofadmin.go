package model

import (
	"github.com/freemed/freemed-server/db"
)

const (
	TABLE_ROUTEOFADMIN = "bodysite"
)

type RouteOfAdministrationModel struct {
	Abbreviation string `db:"abbrev" json:"abbrev"`
	DisplayValue string `db:"display_value" json:"description"`
	Id           int64  `db:"id" json:"id"`
}

func init() {
	db.DbTables = append(db.DbTables, db.DbTable{TableName: TABLE_ROUTEOFADMIN, Obj: RouteOfAdministrationModel{}, Key: "Id"})
}

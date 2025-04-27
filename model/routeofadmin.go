package model

import "gorm.io/gorm"

const (
	TABLE_ROUTEOFADMIN = "bodysite"
)

type RouteOfAdministrationModel struct {
	gorm.Model
	Abbreviation string `db:"abbrev" json:"abbrev"`
	DisplayValue string `db:"display_value" json:"description"`
	Id           int64  `db:"id" json:"id"`
}

func (RouteOfAdministrationModel) TableName() string {
	return TABLE_ROUTEOFADMIN
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_ROUTEOFADMIN, Obj: RouteOfAdministrationModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "inscogroup", Query: "SELECT CONCAT(abbrev, ' ', description) AS v, id AS k FROM " + TABLE_ROUTEOFADMIN + " WHERE abbrev LIKE CONCAT('%', :query, '%') OR description LIKE CONCAT('%', :query, '%') ORDER BY abbrev, description"})
}

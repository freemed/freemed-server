package model

import "gorm.io/gorm"

const (
	TABLE_BODYSITE = "bodysite"
)

type BodySiteModel struct {
	gorm.Model
	Abbreviation string `db:"abbrev" json:"abbrev"`
	Language     string `db:"display_value" json:"description"`
	Id           int64  `db:"id" json:"id"`
}

func (BodySiteModel) TableName() string {
	return TABLE_BODYSITE
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_BODYSITE, Obj: BodySiteModel{}, Key: "Id"})
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "bodysite", Query: "SELECT CONCAT(display_value, ' (', abbrev, ')') AS v, id AS k FROM " + TABLE_BODYSITE + " WHERE CONCAT(display_value, ' (', abbrev, ')') LIKE CONCAT('%', :query, '%') ORDER BY display_value, abbrev"})
}

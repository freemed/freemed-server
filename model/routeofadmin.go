package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_ROUTEOFADMIN = "bodysite"
)

type RouteOfAdministrationModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Abbreviation string `db:"abbrev" json:"abbrev"`
	DisplayValue string `db:"display_value" json:"description"`
}

func (RouteOfAdministrationModel) TableName() string {
	return TABLE_ROUTEOFADMIN
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "routeofadmin", Query: "SELECT CONCAT(abbrev, ' ', description) AS v, id AS k FROM " + TABLE_ROUTEOFADMIN + " WHERE abbrev LIKE CONCAT('%', :query, '%') OR description LIKE CONCAT('%', :query, '%') ORDER BY abbrev, description"})
}

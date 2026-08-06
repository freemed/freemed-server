package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_BODYSITE = "bodysite"
)

type BodySiteModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Abbreviation string `db:"abbrev" json:"abbrev"`
	Language     string `db:"display_value" json:"description"`
}

func (BodySiteModel) TableName() string {
	return TABLE_BODYSITE
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "bodysite", Query: "SELECT CONCAT(display_value, ' (', abbrev, ')') AS v, id AS k FROM " + TABLE_BODYSITE + " WHERE CONCAT(display_value, ' (', abbrev, ')') LIKE CONCAT('%', :query, '%') ORDER BY display_value, abbrev"})
}

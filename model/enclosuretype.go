package model

import (
	"database/sql"
	"time"

)

const (
	TABLE_ENCLOSURETYPE = "enctype"
)

type EnclosureTypeModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	EnclosureType string `db:"enclosure" json:"enclosure"`
}

func (EnclosureTypeModel) TableName() string {
	return TABLE_ENCLOSURETYPE
}

func init() {
	DbSupportPicklists = append(DbSupportPicklists, DbSupportPicklist{ModuleName: "enclosuretype", Query: "SELECT enclosure AS v, id AS k FROM " + TABLE_ENCLOSURETYPE + " WHERE enclosure LIKE CONCAT('%', :query, '%') ORDER BY enclosure"})
}
